package plugin

import (
	"fmt"

	// proto "github.com/gogo/protobuf/proto"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/golangper/protoc-gen-rorm/options"
)

func CheckUidSeed(msg *descriptor.DescriptorProto, uid *options.UidOptions) error {
	strs := strings.Split(uid.Seed, ".")
	f := strs[len(strs)-1]
	for _, field := range msg.Field {
		// fmt.Println(`=========`,field.GetName(),f,field.GetType())
		if field.GetName() == f && field.GetType() == descriptor.FieldDescriptorProto_TYPE_INT64 {
			return nil
		}
	}
	return fmt.Errorf("input message must contain seed field and type must be int64 ")
}

func CamelField(str string) string {
	str = strings.TrimSpace(str)
	if str == "" {
		return ""
	}
	if strings.Contains(str, `"`) {
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
		if i == 0 {
			res += s
		} else {
			res += "."
			res += generator.CamelCase(s)
		}

	}
	return res
}

func checkStrIsNum(str string) bool {
	b := true
	for i := 0; i < len(str); i++ {
		if str[i] < 48 || str[i] > 57 {
			if str[i] == '.' {
				b = true
			} else {
				b = false
			}
		}
	}
	return b
}

func checkStrIsInt(str string) bool {
	b := true
	for i := 0; i < len(str); i++ {
		if str[i] < 48 || str[i] > 57 {
			b = false
		}
	}
	return b
}

var unneededImports = []string{
	"import proto \"github.com/gogo/protobuf/proto\"\n",
	"import _ \"github.com/golangper/protoc-gen-rorm/options\"\n",
	// if needed will be imported with an alias
	"var _ = proto.Marshal\n",
	"var _ = fmt.Errorf\n",
	"var _ = math.Inf\n",
	"import fmt \"fmt\"\n",
	"import math \"math\"\n",
}

// CleanImports removes extraneous imports and lines from a proto response
// file Content
func CleanImports(pFileText *string) *string {
	if pFileText == nil {
		return pFileText
	}
	fileText := *pFileText
	for _, dep := range unneededImports {
		fileText = strings.Replace(fileText, dep, "", -1)
	}
	return &fileText
}

func  GetMessageName(str string) string {
	if str == "" {
		return ""
	}
	s := strings.Split(str, ".")
	return s[len(s) - 1]
}