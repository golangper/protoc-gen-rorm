package plugin

import "github.com/gogo/protobuf/protoc-gen-gogo/generator"

func (p *RormPlugin) GenerateImports(file *generator.FileDescriptor) {
	var pkg generator.GoPackageName = ""
	for _, m := range p.imports {
		p.PrintImport(pkg, m)
	}
}
