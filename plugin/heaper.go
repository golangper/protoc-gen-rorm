package plugin

import (
	// "fmt"
	proto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

	// "github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/golangper/protoc-gen-rorm/options"
)

func GetUidExtension(opt *descriptor.MethodOptions) *options.UidOptions {
	val, err := proto.GetExtension(opt, options.E_Uid)
	if err != nil {
		return nil
	}
	if val == nil {
		return nil
	}
	return val.(*options.UidOptions)
}

func GetApiExtension(opt *descriptor.MethodOptions) *options.RormApi {
	val, err := proto.GetExtension(opt, options.E_Api)
	if err != nil {
		return nil
	}
	if val == nil {
		return nil
	}
	return val.(*options.RormApi)
}

func GetOptsExtension(opt *descriptor.MethodOptions) *options.RormOptions {
	val, err := proto.GetExtension(opt, options.E_Opts)
	if err != nil {
		return nil
	}
	if val == nil {
		return nil
	}
	return val.(*options.RormOptions)
}
