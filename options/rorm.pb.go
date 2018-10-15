// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: options/rorm.proto

/*
Package options is a generated protocol buffer package.

It is generated from these files:
	options/rorm.proto

It has these top-level messages:
	RormOptions
	MzsetOptions
	TranOptions
	UidOptions
	Variable
*/
package options

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type RormOptions struct {
	Target   string         `protobuf:"bytes,1,opt,name=target,proto3" json:"target,omitempty"`
	Method   string         `protobuf:"bytes,2,opt,name=method,proto3" json:"method,omitempty"`
	Param    string         `protobuf:"bytes,3,opt,name=param,proto3" json:"param,omitempty"`
	SqlxTran []*TranOptions `protobuf:"bytes,4,rep,name=sqlx_tran,json=sqlxTran" json:"sqlx_tran,omitempty"`
	Mzset    *MzsetOptions  `protobuf:"bytes,5,opt,name=mzset" json:"mzset,omitempty"`
	Success  *RormOptions   `protobuf:"bytes,7,opt,name=success" json:"success,omitempty"`
	Failure  *RormOptions   `protobuf:"bytes,8,opt,name=failure" json:"failure,omitempty"`
}

func (m *RormOptions) Reset()                    { *m = RormOptions{} }
func (m *RormOptions) String() string            { return proto.CompactTextString(m) }
func (*RormOptions) ProtoMessage()               {}
func (*RormOptions) Descriptor() ([]byte, []int) { return fileDescriptorRorm, []int{0} }

func (m *RormOptions) GetTarget() string {
	if m != nil {
		return m.Target
	}
	return ""
}

func (m *RormOptions) GetMethod() string {
	if m != nil {
		return m.Method
	}
	return ""
}

func (m *RormOptions) GetParam() string {
	if m != nil {
		return m.Param
	}
	return ""
}

func (m *RormOptions) GetSqlxTran() []*TranOptions {
	if m != nil {
		return m.SqlxTran
	}
	return nil
}

func (m *RormOptions) GetMzset() *MzsetOptions {
	if m != nil {
		return m.Mzset
	}
	return nil
}

func (m *RormOptions) GetSuccess() *RormOptions {
	if m != nil {
		return m.Success
	}
	return nil
}

func (m *RormOptions) GetFailure() *RormOptions {
	if m != nil {
		return m.Failure
	}
	return nil
}

type MzsetOptions struct {
	Target string `protobuf:"bytes,1,opt,name=target,proto3" json:"target,omitempty"`
	Method string `protobuf:"bytes,2,opt,name=method,proto3" json:"method,omitempty"`
	Key    string `protobuf:"bytes,3,opt,name=key,proto3" json:"key,omitempty"`
	Field  string `protobuf:"bytes,5,opt,name=field,proto3" json:"field,omitempty"`
	Value  string `protobuf:"bytes,6,opt,name=value,proto3" json:"value,omitempty"`
	Slice  string `protobuf:"bytes,7,opt,name=slice,proto3" json:"slice,omitempty"`
}

func (m *MzsetOptions) Reset()                    { *m = MzsetOptions{} }
func (m *MzsetOptions) String() string            { return proto.CompactTextString(m) }
func (*MzsetOptions) ProtoMessage()               {}
func (*MzsetOptions) Descriptor() ([]byte, []int) { return fileDescriptorRorm, []int{1} }

func (m *MzsetOptions) GetTarget() string {
	if m != nil {
		return m.Target
	}
	return ""
}

func (m *MzsetOptions) GetMethod() string {
	if m != nil {
		return m.Method
	}
	return ""
}

func (m *MzsetOptions) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *MzsetOptions) GetField() string {
	if m != nil {
		return m.Field
	}
	return ""
}

func (m *MzsetOptions) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

func (m *MzsetOptions) GetSlice() string {
	if m != nil {
		return m.Slice
	}
	return ""
}

type TranOptions struct {
	Target string `protobuf:"bytes,1,opt,name=target,proto3" json:"target,omitempty"`
	Method string `protobuf:"bytes,2,opt,name=method,proto3" json:"method,omitempty"`
	Param  string `protobuf:"bytes,3,opt,name=param,proto3" json:"param,omitempty"`
	Slice  string `protobuf:"bytes,4,opt,name=slice,proto3" json:"slice,omitempty"`
}

