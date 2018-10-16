
protoc \
  -I=. \
  -I=${GOPATH}/src \
  -I=${GOPATH}/src/github.com/lyft/protoc-gen-validate \
  -I=${GOPATH}/src/github.com/gogo/protobuf/protobuf \
  --gofast_out=plugins=grpc:. \
  --validate_out="lang=gogo:." \
  --rorm_out=. ./example/prod.proto