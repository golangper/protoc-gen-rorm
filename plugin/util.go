package plugin

import (
	"fmt"

	// proto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/golangper/protoc-gen-rorm/options"
	"strings"
)

func CheckUidSeed(msg *generator.Descriptor, uid *options.UidOptions) error {
	for _, field := range msg.Field {
		if field.GetName() == uid.Seed && field.GetType() == descriptor.FieldDescriptorProto_TYPE_INT64  {
			return nil
		}
	}
	return fmt.Errorf("input message must contain seed field and type must be int64 ")
}

func CamelField(str string) string  {
	str = strings.TrimSpace(str)
	if str == "" {
		return ""
	}
	if strings.Contains(str,`"`) {
		return str
	}
	if checkStrIsNum(str) {
		return str
	}
	vars := strings.Split(str, ".")
	if vars[0] != "in" && vars[0] != "out" && vars[0] != "obj" {
		return str
	}
    res := ""
	for i, s := range vars {
		if i==0 {
			res += s
		}
		res += "."
		res += generator.CamelCase(s)
	}
	return res
}

func checkStrIsNum(str string) bool {
	b := true
	for i := 0 ; i<len(str) ; i++ {
		if str[i] < 48 || str[i] > 57 {
			if str[i] == '.' {
				b = true
			}else {
				b = false
			}
		}
	}
	return b
}

func checkStrIsInt(str string) bool {
	b := true
	for i := 0 ; i<len(str) ; i++ {
		if str[i] < 48 || str[i] > 57 {
			b = false
		}
	}
	return b
}