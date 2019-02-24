#!/bin/sh
protoc \
  -I=. \
  -I=${GOPATH}/src \
  -I=${GOPATH}/src/github.com/gogo/protobuf/protobuf \
  -I=${GOPATH}/src/myfoodie/monitor \
  --gogofaster_out=plugins=grpc:./ \
  --rorm_out=,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:./ \
  --validate_out="lang=gogo:." \
  --ts_out=:./  prod.proto

protoc-go-inject-tag -input=./prod.pb.go