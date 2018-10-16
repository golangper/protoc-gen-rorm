package example

import  "golang.org/x/net/context"
import  "github.com/jmoiron/sqlx"
import  "net/http"
import  "github.com/gin-gonic/gin"
import  "github.com/gin-gonic/gin/binding"
import  "github.com/op/go-logging"
import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/golangper/protoc-gen-rorm/options"



// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type _ProductImp struct {
        db *sqlx.DB
        log *logging.Logger
}

func (s *_ProductImp) GetProd(c context.Context, in *example.ProdId) (*example.Prod, error) {
        var err error
        out := &example.Prod{}
        err = in.Validate()
        if err != nil {
                return out, err
        }
        stmtout err := s.db.PrepareNamed("select * from prod where id = ?")
        if err != nil {
                s.log.Error(err.Error())
                return out, err
        }
        err = stmtout.Get(out ,in.Id)
        if err != nil {
                s.log.Error(err.Error())
                return out, err
        }
        stmtSkus err := s.db.PrepareNamed("select * from sku where prod_id=?")
        if err != nil {
                s.log.Error(err.Error())
                return out, err
        }
        err = stmtSkus.Select( &out.Skus ,in.Id)
        if err != nil {
                s.log.Error(err.Error())
                return out, err
        }
        return out, nil
}

func (s *_ProductImp) SetProd(c context.Context, in *example.Prod) (*example.empty, error) {
        var err error
        out := &example.empty{}
        err = in.Validate()
        if err != nil {
                return out, err
        }
        _s := in.Id % 256
        _worker, err := snowflake.NewChannelWorker(s)
        if err != nil {
                return out, err
        }
        uid , _ := _worker.Next()
        var _ = uid
        tx, err := s.db.Beginx()
        if err != nil {
                s.log.Error(err.Error())
                return out, err
        }
        _, err = s.db.Exec("insert into prod (id,name,details) values (?,?,?)" ,uid,in.Name,in.Details)
        if err != nil {
                tx.Rollback()
                s.log.Error(err.Error())
                return out, err
        }
        for _, obj := range in.Skus{
                _, err = s.db.Exec("insert into sku (sku_id,price,bn,weight,prod_id) values (?,?,?,?,?)" ,obj.SkuId,obj.Price,obj.Bn,obj.Weight,in.Id)
                if err != nil {
                        tx.Rollback()
                        s.log.Error(err.Error())
                        return out, err
                }
        }
        tx.Commit()
        return out, nil
}

type ProductImp struct {
        _ProductImp
}

func NewProductImp(db *sqlx.DB, log *logging.Logger) ProductImp {
        res := ProductImp{}
        res.db = db
        res.log = log
        return res
}

func (s *ProductImp) GetProdHandler(c *gin.Context) {
        var prm *example.ProdId
        var err error
        err = c.ShouldBindWith(prm, binding.JSON)
        if err != nil {
                s.log.Error(err.Error())
                c.JSON(http.StatusBadRequest,gin.H{"resp": err.Error()})
                return
        }
        if err = prm.Validate(); err != nil {
                s.log.Error(err.Error())
                c.JSON(http.StatusBadRequest,gin.H{"resp": err.Error()})
                return
        }
        res, err := s.GetProd(context.Background(), prm)
        if err != nil {
                s.log.Error(err.Error())
                c.JSON(http.StatusServiceUnavailable,gin.H{"resp": err.Error()})
                return
        }
        c.JSON(http.StatusOK,gin.H{"resp": res})
}

func (s *ProductImp) SetProdHandler(c *gin.Context) {
        var prm *example.Prod
        var err error
        err = c.ShouldBindWith(prm, binding.JSON)
        if err != nil {
                s.log.Error(err.Error())
                c.JSON(http.StatusBadRequest,gin.H{"resp": err.Error()})
                return
        }
        if err = prm.Validate(); err != nil {
                s.log.Error(err.Error())
                c.JSON(http.StatusBadRequest,gin.H{"resp": err.Error()})
                return
        }
        res, err := s.SetProd(context.Background(), prm)
        if err != nil {
                s.log.Error(err.Error())
                c.JSON(http.StatusServiceUnavailable,gin.H{"resp": err.Error()})
                return
        }
        c.JSON(http.StatusOK,gin.H{"resp": res})
}