func (m *TranOptions) Reset()                    { *m = TranOptions{} }
func (m *TranOptions) String() string            { return proto.CompactTextString(m) }
func (*TranOptions) ProtoMessage()               {}
func (*TranOptions) Descriptor() ([]byte, []int) { return fileDescriptorRorm, []int{2} }

func (m *TranOptions) GetTarget() string {
	if m != nil {
		return m.Target
	}
	return ""
}

func (m *TranOptions) GetMethod() string {
	if m != nil {
		return m.Method
	}
	return ""
}

func (m *TranOptions) GetParam() string {
	if m != nil {
		return m.Param
	}
	return ""
}

func (m *TranOptions) GetSlice() string {
	if m != nil {
		return m.Slice
	}
	return ""
}

type UidOptions struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Seed string `protobuf:"bytes,2,opt,name=seed,proto3" json:"seed,omitempty"`
}

func (m *UidOptions) Reset()                    { *m = UidOptions{} }
func (m *UidOptions) String() string            { return proto.CompactTextString(m) }
func (*UidOptions) ProtoMessage()               {}
func (*UidOptions) Descriptor() ([]byte, []int) { return fileDescriptorRorm, []int{3} }

func (m *UidOptions) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *UidOptions) GetSeed() string {
	if m != nil {
		return m.Seed
	}
	return ""
}

type Variable struct {
	VarName string `protobuf:"bytes,1,opt,name=varName,proto3" json:"varName,omitempty"`
	VarType string `protobuf:"bytes,2,opt,name=varType,proto3" json:"varType,omitempty"`
}

func (m *Variable) Reset()                    { *m = Variable{} }
func (m *Variable) String() string            { return proto.CompactTextString(m) }
func (*Variable) ProtoMessage()               {}
func (*Variable) Descriptor() ([]byte, []int) { return fileDescriptorRorm, []int{4} }

func (m *Variable) GetVarName() string {
	if m != nil {
		return m.VarName
	}
	return ""
}

func (m *Variable) GetVarType() string {
	if m != nil {
		return m.VarType
	}
	return ""
}

var E_SqlType = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.ServiceOptions)(nil),
	ExtensionType: (*int64)(nil),
	Field:         44401,
	Name:          "rorm.sql_type",
	Tag:           "varint,44401,opt,name=sql_type,json=sqlType",
	Filename:      "options/rorm.proto",
}

var E_RedisType = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.ServiceOptions)(nil),
	ExtensionType: (*int64)(nil),
	Field:         44402,
	Name:          "rorm.redis_type",
	Tag:           "varint,44402,opt,name=redis_type,json=redisType",
	Filename:      "options/rorm.proto",
}

var E_UseUid = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.ServiceOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         44403,
	Name:          "rorm.use_uid",
	Tag:           "varint,44403,opt,name=use_uid,json=useUid",
	Filename:      "options/rorm.proto",
}

var E_UseNsq = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.ServiceOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         44404,
	Name:          "rorm.use_nsq",
	Tag:           "varint,44404,opt,name=use_nsq,json=useNsq",
	Filename:      "options/rorm.proto",
}

var E_GinHandler = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.ServiceOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         44406,
	Name:          "rorm.gin_handler",
	Tag:           "varint,44406,opt,name=gin_handler,json=ginHandler",
	Filename:      "options/rorm.proto",
}

var E_GrpcApiImp = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.ServiceOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         44407,
	Name:          "rorm.grpc_api_imp",
	Tag:           "varint,44407,opt,name=grpc_api_imp,json=grpcApiImp",
	Filename:      "options/rorm.proto",
}

var E_Opts = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MethodOptions)(nil),
	ExtensionType: (*RormOptions)(nil),
	Field:         44401,
	Name:          "rorm.opts",
	Tag:           "bytes,44401,opt,name=opts",
	Filename:      "options/rorm.proto",
}

var E_Uid = &proto.ExtensionDesc{
	ExtendedType:  (*google_protobuf.MethodOptions)(nil),
	ExtensionType: (*UidOptions)(nil),
	Field:         44402,
	Name:          "rorm.uid",
	Tag:           "bytes,44402,opt,name=uid",
	Filename:      "options/rorm.proto",
}

