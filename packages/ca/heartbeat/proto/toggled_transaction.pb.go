// Code generated by protoc-gen-go. DO NOT EDIT.
// source: toggled_transaction.proto

package proto

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ToggledTransaction struct {
	TransactionId        []byte     `protobuf:"bytes,1,opt,name=transactionId,proto3" json:"transactionId,omitempty"`
	ToggleReason         ToggleType `protobuf:"varint,2,opt,name=toggleReason,proto3,enum=proto.ToggleType" json:"toggleReason,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *ToggledTransaction) Reset()         { *m = ToggledTransaction{} }
func (m *ToggledTransaction) String() string { return proto.CompactTextString(m) }
func (*ToggledTransaction) ProtoMessage()    {}
func (*ToggledTransaction) Descriptor() ([]byte, []int) {
	return fileDescriptor_ddb0b50608ac9f9d, []int{0}
}

func (m *ToggledTransaction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ToggledTransaction.Unmarshal(m, b)
}
func (m *ToggledTransaction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ToggledTransaction.Marshal(b, m, deterministic)
}
func (m *ToggledTransaction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ToggledTransaction.Merge(m, src)
}
func (m *ToggledTransaction) XXX_Size() int {
	return xxx_messageInfo_ToggledTransaction.Size(m)
}
func (m *ToggledTransaction) XXX_DiscardUnknown() {
	xxx_messageInfo_ToggledTransaction.DiscardUnknown(m)
}

var xxx_messageInfo_ToggledTransaction proto.InternalMessageInfo

func (m *ToggledTransaction) GetTransactionId() []byte {
	if m != nil {
		return m.TransactionId
	}
	return nil
}

func (m *ToggledTransaction) GetToggleReason() ToggleType {
	if m != nil {
		return m.ToggleReason
	}
	return ToggleType_Like
}

func init() {
	proto.RegisterType((*ToggledTransaction)(nil), "proto.ToggledTransaction")
}

func init() { proto.RegisterFile("toggled_transaction.proto", fileDescriptor_ddb0b50608ac9f9d) }

var fileDescriptor_ddb0b50608ac9f9d = []byte{
	// 130 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2c, 0xc9, 0x4f, 0x4f,
	0xcf, 0x49, 0x4d, 0x89, 0x2f, 0x29, 0x4a, 0xcc, 0x2b, 0x4e, 0x4c, 0x2e, 0xc9, 0xcc, 0xcf, 0xd3,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x05, 0x53, 0x52, 0x82, 0x10, 0x15, 0xf1, 0x25, 0x95,
	0x05, 0xa9, 0x10, 0x19, 0xa5, 0x42, 0x2e, 0xa1, 0x10, 0x88, 0xb6, 0x10, 0x84, 0x2e, 0x21, 0x15,
	0x2e, 0x5e, 0x24, 0x43, 0x3c, 0x53, 0x24, 0x18, 0x15, 0x18, 0x35, 0x78, 0x82, 0x50, 0x05, 0x85,
	0x4c, 0xb9, 0x78, 0x20, 0x06, 0x06, 0xa5, 0x26, 0x16, 0xe7, 0xe7, 0x49, 0x30, 0x29, 0x30, 0x6a,
	0xf0, 0x19, 0x09, 0x42, 0x4c, 0xd6, 0x83, 0x18, 0x1b, 0x52, 0x59, 0x90, 0x1a, 0x84, 0xa2, 0x2c,
	0x89, 0x0d, 0x2c, 0x6f, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0xed, 0xb6, 0x2b, 0x6a, 0xb0, 0x00,
	0x00, 0x00,
}
