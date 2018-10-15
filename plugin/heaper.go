package plugin

import (
	"fmt"
	proto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

	// "github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/golangper/protoc-gen-rorm/options"
)

func GetUidExtension(opt *descriptor.MethodOptions) *options.UidOptions {
	val, err := proto.GetExtension(opt, options.E_Uid)
	if err != nil {
		// fmt.Println("GetUidExtension:", err.Error())
		return nil
	}
	if val == nil {
		return nil
	}
	return val.(*options.UidOptions)
}

func GetOptsExtension(opt *descriptor.MethodOptions) *options.RormOptions {
	val, err := proto.GetExtension(opt, options.E_Opts)
	if err != nil {
		fmt.Println("GetOptsExtension:", err.Error())
		return nil
	}
	if val == nil {
		return nil
	}
	return val.(*options.RormOptions)
}