func init() {
	proto.RegisterType((*RormOptions)(nil), "rorm.RormOptions")
	proto.RegisterType((*MzsetOptions)(nil), "rorm.MzsetOptions")
	proto.RegisterType((*TranOptions)(nil), "rorm.TranOptions")
	proto.RegisterType((*UidOptions)(nil), "rorm.UidOptions")
	proto.RegisterType((*Variable)(nil), "rorm.Variable")
	proto.RegisterExtension(E_SqlType)
	proto.RegisterExtension(E_RedisType)
	proto.RegisterExtension(E_UseUid)
	proto.RegisterExtension(E_UseNsq)
	proto.RegisterExtension(E_GinHandler)
	proto.RegisterExtension(E_GrpcApiImp)
	proto.RegisterExtension(E_Opts)
	proto.RegisterExtension(E_Uid)
}

func init() { proto.RegisterFile("options/rorm.proto", fileDescriptorRorm) }

var fileDescriptorRorm = []byte{
	// 532 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x94, 0xcf, 0x6e, 0xd3, 0x40,
	0x10, 0xc6, 0x95, 0xe6, 0x8f, 0xe3, 0x49, 0x0f, 0x65, 0x85, 0x90, 0xc5, 0x01, 0xa2, 0x9c, 0x22,
	0x21, 0x39, 0x52, 0xe1, 0x64, 0x21, 0x04, 0x05, 0x09, 0x38, 0xb4, 0x48, 0xa6, 0xe5, 0xc0, 0xc5,
	0xda, 0xd8, 0x13, 0x77, 0x85, 0xed, 0xdd, 0xec, 0xda, 0x11, 0xe1, 0x05, 0xb8, 0xf3, 0x78, 0xbc,
	0x05, 0x05, 0xc1, 0x15, 0xed, 0x1f, 0xb7, 0x41, 0x80, 0x0c, 0x52, 0x6f, 0x3b, 0x5f, 0xbe, 0xdf,
	0x37, 0xbb, 0x93, 0x49, 0x80, 0x70, 0x51, 0x33, 0x5e, 0xa9, 0x85, 0xe4, 0xb2, 0x0c, 0x85, 0xe4,
	0x35, 0x27, 0x03, 0x7d, 0xbe, 0x3d, 0xcd, 0x39, 0xcf, 0x0b, 0x5c, 0x18, 0x6d, 0xd9, 0xac, 0x16,
	0x19, 0xaa, 0x54, 0x32, 0x51, 0x73, 0x69, 0x7d, 0xb3, 0x8f, 0x7b, 0x30, 0x89, 0xb9, 0x2c, 0x5f,
	0xd9, 0x08, 0x72, 0x0b, 0x46, 0x35, 0x95, 0x39, 0xd6, 0x41, 0x6f, 0xda, 0x9b, 0xfb, 0xb1, 0xab,
	0xb4, 0x5e, 0x62, 0x7d, 0xce, 0xb3, 0x60, 0xcf, 0xea, 0xb6, 0x22, 0x37, 0x61, 0x28, 0xa8, 0xa4,
	0x65, 0xd0, 0x37, 0xb2, 0x2d, 0x48, 0x08, 0xbe, 0x5a, 0x17, 0xef, 0x93, 0x5a, 0xd2, 0x2a, 0x18,
	0x4c, 0xfb, 0xf3, 0xc9, 0xe1, 0x8d, 0xd0, 0xdc, 0xee, 0x54, 0xd2, 0xca, 0xf5, 0x8a, 0xc7, 0xda,
	0xa3, 0x05, 0x32, 0x87, 0x61, 0xf9, 0x41, 0x61, 0x1d, 0x0c, 0xa7, 0xbd, 0xf9, 0xe4, 0x90, 0x58,
	0xef, 0xb1, 0x96, 0x5a, 0xb3, 0x35, 0x90, 0x7b, 0xe0, 0xa9, 0x26, 0x4d, 0x51, 0xa9, 0xc0, 0x33,
	0x5e, 0x97, 0xbb, 0xf3, 0x86, 0xb8, 0x75, 0x68, 0xf3, 0x8a, 0xb2, 0xa2, 0x91, 0x18, 0x8c, 0xff,
	0x6a, 0x76, 0x8e, 0xd9, 0xa7, 0x1e, 0xec, 0xef, 0x76, 0xfc, 0xef, 0x51, 0x1c, 0x40, 0xff, 0x1d,
	0x6e, 0xdd, 0x20, 0xf4, 0x51, 0x0f, 0x67, 0xc5, 0xb0, 0xc8, 0xcc, 0xb3, 0xfc, 0xd8, 0x16, 0x5a,
	0xdd, 0xd0, 0xa2, 0xc1, 0x60, 0x64, 0x55, 0x53, 0x68, 0x55, 0x15, 0x2c, 0x45, 0xf3, 0x2c, 0x3f,
	0xb6, 0xc5, 0x8c, 0xc1, 0x64, 0x67, 0x62, 0xd7, 0xf4, 0xed, 0x5c, 0xb6, 0x1a, 0xec, 0xb6, 0x7a,
	0x00, 0x70, 0xc6, 0xb2, 0xb6, 0x13, 0x81, 0x41, 0x45, 0x4b, 0x74, 0x7d, 0xcc, 0x59, 0x6b, 0x0a,
	0xb1, 0xed, 0x61, 0xce, 0xb3, 0x47, 0x30, 0x7e, 0x43, 0x25, 0xa3, 0xcb, 0x02, 0x49, 0x00, 0xde,
	0x86, 0xca, 0x93, 0x2b, 0xac, 0x2d, 0xdd, 0x27, 0xa7, 0x5b, 0x81, 0x0e, 0x6e, 0xcb, 0xe8, 0x21,
	0xe8, 0x2d, 0x48, 0xea, 0xad, 0x40, 0x72, 0x37, 0xb4, 0xeb, 0x1a, 0xb6, 0xeb, 0x1a, 0xbe, 0x46,
	0xb9, 0x61, 0x29, 0xba, 0x4b, 0x05, 0x5f, 0x3e, 0x6b, 0xba, 0x1f, 0x7b, 0x6a, 0x5d, 0x18, 0xfa,
	0x31, 0x80, 0xc4, 0x8c, 0xa9, 0x7f, 0xe4, 0x2f, 0x1c, 0xef, 0x1b, 0xc8, 0x24, 0x44, 0xe0, 0x35,
	0x0a, 0x93, 0x86, 0x65, 0xdd, 0xf8, 0x57, 0x83, 0x8f, 0xe3, 0x51, 0xa3, 0xf0, 0x8c, 0x65, 0x2d,
	0x5b, 0xa9, 0x75, 0x37, 0xfb, 0x6d, 0x87, 0x3d, 0x51, 0xeb, 0xe8, 0x08, 0x26, 0x39, 0xab, 0x92,
	0x73, 0x5a, 0x65, 0x05, 0xca, 0x6e, 0xfe, 0xbb, 0xe3, 0x21, 0x67, 0xd5, 0x0b, 0x0b, 0x45, 0x4f,
	0x61, 0x3f, 0x97, 0x22, 0x4d, 0xa8, 0x60, 0x09, 0x2b, 0x45, 0x77, 0xc8, 0x8f, 0xcb, 0x10, 0x29,
	0xd2, 0x27, 0x82, 0xbd, 0x2c, 0x45, 0xf4, 0x1c, 0x06, 0x5c, 0xd4, 0x8a, 0xdc, 0xf9, 0x0d, 0x3e,
	0x36, 0x3b, 0xf4, 0xeb, 0xec, 0xff, 0xf8, 0x13, 0x32, 0x01, 0xd1, 0x33, 0xe8, 0xeb, 0x29, 0x76,
	0xe5, 0x5c, 0xb8, 0x9c, 0x03, 0x9b, 0x73, 0xb5, 0x72, 0xb1, 0xc6, 0x8f, 0xfc, 0xb7, 0x9e, 0xfb,
	0x37, 0x5b, 0x8e, 0x4c, 0xc2, 0xfd, 0x9f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x88, 0x0b, 0x18, 0xb6,
	0xdf, 0x04, 0x00, 0x00,
}