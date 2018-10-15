package main

import (
  "github.com/gogo/protobuf/vanity/command"
  "github.com/golangper/protoc-gen-rorm/plugin"
)

func main() {
  response := command.GeneratePlugin(command.Read(), &plugin.RormPlugin{}, ".service.go")
  // for _, file := range response.GetFile() {
  //   file.Content = plugin.CleanImports(file.Content)
  // }
  command.Write(response)
}