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
	file      *generator.FileDescriptor
}

type Api struct {
	method   string
	path     string
	funcName string
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
	p.file = file
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
	
		p.useNsq = proto.GetBoolExtension(svc.Options, options.E_UseNsq, false)

		if p.useNsq {
			p.imports["nsqpool"] = "github.com/qgymje/nsqpool"
		}

		gin := proto.GetBoolExtension(svc.Options, options.E_GinHandler, false)

		if gin {
			p.imports["http"] = "net/http"
			p.imports["gin"] = "github.com/gin-gonic/gin"
			// p.imports["binding"] = "github.com/gin-gonic/gin/binding"
		}
		p.imports["logging"] = "github.com/op/go-logging"
		p.imports["context"] = "golang.org/x/net/context"
		grpcSvcName := "_" + generator.CamelCase(svc.GetName()) + "Imp"
		impName := generator.CamelCase(svc.GetName()) + "Imp"
		//grpc impl struct
		p.P(`type `, grpcSvcName, ` struct {`)
		p.In()
		if p.sqlType == 1 {
			p.P(`db *sqlx.DB`)
		} else if p.sqlType == 2 {
			p.P(`db *sqlt.DB`)
		}
		if p.redisType == 1 {
			p.P(`redis *redis.Client`)
		} else if p.redisType == 2 {
			p.P(`redis *redis.ClusterClient`)
		}
		if p.useNsq {
			p.P(`nsq *pool.Pool`)
		}
		p.P(`log *logging.Logger`)
		p.Out()
		p.P(`}`)

		apilist := make([]*Api, 0)
		//grpc method impl
		for _, method := range svc.GetMethod() {
			mname := generator.CamelCase(method.GetName())

			inputType := generator.CamelCase(method.GetInputType())
			//outputType := generator.CamelCase(method.GetOutputType())

			uid := GetUidExtension(method.Options)
			opts := GetOptsExtension(method.Options)
			inputMsg := p.file.GetMessage(GetMessageName(method.GetInputType()))
			if inputMsg == nil {
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
			// in := inputType[1:]
			// out := outputType[1:]
			p.P(``)
			p.P(`func (s *`, grpcSvcName, `) `, mname, `(c context.Context, in *`, generator.CamelCase(GetMessageName(method.GetInputType())), `) (*`, generator.CamelCase(GetMessageName(method.GetOutputType())), `, error) {`)
			p.In()
			p.P(`var err error`)
			p.outAndValid(GetMessageName(method.GetInputType()), GetMessageName(method.GetOutputType()))
			if uid != nil {
				p.newUid(uid)
			}
			if opts != nil {
				p.dealMethods(opts, GetMessageName(method.GetInputType()), GetMessageName(method.GetOutputType()))
			}

			p.P(`return out, nil`)
			p.Out()
			p.P(`}`)

		}
		p.P(``)
		p.P(`type `, impName, ` struct {`)
		p.In()
		p.P(grpcSvcName)
		p.Out()
		p.P(`}`)

		prm := ""
		if p.sqlType == 1 {
			if prm != "" {
				prm += ", "
			}
			prm += `db *sqlx.DB`
		} else if p.sqlType == 2 {
			if prm != "" {
				prm += ", "
			}
			prm += `db *sqlt.DB`
		}
		if p.redisType == 1 {
			if prm != "" {
				prm += ", "
			}
			prm += `redis *redis.Client`
		} else if p.redisType == 2 {
			if prm != "" {
				prm += ", "
			}
			prm += `redis *redis.ClusterClient`
		}
		if p.useNsq {
			if prm != "" {
				prm += ", "
			}
			prm += `nsq *pool.Pool`
		}
		if prm != "" {
			prm += ", "
		}
		prm += `log *logging.Logger`
		p.P(``)
		p.P(`func New`, impName, `(`, prm, `) `, impName, ` {`)
		p.In()
		p.P(`res := `, impName, `{}`)
		if p.sqlType > 0 {
			p.P(`res.db = db`)
		}
		if p.redisType > 0 {
			p.P(`res.redis = redis`)
		}
		if p.useNsq {
			p.P(`res.nsq = nsq`)
		}
		p.P(`res.log = log`)
		p.P(`return res`)
		p.Out()
		p.P(`}`)
		p.P(``)
	
		for _, m := range svc.GetMethod() {
			mname := generator.CamelCase(m.GetName())
			//inputType := generator.CamelCase(m.GetInputType())
			api := GetApiExtension(m.Options)
			if api != nil {
				myapi := &Api{method: api.Method, path: api.Path, funcName: mname + "GinHandler"}
				apilist = append(apilist, myapi)
			}

			p.P(``)
			p.P(`func (s *`, impName, `) `, mname, `GinHandler(c *gin.Context) {`)
			p.In()

			p.P(`var prm *`, generator.CamelCase(GetMessageName(m.GetInputType())))
			p.P(`var err error`)

			p.P(`err = c.ShouldBind(prm)`)
			p.P(`if err != nil {`)
			p.In()
			p.P(`s.log.Error(err.Error())`)
			p.P(`c.JSON(http.StatusBadRequest,gin.H{"resp": err.Error()})`)
			p.P(`return`)
			p.Out()
			p.P(`}`)

			p.P(`if err = prm.Validate(); err != nil {`)
			p.In()
			p.P(`s.log.Error(err.Error())`)
			p.P(`c.JSON(http.StatusBadRequest,gin.H{"resp": err.Error()})`)
			p.P(`return`)
			p.Out()
			p.P(`}`)

			p.P(`res, err := s.`, mname, `(context.Background(), prm)`)
			p.P(`if err != nil {`)
			p.In()
			p.P(`s.log.Error(err.Error())`)
			p.P(`c.JSON(http.StatusServiceUnavailable,gin.H{"resp": err.Error()})`)
			p.P(`return`)
			p.Out()
			p.P(`}`)
			p.P(`c.JSON(http.StatusOK,gin.H{"resp": res})`)
			p.Out()
			p.P(`}`)
		}
		// p.GenerateImports(file)
		p.P(``)
		p.P(`func (s *`, impName, `) InitApi(g *gin.Engine) {`)
		p.In()
		for _, l := range apilist {
			if l.method == "post" || l.method == "POST" || l.method == "Post" {
				p.P(`g.POST("`, l.path, `", s.`, l.funcName, `)`)
			} else if l.method == "get" || l.method == "GET" || l.method == "Get"{
				p.P(`g.GET("`, l.path, `", s.`, l.funcName, `)`)
			} else {
				fmt.Println("not not support the method", l.method)
			}
		}
		p.Out()
		p.P(`}`)

	}
}

