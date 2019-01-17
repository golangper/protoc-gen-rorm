package plugin

import (
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

func (p *RormPlugin) GenerateImports(file *generator.FileDescriptor) {
	for pkg, m := range p.imports {
		p.PrintImport(pkg, m)
	}
}

func (p *TsPlugin) GenerateImports(file *generator.FileDescriptor) {

}
