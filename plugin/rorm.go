package plugin

import (
	"fmt"
	"strings"

	proto "github.com/gogo/protobuf/proto"
	descriptor "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/golangper/protoc-gen-rorm/options"
)

type RormPlugin struct {
	*generator.Generator
	imports   map[generator.GoPackageName]generator.GoImportPath
	redisType int64
	sqlType   int64
	useUid    bool
	useNsq    bool
	msgMap    map[string]*generator.Descriptor
}

// Name identifies the plugin
func (p *RormPlugin) Name() string {
	return "rorm"
}

func (p *RormPlugin) Init(g *generator.Generator) {
	p.Generator = g
	p.imports = make(map[generator.GoPackageName]generator.GoImportPath)
}

func (p *RormPlugin) Generate(file *generator.FileDescriptor) {
	p.imports["error"] = "error"
	p.msgMap = make(map[string]*generator.Descriptor, len(file.Messages()))
	for _, msg := range file.Messages() {
		msgName := generator.CamelCase(msg.GetName())
		p.msgMap[msgName] = msg
	}

	for _, svc := range file.GetService() {
		value, err := proto.GetExtension(svc.Options, options.E_RedisType)
		if err != nil || value == nil {
			//fmt.Println("===",err)
			p.redisType = 0
		} else {
			p.redisType = *(value.(*int64))
		}

		value, err = proto.GetExtension(svc.Options, options.E_SqlType)
		if err != nil || value == nil {
			p.sqlType = 0
		} else {
			p.sqlType = *(value.(*int64))
		}

		if p.sqlType == 2 {
			p.imports["sqlt"] = "github.com/albertwidi/sqlt"
		} else if p.sqlType == 1 {
			p.imports["sqlx"] = "github.com/jmoiron/sqlx"
		} else {
			fmt.Println("sqlType not set")
			return
		}
		if p.redisType > 0 {
			p.imports["redis"] = "github.com/go-redis/redis"
			//p.imports["strconv"] = "strconv"
		}
		p.useUid = proto.GetBoolExtension(svc.Options, options.E_UseUid, false)
		// if err == nil {
		// 	p.useUid = *(value.(*bool))
		// }else {
		// 	p.useUid = false
		// }

		if p.useUid {
			p.imports["snowflake"] = "github.com/fainted/snowflake"
		}

		p.useNsq = proto.GetBoolExtension(svc.Options, options.E_UseNsq, false)

		if p.useNsq {
			p.imports["nsqpool"] = "github.com/qgymje/nsqpool"
		}

		gin := proto.GetBoolExtension(svc.Options, options.E_GinHandler, false)

		if gin {
			p.imports["http"] = "net/http"
			p.imports["gin"] = "github.com/gin-gonic/gin"
			p.imports["binding"] = "github.com/gin-gonic/gin/binding"

		}
		p.imports["logging"] = "github.com/op/go-logging"
		p.imports["context"] = "golang.org/x/net/context"
		grpcSvcName := generator.CamelCase(svc.GetName()) + "GrpcImp"
		impName := generator.CamelCase(svc.GetName()) + "Imp"
		//grpc impl struct
		p.P(`type `, grpcSvcName, ` struct {`)
		p.In()
		if p.sqlType == 1 {
			p.P(`Db *sqlx.Db`)
		} else if p.sqlType == 2 {
			p.P(`Db *sqlt.Db`)
		}
		if p.redisType == 1 {
			p.P(`Redis *redis.Client`)
		} else if p.redisType == 2 {
			p.P(`Redis *redis.ClusterClient`)
		}
		if p.useNsq {
			p.P(`Nsq *pool.Pool`)
		}
		p.P(`Log *logging.Logger`)
		p.Out()
		p.P(`}`)

		//grpc method impl
		for _, method := range svc.GetMethod() {
			mname := generator.CamelCase(method.GetName())

			inputType := generator.CamelCase(strings.Split(method.GetInputType(), ".")[2])
			outputType := generator.CamelCase(strings.Split(method.GetOutputType(), ".")[2])

			uid := GetUidExtension(method.Options)
			opts := GetOptsExtension(method.Options)
			inputMsg, ok := p.msgMap[inputType]
			if !ok {
				fmt.Println("inputType: ", inputType, "not find")
				return
			}

			if uid != nil {
				err := CheckUidSeed(inputMsg, uid)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
			}

			p.P(`func (s *`, grpcSvcName, `) `, mname, `(c context.Context, in *`, inputType, `) (*`, outputType, `, error) {`)
			p.In()
			p.P(`var err error`)
			p.outAndValid(inputType, outputType)
			if uid != nil {
				p.newUid(inputMsg, uid)
			}
			if opts != nil {
				err := p.dealMethods(opts, inputType, outputType)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			p.P(`return out, nil`)
			p.Out()
			p.P(`}`)
		}

		p.P(`type `, impName, ` struct {`)
		p.In()
		p.P(grpcSvcName)
		p.Out()
		p.P(`}`)

		for _, m := range svc.GetMethod() {
			mname := generator.CamelCase(m.GetName())
			inputType := generator.CamelCase(strings.Split(m.GetInputType(), ".")[2])
			p.P(`func (s *`, impName, `) `, mname, `Handler(c *gin.Context) {`)
			p.In()

			p.P(`var prm *`, inputType)
			p.P(`var err error`)

			p.P(`err = c.ShouldBindWith(prm, binding.JSON)`)
			p.P(`if err != nil {`)
			p.In()
			p.P(`s.Log.Error(err.Error())`)
			p.P(`c.JSON(http.StatusBadRequest,gin.H{"resp": err.Error()})`)
			p.P(`return`)
			p.Out()
			p.P(`}`)

			p.P(`if err = prm.Validate(); err != nil {`)
			p.In()
			p.P(`s.Log.Error(err.Error())`)
			p.P(`c.JSON(http.StatusBadRequest,gin.H{"resp": err.Error()})`)
			p.P(`return`)
			p.Out()
			p.P(`}`)

			p.P(`res, err := s.`, mname, `(context.Background(), prm)`)
			p.P(`if err != nil {`)
			p.In()
			p.P(`s.Log.Error(err.Error())`)
			p.P(`c.JSON(http.StatusServiceUnavailable,gin.H{"resp": err.Error()})`)
			p.P(`return`)
			p.Out()
			p.P(`}`)
			p.P(`c.JSON(http.StatusOK,gin.H{"resp": res})`)
			p.Out()
			p.P(`}`)
		}
		// p.GenerateImports(file)
	}
}

func (p *RormPlugin) outAndValid(in, out string) {
	if in == out {
		p.P(`out := in`)
	} else {
		p.P(`out := &`, out, `{}`)
	}
	p.P(`err = in.Validate()`)
	p.P(`if err != nil {`)
	p.In()
	p.P(`return out, err`)
	p.Out()
	p.P(`}`)
}

func (p *RormPlugin) newUid(msg *generator.Descriptor, uid *options.UidOptions) {
	for _, field := range msg.Field {
		if field.GetName() == uid.Seed {
			p.P(`_s := in.`, generator.CamelCase(field.GetName()), ` % 256`)
			p.P(`_worker, err := snowflake.NewChannelWorker(s)`)
			p.P(`if err != nil {`)
			p.In()
			p.P("return out, err")
			p.Out()
			p.P(`}`)
			p.P(`_`, uid.Name, ` , _ := _worker.Next()`)
			p.P(`var _ = `, `_`, uid.Name)
			break
		}
	}
}

func (p *RormPlugin) dealMethods(opts *options.RormOptions, in, out string) error {
	err := p.dealMethod(opts, false, false, in, out)
	if err != nil {
		return err
	}

	return nil
}

func (p *RormPlugin) dealMethod(opt *options.RormOptions, end bool, els bool, in, out string) error {
	var err error
	if opt.GetParam() == "" {
		return fmt.Errorf("sqlx.MustExec param can not bu null")
	}
	param := strings.Replace(opt.GetParam(), "\n", "", -1)

	str := strings.Replace(param, `'`, `"`, -1)

	strArry := strings.Split(str, ";")
	str1 := strings.TrimSpace(strArry[0])
	str2 := " "
	for _, s := range strArry[1:] {
		str2 += ","
		str2 += CamelField(strings.TrimSpace(s))
	}
	str = strings.Replace(str, `;`, `,`, -1)
	//var tp descriptor.FieldDescriptorProto_Type
	tp, lb, err := p.getVarType(opt.GetTarget(), in, out)
	if err != nil {
		return err
	}
	switch opt.GetMethod() {
	case "sqlx.Exec":
		if lb == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("sqlx.Exec's target can not be repeated ")
		}
		p.P(`_, err = s.Db.Exec(`, str1, str2, `)`)
		p.dealErrBool(opt, tp)
	case "sqlx.Get":
		if lb == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("sqlx.GetMessage's target can not be repeated ")
		}
		if tp == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			p.P(`err = s.Db.Get(`, CamelField(opt.GetTarget()), ` , `, str1, str2, `)`)
		} else {
			p.P(`err = s.Db.Get( &`, CamelField(opt.GetTarget()), ` , `, str1, str2, `)`)
		}
		if opt.Failure == nil {
			p.dealErrReturn()
		}
	case "sqlx.Select":
		if lb != descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("sqlx.Select's target must be repeated ")
		}
		p.P(`err = s.Db.Select( &`, CamelField(opt.GetTarget()), ` , `, str1, str2, `)`)
		if opt.Failure == nil {
			p.dealErrReturn()
		}
	case "sqlx.PGet":
		if lb == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("sqlx.PGetMessage's target can not be repeated ")
		}
		t := strings.Split(CamelField(opt.GetTarget()), ".")
		n := t[len(t)-1]
		p.P(`stmt`, n, ` err := s.Db.Preparex(`, str1, `)`)
		p.dealErrReturn()
		if tp == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			p.P(`err = stmt`, n, `.Get(`, CamelField(opt.GetTarget()), str2, `)`)
		} else {
			p.P(`err = stmt`, n, `.Get(&`, CamelField(opt.GetTarget()), str2, `)`)
		}
		if opt.Failure == nil {
			p.dealErrReturn()
		}
	case "sqlx.PSelect":
		if lb != descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("sqlx.PSelect's target must be repeated ")
		}
		t := strings.Split(CamelField(opt.GetTarget()), ".")
		n := t[len(t)-1]
		p.P(`stmt`, n, ` err := s.Db.Preparex(`, str1, `)`)
		p.dealErrReturn()
		p.P(`err = stmt`, n, `.Select( &`, CamelField(opt.GetTarget()), ` , `, str2, `)`)
		if opt.Failure == nil {
			p.dealErrReturn()
		}
	case "sqlx.NExec":
		if lb == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("sqlx.NExec's target can not be repeated ")
		}
		p.P(`_, err = s.Db.NamedExec(`, str1, str2, `)`)
		p.dealErrBool(opt, tp)
	case "sqlx.PNGet":
		if lb == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("sqlx.PNGetMessage's target can not be repeated ")
		}
		t := strings.Split(CamelField(opt.GetTarget()), ".")
		n := t[len(t)-1]
		p.P(`stmt`, n, ` err := s.Db.PrepareNamed(`, str1, `)`)
		p.dealErrReturn()
		if tp == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			p.P(`err = stmt`, n, `.Get(`, CamelField(opt.GetTarget()), str2, `)`)
		} else {
			p.P(`err = stmt`, n, `.Get(&`, CamelField(opt.GetTarget()), str2, `)`)
		}
		if opt.Failure == nil {
			p.dealErrReturn()
		}
	case "sqlx.PNSelect":
		if lb != descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("sqlx.PNSelect's target must be repeated ")
		}
		t := strings.Split(CamelField(opt.GetTarget()), ".")
		n := t[len(t)-1]
		p.P(`stmt`, n, ` err := s.Db.PrepareNamed(`, str1, `)`)
		p.dealErrReturn()
		p.P(`err = stmt`, n, `.Select( &`, CamelField(opt.GetTarget()), ` , `, str2, `)`)
		if opt.Failure == nil {
			p.dealErrReturn()
		}

	case "redis.Get":
		key, err := p.getString(str1, in, out)
		if err != nil {
			return err
		}
		t := strings.Split(CamelField(opt.GetTarget()), ".")
		n := t[len(t)-1]
		switch tp {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			p.P(`rds`, n, `, err := s.Redis.Get(`, key, `).Bytes()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}

			p.P(`err = `, CamelField(opt.GetTarget()), `.Unmarshal(rds`, n, `)`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			p.P(CamelField(opt.GetTarget()), `, err := s.Redis.Get(`, key, `).String()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_INT64:
			p.P(CamelField(opt.Target), `, err := s.Redis.Get(`, key, `).Int64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
			p.P(`rds`, n, `, err := s.Redis.Get(`, key, `).Int64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
				p.P(CamelField(opt.Target), ` = uint64(`, `rds`, n, `)`)
			} else {
				p.P(`if err == nil {`)
				p.In()
				p.P(CamelField(opt.Target), ` = uint64(`, `rds`, n, `)`)
				p.Out()
				p.P(`}`)
			}
		case descriptor.FieldDescriptorProto_TYPE_INT32:
			p.P(`rds`, n, `, err := s.Redis.Get(`, key, `).Int64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
				p.P(CamelField(opt.Target), ` = int32(`, `rds`, n, `)`)
			} else {
				p.P(`if err == nil {`)
				p.In()
				p.P(CamelField(opt.Target), ` = int32(`, `rds`, n, `)`)
				p.Out()
				p.P(`}`)
			}
		case descriptor.FieldDescriptorProto_TYPE_FIXED32:
			p.P(`rds`, n, `, err := s.Redis.Get(`, key, `).Int64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
				p.P(CamelField(opt.Target), ` = uint32(`, `rds`, n, `)`)
			} else {
				p.P(`if err == nil {`)
				p.In()
				p.P(CamelField(opt.Target), ` = uint32(`, `rds`, n, `)`)
				p.Out()
				p.P(`}`)
			}
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
			p.P(CamelField(opt.Target), `, err := s.Redis.Get(`, key, `).Float64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			p.P(`rds`, n, `, err := s.Redis.Get(`, key, `).Float64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
				p.P(CamelField(opt.Target), ` = float32(`, `rds`, n, `)`)
			} else {
				p.P(`if err == nil {`)
				p.In()
				p.P(CamelField(opt.Target), ` = float32(`, `rds`, n, `)`)
				p.Out()
				p.P(`}`)
			}
		default:
			return fmt.Errorf("redis.Get's target type can not support ")
		}

	case "redis.Set":
		if len(strArry) < 3 {
			return fmt.Errorf("redis.Set's param must have 2 ")
		}
		key, err := p.getString(str1, in, out)
		if err != nil {
			return err
		}
		tp1, _, err := p.getVarType(strArry[1], in, out)
		if err != nil {
			return err
		}
		if !checkStrIsNum(strings.TrimSpace(strArry[2])) {
			return fmt.Errorf("redis.Set's the thired param must be int ")
		}
		t := strings.Split(CamelField(opt.GetTarget()), ".")
		n := t[len(t)-1]
		p.imports["time"] = "time"
		if tp1 == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			p.P(`set`, n, `, err := `, CamelField(strArry[1]), `.Marshal()`)
			p.dealErrReturn()
			p.P(`err = s.Redis.Set(`, key, `,`, `set`, n, `,int64(time.Duration(`, strArry[2], `) * time.Second)).Err()`)
		} else if tp1 == descriptor.FieldDescriptorProto_TYPE_BOOL {
			return fmt.Errorf("redis.Set's target can not be bool ")
		} else {
			p.P(`err = s.Redis.Set(`, key, `,`, CamelField(strArry[1]), `,int64(time.Duration(`, strArry[2], `) * time.Second)).Err()`)
		}
		p.dealErrBool(opt, tp)
	case "redis.Del":
		param := ""
		for _, st := range strArry {
			key, err := p.getString(strings.TrimSpace(st), in, out)
			if err != nil {
				return err
			}
			if param != "" {
				param += ","
			}
			param += key
		}
		p.P(`err = s.Redis.Del(`, param, `).Err()`)
		p.dealErrBool(opt, tp)
	case "redis.IncrByX":
		if len(strArry) < 2 {
			return fmt.Errorf("redis.IncrByX's param must have 2 ")
		}
		key, err := p.getString(str1, in, out)
		if err != nil {
			return err
		}
		tp1, _, err := p.getVarType(opt.GetTarget(), in, out)
		if err != nil {
			return err
		}
		switch tp1 {
		case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_INT32:
			p.P(`err = s.Redis.IncrBy(`, key, `, int64(`, CamelField(strArry[1]), `)).Err()`)
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE, descriptor.FieldDescriptorProto_TYPE_FLOAT:
			p.P(`err = s.Redis.IncrByFloat(`, key, `, float64(`, CamelField(strArry[1]), `)).Err()`)
		default:
			return fmt.Errorf("redis.IncrByX's The second param can be int32  int64 float32 float64")
		}
		p.dealErrBool(opt, tp)
	case "redis.DecrBy":
		if len(strArry) < 2 {
			return fmt.Errorf("redis.DecrBy's param must have 2 ")
		}
		key, err := p.getString(str1, in, out)
		if err != nil {
			return err
		}
		tp1, _, err := p.getVarType(opt.GetTarget(), in, out)
		if err != nil {
			return err
		}
		switch tp1 {
		case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_INT32:
			p.P(`_dnum,  err := s.Redis.Get(`, key, `).Int64()`)
			p.dealErrReturn()
			p.P(`if int(_dnum) < int(`, CamelField(strArry[1]), `){`)
			p.In()
			p.P(`return out, fmt.Errorf("Inventory shortage")`)
			p.Out()
			p.P(`}`)
			p.P(`err = s.Redis.IncrBy(`, key, `, int64(`, CamelField(strArry[1]), `)).Err()`)
		default:
			return fmt.Errorf("redis.IncrByX's The second param can be int32  int64")
		}
		p.dealErrBool(opt, tp)
	case "redis.Expire":
		if len(strArry) < 2 {
			return fmt.Errorf("redis.Expire's param must have 2 ")
		}
		if !checkStrIsInt(strings.TrimSpace(strArry[1])) {
			return fmt.Errorf("redis.Expire: The second param must be int num")
		}
		key, err := p.getString(str1, in, out)
		if err != nil {
			return err
		}
		p.imports["time"] = "time"
		p.P(`err = s.Redis.Expire(`, key, `, int64(time.Duration(`, strArry[1], `) * time.Second)).Err()`)
		p.dealErrBool(opt, tp)
	case "redis.HGet":
		if len(strArry) < 2 {
			return fmt.Errorf("redis.HGet's param must have 2 ")
		}
		key, err := p.getString(str1, in, out)
		if err != nil {
			return err
		}

		field, err := p.getString(strArry[1], in, out)
		if err != nil {
			return err
		}
		t := strings.Split(CamelField(opt.GetTarget()), ".")
		n := t[len(t)-1]
		switch tp {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			p.P(`rds`, n, `, err := s.Redis.HGet(`, key, `,`, field, `).Bytes()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}

			p.P(`err = `, CamelField(opt.GetTarget()), `.Unmarshal(rds`, n, `)`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			p.P(CamelField(opt.GetTarget()), `, err := s.Redis.HGet(`, key, `,`, field, `).String()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_INT64:
			p.P(CamelField(opt.GetTarget()), `, err := s.Redis.HGet(`, key, `,`, field, `).Int64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
			p.P(`rds`, n, `, err := s.Redis.HGet(`, key, `,`, field, `).Int64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
				p.P(CamelField(opt.Target), ` = uint64(`, `rds`, n, `)`)
			} else {
				p.P(`if err == nil {`)
				p.In()
				p.P(CamelField(opt.Target), ` = uint64(`, `rds`, n, `)`)
				p.Out()
				p.P(`}`)
			}
		case descriptor.FieldDescriptorProto_TYPE_INT32:
			p.P(`rds`, n, `, err := s.Redis.HGet(`, key, `,`, field, `).Int64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
				p.P(CamelField(opt.Target), ` = int32(`, `rds`, n, `)`)
			} else {
				p.P(`if err == nil {`)
				p.In()
				p.P(CamelField(opt.Target), ` = int32(`, `rds`, n, `)`)
				p.Out()
				p.P(`}`)
			}
		case descriptor.FieldDescriptorProto_TYPE_FIXED32:
			p.P(`rds`, n, `, err := s.Redis.HGet(`, key, `,`, field, `).Int64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
				p.P(CamelField(opt.Target), ` = uint32(`, `rds`, n, `)`)
			} else {
				p.P(`if err == nil {`)
				p.In()
				p.P(CamelField(opt.Target), ` = uint32(`, `rds`, n, `)`)
				p.Out()
				p.P(`}`)
			}
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
			p.P(CamelField(opt.GetTarget()), `, err := s.Redis.HGet(`, key, `,`, field, `).Float64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			p.P(`rds`, n, `, err := s.Redis.HGet(`, key, `,`, field, `).Float64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
				p.P(CamelField(opt.Target), ` = float32(`, `rds`, n, `)`)
			} else {
				p.P(`if err == nil {`)
				p.In()
				p.P(CamelField(opt.Target), ` = float32(`, `rds`, n, `)`)
				p.Out()
				p.P(`}`)
			}
		default:
			return fmt.Errorf("redis.Get's target can not be repeated ")
		}
	case "redis.HSet":
		if len(strArry) < 3 {
			return fmt.Errorf("redis.HSet's param must have 3 ")
		}
		key, err := p.getString(str1, in, out)
		if err != nil {
			return err
		}
		_, lb2, err := p.getVarType(strArry[1], in, out)
		if err != nil {
			return err
		}
		if lb2 == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("redis.HSet field param can not be repeated ")
		}
		field, err := p.getString(strArry[1], in, out)
		if err != nil {
			return err
		}

		tp1, lb1, err := p.getVarType(strArry[2], in, out)
		if err != nil {
			return err
		}
		if lb1 == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("redis.HSet value param can not be repeated ")
		}

		t := strings.Split(CamelField(strArry[2]), ".")
		n := t[len(t)-1]
		if tp1 == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			p.P(`set`, n, `, err := `, CamelField(strArry[2]), `.Marshal()`)
			p.dealErrReturn()
			p.P(`err = s.Redis.HSet(`, key, `,`, field, `, set`, n, `).Err()`)
		} else if tp1 == descriptor.FieldDescriptorProto_TYPE_BOOL {
			return fmt.Errorf("redis.Set's target can not be bool ")
		} else {
			p.P(`err = s.Redis.HSet(`, key, `,`, field, `,`, CamelField(strArry[2]), `).Err()`)
		}
		p.dealErrBool(opt, tp)
	case "redis.HDel":
		if len(strArry) < 2 {
			return fmt.Errorf("redis.HDel's param must have 2 ")
		}
		key, err := p.getString(str1, in, out)
		if err != nil {
			return err
		}
		field, err := p.getString(strArry[1], in, out)
		p.P(`err := s.Redis.HDel(`, key, `,`, field, `).Err()`)
		p.dealErrBool(opt, tp)
	case "redis.HincrByX":
		if len(strArry) < 3 {
			return fmt.Errorf("redis.HincrByX's param must have 3 ")
		}
		key, err := p.getString(str1, in, out)
		if err != nil {
			return err
		}
		field, err := p.getString(strArry[1], in, out)
		if err != nil {
			return err
		}
		tp2, _, err := p.getVarType(strArry[2], in, out)
		if err != nil {
			return err
		}
		switch tp2 {
		case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_INT32:
			p.P(`err = s.Redis.HIncrBy(`, key, `,`, field, `, int64(`, CamelField(strArry[2]), `)).Err()`)
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE, descriptor.FieldDescriptorProto_TYPE_FLOAT:
			p.P(`err = s.Redis.HIncrByFloat(`, key, `,`, field, `, float64(`, CamelField(strArry[2]), `)).Err()`)
		default:
			return fmt.Errorf("redis.HincrByX's The second param can be int32  int64 float32 float64")
		}
		p.dealErrBool(opt, tp)
	case "nsq.Producer":
		if len(strArry) < 2 {
			return fmt.Errorf("nsq.Producer's param must have 3 ")
		}
		topic, err := p.getString(str1, in, out)
		if err != nil {
			return err
		}
		tp1, _, err := p.getVarType(strArry[1], in, out)
		if err != nil {
			return err
		}
		if tp1 == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			p.P(`pdc, err := (*s.Nsq).Get()`)
			p.P(`defer pdc.Close()`)
			p.dealErrReturn()
			p.P(`_nsqData, err := `, CamelField(strArry[1]), `.Marshal()`)
			p.dealErrReturn()
			p.P(`err = pdc.Publish(`, topic, ` _nsqData)`)
			p.dealErrReturn()
		} else {
			return fmt.Errorf("nsq.Producer: the second param must be message ")
		}
	default:
		if opt.GetSqlxTran() != nil && len(opt.GetSqlxTran()) > 0 {
			p.P(`tx, err := s.Db.Beginx()`)
			p.dealErrReturn()
			for _, o := range opt.GetSqlxTran() {
				strArry := strings.Split(o.GetParam(), ";")
				str1 := strings.TrimSpace(strArry[0])
				str2 := " "
				for _, s := range strArry[1:] {
					str2 += ","
					str2 += CamelField(strings.TrimSpace(s))
				}
				_, lb1, err := p.getVarType(o.GetSlice(), in, out)
				if err != nil {
					return err
				}
				if o.Method == "sqlx.Exec" {
					if lb1 == descriptor.FieldDescriptorProto_LABEL_REPEATED {
						p.P(`for _, obj := range `, CamelField(o.GetSlice()), `{`)
						p.In()
						p.P(`_, err = s.Db.Exec(`, str1, str2, `)`)
						p.P(`if err != nil {`)
						p.In()
						p.P(`tx.Rollback()`)
						p.P(`s.Log.Error(err.Error())`)
						p.P(`return out, err`)
						p.Out()
						p.P(`}`)
						p.Out()
						p.P(`}`)

					} else {
						p.P(`_, err = s.Db.Exec(`, str1, str2, `)`)
						p.P(`if err != nil {`)
						p.In()
						p.P(`tx.Rollback()`)
						p.P(`s.Log.Error(err.Error())`)
						p.P(`return out, err`)
						p.Out()
						p.P(`}`)
					}

				} else if o.Method == "sqlx.NExec" {
					if lb1 == descriptor.FieldDescriptorProto_LABEL_REPEATED {
						p.P(`for _, obj := range `, CamelField(o.GetSlice()), `{`)
						p.In()
						p.P(`_, err = s.Db.NamedExec(`, str1, str2, `)`)
						p.P(`if err != nil {`)
						p.In()
						p.P(`tx.Rollback()`)
						p.P(`s.Log.Error(err.Error())`)
						p.P(`return out, err`)
						p.Out()
						p.P(`}`)
						p.Out()
						p.P(`}`)

					} else {
						p.P(`_, err = s.Db.NamedExec(`, str1, str2, `)`)
						p.P(`if err != nil {`)
						p.In()
						p.P(`tx.Rollback()`)
						p.P(`s.Log.Error(err.Error())`)
						p.P(`return out, err`)
						p.Out()
						p.P(`}`)
					}

				} else {
					err = fmt.Errorf("Does not support functions: %s", opt.GetMethod())
				}

			}
			p.P(`tx.Commit()`)
		} else if opt.GetMzset() != nil {

		} else {
			err = fmt.Errorf("Does not support functions: %s", opt.GetMethod())
		}
	}
	if end {
		p.Out()
		if els {
			p.P(`} else {`)
			p.In()
		} else {
			p.P(`}`)
		}
	}
	if opt.Success != nil && opt.Failure != nil {
		p.P(`if err != nil {`)
		p.In()
		p.P(`s.Log.Error(err.Error())`)
		err = p.dealMethod(opt.Failure, true, true, in, out)
		if err == nil {
			err = p.dealMethod(opt.Success, true, false, in, out)
		}
	} else if opt.Failure != nil {
		p.P(`if err != nil {`)
		p.In()
		p.P(`s.Log.Error(err.Error())`)
		err = p.dealMethod(opt.Failure, true, false, in, out)
	} else if opt.Success != nil {
		err = p.dealMethod(opt.Success, false, false, in, out)
	}
	return err
}
func (p *RormPlugin) dealErrReturn() {
	p.P(`if err != nil {`)
	p.In()
	p.P(`s.Log.Error(err.Error())`)
	p.P(`return out, err`)
	p.Out()
	p.P(`}`)
}
func (p *RormPlugin) dealErrBool(opt *options.RormOptions, tp descriptor.FieldDescriptorProto_Type) {
	if opt.Failure == nil {
		p.P(`if err != nil {`)
		p.In()
		p.P(`s.Log.Error(err.Error())`)
		p.P(`return out, err`)
		p.Out()
		p.P(`}`)
		if descriptor.FieldDescriptorProto_TYPE_BOOL == tp {
			p.P(CamelField(opt.Target), ` = true`)
		}
	} else {
		if descriptor.FieldDescriptorProto_TYPE_BOOL == tp {
			p.P(`if err == nil {`)
			p.In()
			p.P(CamelField(opt.Target), ` = true`)
			p.Out()
			p.P(`}`)
		}
	}
}
func (p *RormPlugin) getString(str, in, out string) (string, error) {
	s := strings.Replace(str, " ", "", -1)
	ss := strings.Split(s, "+")
	res := ""
	for _, st := range ss {
		if strings.Contains(st, `"`) {
			if res != "" {
				res += " + "
			}
			res += st
		} else {
			vars := strings.Split(st, ".")
			if vars[0] != "in" {
				return "", fmt.Errorf("param %s is not valide", st)
			}
			msg := p.msgMap[in]
			var tp descriptor.FieldDescriptorProto_Type
			for _, f := range vars[1:] {
				fd := msg.GetFieldDescriptor(strings.TrimSpace(f))
				tp = fd.GetType()
				if tp == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
					_, ok := p.msgMap[generator.CamelCase(fd.GetTypeName())]
					if !ok {
						return "", fmt.Errorf("can not find message %s in this file", fd.GetTypeName())
					}
					msg = p.msgMap[generator.CamelCase(fd.GetTypeName())]
				}
			}

			switch tp {
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				if res != "" {
					res += " + "
				}
				res += CamelField(st)
			case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_UINT64,
				descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_FIXED64,
				descriptor.FieldDescriptorProto_TYPE_FIXED32:
				if res != "" {
					res += " + "
				}
				p.imports["strconv"] = "strconv"
				res += "strconv.Itoa(int(" + CamelField(st) + "))"
			case descriptor.FieldDescriptorProto_TYPE_FLOAT:
				if res != "" {
					res += " + "
				}
				p.imports["strconv"] = "strconv"
				res += "strconv.FormatFloat(" + CamelField(st) + ",'f',-1,32)"
			case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
				if res != "" {
					res += " + "
				}
				p.imports["strconv"] = "strconv"
				res += "strconv.FormatFloat(" + CamelField(st) + ",'f',-1,64)"
			default:
				return "", fmt.Errorf("field %s can not to string", st)
			}
		}
	}
	return res, nil
}

