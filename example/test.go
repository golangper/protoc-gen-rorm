package prod

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/golangper/protoc-gen-rorm/options"



// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type ProductGrpcImp struct {
        Db *sqlx.Db
        Log *logging.Logger
}
func (s *ProductGrpcImp) GetProd(c context.Context, in *ProdId) (*Prod, error) {
        var err error
        var _ = err
        out := &Prod{}
        err = in.Validate()
        if err != nil {
                return out, err
        }
        err = s.Db.Get(out , "select * from prod where id = ?" ,in.Id)
        if err != nil {
                s.Log.Error(err.Error())
                return out, err
        }
        return out, nil
}
type ProductImp struct {
        ProductGrpcImp
}
func (s *ProductImp) GetProdHandler(c *gin.Context) {
        var prm *ProdId
        var err error
        err = c.ShouldBindWith(prm, binding.JSON)
        if err != nil {
                s.Log.Error(err.Error())
                c.JSON(http.StatusBadRequest,gin.H{"resp": err.Error()}
                return
        }
        if err = prm.Validate(); err != nil {
                s.Log.Error(err.Error())
                c.JSON(http.StatusBadRequest,gin.H{"resp": err.Error()}
                return
        }
        res, err := s.GetProd(context.Background(), prm)
        if err != nil {
                s.Log.Error(err.Error())
                c.JSON(http.StatusServiceUnavailable,gin.H{"resp": err.Error()}
                return
        }
        c.JSON(http.StatusOK,gin.H{"resp": res}
}