func (p *RormPlugin) outAndValid(in, out string) {
	if in == out {
		p.P(`out := in`)
	} else {
		p.P(`out := &`, generator.CamelCase(out), `{}`)
	}
	p.P(`err = in.Validate()`)
	p.P(`if err != nil {`)
	p.In()
	p.P(`return out, err`)
	p.Out()
	p.P(`}`)
}

func (p *RormPlugin) newUid(uid *options.UidOptions) {
	p.imports["snowflake"] = "github.com/fainted/snowflake"
	strs := strings.Split(uid.Seed, ".")
	f := strs[len(strs)-1]
	p.P(`_s := in.`, generator.CamelCase(f), ` % 256`)
	p.P(`_worker, err := snowflake.NewChannelWorker(_s)`)
	p.P(`if err != nil {`)
	p.In()
	p.P("return out, err")
	p.Out()
	p.P(`}`)
	p.P(uid.Name, ` , _ := _worker.Next()`)
	p.P(`var _ = `, uid.Name)
}

func (p *RormPlugin) dealMethods(opts *options.RormOptions, in, out string) {
	err := p.dealMethod(opts, false, false, in, out)
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

func (p *RormPlugin) dealMethod(opt *options.RormOptions, end bool, els bool, in, out string) error {
	var err error
	if opt.GetParam() == "" && opt.GetSqlxTran() == nil {
		return fmt.Errorf("param can not bu null")
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
	tp, lb, sl, err := p.getVarType(opt.GetTarget(), in, out)
	if err != nil {
		return err
	}
	switch opt.GetMethod() {
	case "sqlx.Exec":
		if lb == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("sqlx.Exec's target can not be repeated ")
		}
		p.P(`_, err = s.db.Exec(`, str1, str2, `)`)
		p.dealErrBool(opt, tp)
	case "sqlx.Get":
		if lb == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("sqlx.GetMessage's target can not be repeated ")
		}
		if tp == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			if sl != "" {
				p.P(`for _, obj := range `, CamelField(sl), `{`)
				p.In()
				//s := strings.Split(opt.GetTarget(), ".")
				p.P(`err = s.db.Get(`, CamelField(opt.GetTarget()), ` , `, str1, str2, `)`)
				p.Out()
				p.P(`}`)
			} else {
				p.P(`err = s.db.Get(`, CamelField(opt.GetTarget()), ` , `, str1, str2, `)`)
			}

		} else {
			if sl != "" {
				p.P(`for _, obj := range `, CamelField(sl), `{`)
				p.In()
				//s := strings.Split(opt.GetTarget(), ".")
				p.P(`err = s.db.Get( &`, CamelField(opt.GetTarget()), ` , `, str1, str2, `)`)
				p.Out()
				p.P(`}`)
			} else {
				p.P(`err = s.db.Get( &`, CamelField(opt.GetTarget()), ` , `, str1, str2, `)`)
			}
		}
		if opt.Failure == nil {
			p.dealErrReturn()
		}
	case "sqlx.Select":
		if lb != descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("sqlx.Select's target must be repeated ")
		}
		if sl != "" {
			p.P(`for _, obj := range `, CamelField(sl), `{`)
			p.In()
			// s := strings.Split(opt.GetTarget(), ".")
			p.P(`err = s.db.Select( &`, CamelField(opt.GetTarget()), ` , `, str1, str2, `)`)
			p.Out()
			p.P(`}`)
		} else {
			p.P(`err = s.db.Select( &`, CamelField(opt.GetTarget()), ` , `, str1, str2, `)`)
		}

		if opt.Failure == nil {
			p.dealErrReturn()
		}
	case "sqlx.PGet":
		if lb == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("sqlx.PGetMessage's target can not be repeated ")
		}
		t := strings.Split(CamelField(opt.GetTarget()), ".")
		n := t[len(t)-1]
		p.P(`stmt`, n, `, err := s.db.Preparex(`, str1, `)`)
		p.dealErrReturn()
		if tp == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			if sl != "" {
				p.P(`for _, obj := range `, CamelField(sl), `{`)
				p.In()
				p.P(`err = stmt`, n, `.Get(`, CamelField(opt.GetTarget()), str2, `)`)
				p.Out()
				p.P(`}`)
			} else {
				p.P(`err = stmt`, n, `.Get(`, CamelField(opt.GetTarget()), str2, `)`)
			}

		} else {
			if sl != "" {
				p.P(`for _, obj := range `, CamelField(sl), `{`)
				p.In()
				p.P(`err = stmt`, n, `.Get(&`, CamelField(opt.GetTarget()), str2, `)`)
				p.Out()
				p.P(`}`)
			} else {
				p.P(`err = stmt`, n, `.Get(&`, CamelField(opt.GetTarget()), str2, `)`)
			}
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
		p.P(`stmt`, n, `, err := s.db.Preparex(`, str1, `)`)
		p.dealErrReturn()
		if sl != "" {
			p.P(`for _, obj := range `, CamelField(sl), `{`)
			p.In()
			p.P(`err = stmt`, n, `.Select( &`, CamelField(opt.GetTarget()), str2, `)`)
			p.Out()
			p.P(`}`)
		} else {
			p.P(`err = stmt`, n, `.Select( &`, CamelField(opt.GetTarget()), str2, `)`)
		}

		if opt.Failure == nil {
			p.dealErrReturn()
		}
	case "sqlx.NExec":
		if lb == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("sqlx.NExec's target can not be repeated ")
		}
		p.P(`_, err = s.db.NamedExec(`, str1, str2, `)`)
		p.dealErrBool(opt, tp)
	case "sqlx.PNGet":
		if lb == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			return fmt.Errorf("sqlx.PNGetMessage's target can not be repeated ")
		}
		t := strings.Split(CamelField(opt.GetTarget()), ".")
		n := t[len(t)-1]
		p.P(`stmt`, n, `, err := s.db.PrepareNamed(`, str1, `)`)
		p.dealErrReturn()
		if tp == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			if sl != "" {
				p.P(`for _, obj := range `, CamelField(sl), `{`)
				p.In()
				p.P(`err = stmt`, n, `.Get(`, CamelField(opt.GetTarget()), str2, `)`)
				p.Out()
				p.P(`}`)
			} else {
				p.P(`err = stmt`, n, `.Get(`, CamelField(opt.GetTarget()), str2, `)`)
			}

		} else {
			if sl != "" {
				p.P(`for _, obj := range `, CamelField(sl), `{`)
				p.In()
				p.P(`err = stmt`, n, `.Get(&`, CamelField(opt.GetTarget()), str2, `)`)
				p.Out()
				p.P(`}`)
			} else {
				p.P(`err = stmt`, n, `.Get(&`, CamelField(opt.GetTarget()), str2, `)`)
			}
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
		p.P(`stmt`, n, `, err := s.db.PrepareNamed(`, str1, `)`)
		p.dealErrReturn()
		if sl != "" {
			p.P(`for _, obj := range `, CamelField(sl), `{`)
			p.In()
			p.P(`err = stmt`, n, `.Select( &`, CamelField(opt.GetTarget()), str2, `)`)
			p.Out()
			p.P(`}`)
		} else {
			p.P(`err = stmt`, n, `.Select( &`, CamelField(opt.GetTarget()), str2, `)`)
		}

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
			p.P(`rds`, n, `, err := s.redis.Get(`, key, `).Bytes()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}

			p.P(`err = `, CamelField(opt.GetTarget()), `.Unmarshal(rds`, n, `)`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			p.P(CamelField(opt.GetTarget()), `, err := s.redis.Get(`, key, `).String()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_INT64:
			p.P(CamelField(opt.Target), `, err := s.redis.Get(`, key, `).Int64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
			p.P(`rds`, n, `, err := s.redis.Get(`, key, `).Int64()`)
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
			p.P(`rds`, n, `, err := s.redis.Get(`, key, `).Int64()`)
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
			p.P(`rds`, n, `, err := s.redis.Get(`, key, `).Int64()`)
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
			p.P(CamelField(opt.Target), `, err := s.redis.Get(`, key, `).Float64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			p.P(`rds`, n, `, err := s.redis.Get(`, key, `).Float64()`)
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
		tp1, _, _, err := p.getVarType(strArry[1], in, out)
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
			p.P(`err = s.redis.Set(`, key, `,`, `set`, n, `,int64(time.Duration(`, strArry[2], `) * time.Second)).Err()`)
		} else if tp1 == descriptor.FieldDescriptorProto_TYPE_BOOL {
			return fmt.Errorf("redis.Set's target can not be bool ")
		} else {
			p.P(`err = s.redis.Set(`, key, `,`, CamelField(strArry[1]), `,int64(time.Duration(`, strArry[2], `) * time.Second)).Err()`)
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
		p.P(`err = s.redis.Del(`, param, `).Err()`)
		p.dealErrBool(opt, tp)
	case "redis.IncrByX":
		if len(strArry) < 2 {
			return fmt.Errorf("redis.IncrByX's param must have 2 ")
		}
		key, err := p.getString(str1, in, out)
		if err != nil {
			return err
		}
		tp1, _, _, err := p.getVarType(opt.GetTarget(), in, out)
		if err != nil {
			return err
		}
		switch tp1 {
		case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_INT32:
			p.P(`err = s.redis.IncrBy(`, key, `, int64(`, CamelField(strArry[1]), `)).Err()`)
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE, descriptor.FieldDescriptorProto_TYPE_FLOAT:
			p.P(`err = s.redis.IncrByFloat(`, key, `, float64(`, CamelField(strArry[1]), `)).Err()`)
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
		tp1, _, _, err := p.getVarType(opt.GetTarget(), in, out)
		if err != nil {
			return err
		}
		switch tp1 {
		case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_INT32:
			p.P(`_dnum,  err := s.redis.Get(`, key, `).Int64()`)
			p.dealErrReturn()
			p.P(`if int(_dnum) < int(`, CamelField(strArry[1]), `){`)
			p.In()
			p.P(`return out, fmt.Errorf("Inventory shortage")`)
			p.Out()
			p.P(`}`)
			p.P(`err = s.redis.IncrBy(`, key, `, int64(`, CamelField(strArry[1]), `)).Err()`)
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
		p.P(`err = s.redis.Expire(`, key, `, int64(time.Duration(`, strArry[1], `) * time.Second)).Err()`)
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
			p.P(`rds`, n, `, err := s.redis.HGet(`, key, `,`, field, `).Bytes()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}

			p.P(`err = `, CamelField(opt.GetTarget()), `.Unmarshal(rds`, n, `)`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			p.P(CamelField(opt.GetTarget()), `, err := s.redis.HGet(`, key, `,`, field, `).String()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_INT64:
			p.P(CamelField(opt.GetTarget()), `, err := s.redis.HGet(`, key, `,`, field, `).Int64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
			p.P(`rds`, n, `, err := s.redis.HGet(`, key, `,`, field, `).Int64()`)
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
			p.P(`rds`, n, `, err := s.redis.HGet(`, key, `,`, field, `).Int64()`)
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
			p.P(`rds`, n, `, err := s.redis.HGet(`, key, `,`, field, `).Int64()`)
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
			p.P(CamelField(opt.GetTarget()), `, err := s.redis.HGet(`, key, `,`, field, `).Float64()`)
			if opt.Failure == nil {
				p.dealErrReturn()
			}
		case descriptor.FieldDescriptorProto_TYPE_FLOAT:
			p.P(`rds`, n, `, err := s.redis.HGet(`, key, `,`, field, `).Float64()`)
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
		_, lb2, _, err := p.getVarType(strArry[1], in, out)
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

		tp1, lb1, _, err := p.getVarType(strArry[2], in, out)
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
			p.P(`err = s.redis.HSet(`, key, `,`, field, `, set`, n, `).Err()`)
		} else if tp1 == descriptor.FieldDescriptorProto_TYPE_BOOL {
			return fmt.Errorf("redis.Set's target can not be bool ")
		} else {
			p.P(`err = s.redis.HSet(`, key, `,`, field, `,`, CamelField(strArry[2]), `).Err()`)
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
		p.P(`err := s.redis.HDel(`, key, `,`, field, `).Err()`)
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
		tp2, _, _, err := p.getVarType(strArry[2], in, out)
		if err != nil {
			return err
		}
		switch tp2 {
		case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_INT32:
			p.P(`err = s.redis.HIncrBy(`, key, `,`, field, `, int64(`, CamelField(strArry[2]), `)).Err()`)
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE, descriptor.FieldDescriptorProto_TYPE_FLOAT:
			p.P(`err = s.redis.HIncrByFloat(`, key, `,`, field, `, float64(`, CamelField(strArry[2]), `)).Err()`)
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
		tp1, _, _, err := p.getVarType(strArry[1], in, out)
		if err != nil {
			return err
		}
		if tp1 == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			p.P(`pdc, err := (*s.nsq).Get()`)
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
			p.P(`tx, err := s.db.Beginx()`)
			p.dealErrReturn()
			for _, o := range opt.GetSqlxTran() {
				str := strings.Replace(o.GetParam(), `'`, `"`, -1)
				strArry := strings.Split(str, ";")
				str1 := strings.TrimSpace(strArry[0])
				str2 := " "
				for _, s := range strArry[1:] {
					str2 += ","
					str2 += CamelField(strings.TrimSpace(s))
				}
				_, lb1, _, err := p.getVarType(o.GetSlice(), in, out)
				if err != nil {
					return err
				}
				if o.Method == "sqlx.Exec" {
					if lb1 == descriptor.FieldDescriptorProto_LABEL_REPEATED {
						p.P(`for _, obj := range `, CamelField(o.GetSlice()), `{`)
						p.In()
						p.P(`_, err = s.db.Exec(`, str1, str2, `)`)
						p.P(`if err != nil {`)
						p.In()
						p.P(`tx.Rollback()`)
						p.P(`s.log.Error(err.Error())`)
						p.P(`return out, err`)
						p.Out()
						p.P(`}`)
						p.Out()
						p.P(`}`)

					} else {
						p.P(`_, err = s.db.Exec(`, str1, str2, `)`)
						p.P(`if err != nil {`)
						p.In()
						p.P(`tx.Rollback()`)
						p.P(`s.log.Error(err.Error())`)
						p.P(`return out, err`)
						p.Out()
						p.P(`}`)
					}

				} else if o.Method == "sqlx.NExec" {
					if lb1 == descriptor.FieldDescriptorProto_LABEL_REPEATED {
						p.P(`for _, obj := range `, CamelField(o.GetSlice()), `{`)
						p.In()
						p.P(`_, err = s.db.NamedExec(`, str1, str2, `)`)
						p.P(`if err != nil {`)
						p.In()
						p.P(`tx.Rollback()`)
						p.P(`s.log.Error(err.Error())`)
						p.P(`return out, err`)
						p.Out()
						p.P(`}`)
						p.Out()
						p.P(`}`)

					} else {
						p.P(`_, err = s.db.NamedExec(`, str1, str2, `)`)
						p.P(`if err != nil {`)
						p.In()
						p.P(`tx.Rollback()`)
						p.P(`s.log.Error(err.Error())`)
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
		p.P(`s.log.Error(err.Error())`)
		err = p.dealMethod(opt.Failure, true, true, in, out)
		if err == nil {
			err = p.dealMethod(opt.Success, true, false, in, out)
		}
	} else if opt.Failure != nil {
		p.P(`if err != nil {`)
		p.In()
		p.P(`s.log.Error(err.Error())`)
		err = p.dealMethod(opt.Failure, true, false, in, out)
	} else if opt.Success != nil {
		err = p.dealMethod(opt.Success, false, false, in, out)
	}
	return err
}
func (p *RormPlugin) dealErrReturn() {
	p.P(`if err != nil {`)
	p.In()
	p.P(`s.log.Error(err.Error())`)
	p.P(`return out, err`)
	p.Out()
	p.P(`}`)
}
func (p *RormPlugin) dealErrBool(opt *options.RormOptions, tp descriptor.FieldDescriptorProto_Type) {
	if opt.Failure == nil {
		p.P(`if err != nil {`)
		p.In()
		p.P(`s.log.Error(err.Error())`)
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
			msg := p.file.GetMessage(in)
			var tp descriptor.FieldDescriptorProto_Type
			for _, f := range vars[1:] {
				fd := msg.GetFieldDescriptor(strings.TrimSpace(f))
				tp = fd.GetType()
				if tp == descriptor.FieldDescriptorProto_TYPE_MESSAGE {

					if msg == nil {
						return "", fmt.Errorf("can not find message %s in this file", fd.GetTypeName())
					}
					msg = p.file.GetMessage(GetMessageName(fd.GetTypeName()))
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

func (p *RormPlugin) getVarType(st string, in, out string) (descriptor.FieldDescriptorProto_Type, descriptor.FieldDescriptorProto_Label, string, error) {
	if st == "" {
		return 0, 0, "", nil
	}
	vars := strings.Split(st, ".")
	var msg *descriptor.DescriptorProto
	if vars[0] == "in" {
		msg = p.file.GetMessage(in)
	} else if vars[0] == "out" {
		msg = p.file.GetMessage(out)
	} else {
		return 0, 0, "", fmt.Errorf("target must start with  'in' or 'out' ")
	}
	if len(vars) == 1 {
		return descriptor.FieldDescriptorProto_TYPE_MESSAGE, descriptor.FieldDescriptorProto_LABEL_OPTIONAL, "", nil
	}
	var tp descriptor.FieldDescriptorProto_Type
	var lb descriptor.FieldDescriptorProto_Label
	var sl string
	for i, f := range vars {
		if i == 0 {
			sl += f
			continue
		}
		if i < len(vars)-1 {
			sl += "."
			sl += strings.TrimSpace(f)
		}

		fd := msg.GetFieldDescriptor(strings.TrimSpace(f))
		if fd == nil {
			return 0, 0, "", fmt.Errorf("can not find field %s in this file", strings.TrimSpace(f))
		}
		tp = fd.GetType()
		lb = fd.GetLabel()
		if i == len(vars)-2 {
			if lb != descriptor.FieldDescriptorProto_LABEL_REPEATED {
				sl = ""
			}
		}
		if len(vars) < 3 {
			sl = ""
		}
		if tp == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
			if msg == nil {
				return 0, 0, "", fmt.Errorf("can not find message %s in this file", fd.GetTypeName())
			}
			msg = p.file.GetMessage(GetMessageName(fd.GetTypeName()))
		}
	}
	return tp, lb, sl, nil
}