func (p *RormPlugin) getVarType(st string, in, out string) (descriptor.FieldDescriptorProto_Type, descriptor.FieldDescriptorProto_Label, error) {
	if st == "" {
		return 0, 0, nil
	}
	vars := strings.Split(st, ".")
	var msg *generator.Descriptor
	if vars[0] == "in" {
		msg = p.msgMap[in]
	} else if vars[0] == "out" {
		msg = p.msgMap[in]
	} else {
		return 0, 0, fmt.Errorf("target must start with  'in' or 'out' ")
	}
	if len(vars) == 1 {
		return descriptor.FieldDescriptorProto_TYPE_MESSAGE, descriptor.FieldDescriptorProto_LABEL_OPTIONAL, nil
	}
	var tp descriptor.FieldDescriptorProto_Type
	var lb descriptor.FieldDescriptorProto_Label
	for _, f := range vars[1:] {
		fd := msg.GetFieldDescriptor(strings.TrimSpace(f))
		tp = fd.GetType()
		lb = fd.GetLabel()
		if tp == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			_, ok := p.msgMap[generator.CamelCase(fd.GetTypeName())]
			if !ok {
				return 0, 0, fmt.Errorf("can not find message %s in this file", fd.GetTypeName())
			}
			msg = p.msgMap[generator.CamelCase(fd.GetTypeName())]
		}
	}
	return tp, lb, nil
}
