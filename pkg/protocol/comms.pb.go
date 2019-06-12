// Code generated by protoc-gen-go. DO NOT EDIT.
// source: comms.proto

package protocol

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
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

type Category int32

const (
	Category_UNKNOWN       Category = 0
	Category_POSITION      Category = 1
	Category_PROFILE       Category = 2
	Category_CHAT          Category = 3
	Category_SCENE_MESSAGE Category = 4
)

var Category_name = map[int32]string{
	0: "UNKNOWN",
	1: "POSITION",
	2: "PROFILE",
	3: "CHAT",
	4: "SCENE_MESSAGE",
}

var Category_value = map[string]int32{
	"UNKNOWN":       0,
	"POSITION":      1,
	"PROFILE":       2,
	"CHAT":          3,
	"SCENE_MESSAGE": 4,
}

func (x Category) String() string {
	return proto.EnumName(Category_name, int32(x))
}

func (Category) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_db39efb7717b7d47, []int{0}
}

type AuthData struct {
	Signature            string   `protobuf:"bytes,1,opt,name=signature,proto3" json:"signature,omitempty"`
	Identity             string   `protobuf:"bytes,2,opt,name=identity,proto3" json:"identity,omitempty"`
	Timestamp            string   `protobuf:"bytes,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	AccessToken          string   `protobuf:"bytes,4,opt,name=access_token,json=accessToken,proto3" json:"access_token,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AuthData) Reset()         { *m = AuthData{} }
func (m *AuthData) String() string { return proto.CompactTextString(m) }
func (*AuthData) ProtoMessage()    {}
func (*AuthData) Descriptor() ([]byte, []int) {
	return fileDescriptor_db39efb7717b7d47, []int{0}
}

func (m *AuthData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AuthData.Unmarshal(m, b)
}
func (m *AuthData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AuthData.Marshal(b, m, deterministic)
}
func (m *AuthData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AuthData.Merge(m, src)
}
func (m *AuthData) XXX_Size() int {
	return xxx_messageInfo_AuthData.Size(m)
}
func (m *AuthData) XXX_DiscardUnknown() {
	xxx_messageInfo_AuthData.DiscardUnknown(m)
}

var xxx_messageInfo_AuthData proto.InternalMessageInfo

func (m *AuthData) GetSignature() string {
	if m != nil {
		return m.Signature
	}
	return ""
}

func (m *AuthData) GetIdentity() string {
	if m != nil {
		return m.Identity
	}
	return ""
}

func (m *AuthData) GetTimestamp() string {
	if m != nil {
		return m.Timestamp
	}
	return ""
}

func (m *AuthData) GetAccessToken() string {
	if m != nil {
		return m.AccessToken
	}
	return ""
}

type DataHeader struct {
	Category             Category `protobuf:"varint,1,opt,name=category,proto3,enum=protocol.Category" json:"category,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DataHeader) Reset()         { *m = DataHeader{} }
func (m *DataHeader) String() string { return proto.CompactTextString(m) }
func (*DataHeader) ProtoMessage()    {}
func (*DataHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_db39efb7717b7d47, []int{1}
}

func (m *DataHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DataHeader.Unmarshal(m, b)
}
func (m *DataHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DataHeader.Marshal(b, m, deterministic)
}
func (m *DataHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DataHeader.Merge(m, src)
}
func (m *DataHeader) XXX_Size() int {
	return xxx_messageInfo_DataHeader.Size(m)
}
func (m *DataHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_DataHeader.DiscardUnknown(m)
}

var xxx_messageInfo_DataHeader proto.InternalMessageInfo

func (m *DataHeader) GetCategory() Category {
	if m != nil {
		return m.Category
	}
	return Category_UNKNOWN
}

type PositionData struct {
	Category             Category `protobuf:"varint,1,opt,name=category,proto3,enum=protocol.Category" json:"category,omitempty"`
	Time                 float64  `protobuf:"fixed64,2,opt,name=time,proto3" json:"time,omitempty"`
	PositionX            float32  `protobuf:"fixed32,3,opt,name=position_x,json=positionX,proto3" json:"position_x,omitempty"`
	PositionY            float32  `protobuf:"fixed32,4,opt,name=position_y,json=positionY,proto3" json:"position_y,omitempty"`
	PositionZ            float32  `protobuf:"fixed32,5,opt,name=position_z,json=positionZ,proto3" json:"position_z,omitempty"`
	RotationX            float32  `protobuf:"fixed32,6,opt,name=rotation_x,json=rotationX,proto3" json:"rotation_x,omitempty"`
	RotationY            float32  `protobuf:"fixed32,7,opt,name=rotation_y,json=rotationY,proto3" json:"rotation_y,omitempty"`
	RotationZ            float32  `protobuf:"fixed32,8,opt,name=rotation_z,json=rotationZ,proto3" json:"rotation_z,omitempty"`
	RotationW            float32  `protobuf:"fixed32,9,opt,name=rotation_w,json=rotationW,proto3" json:"rotation_w,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PositionData) Reset()         { *m = PositionData{} }
func (m *PositionData) String() string { return proto.CompactTextString(m) }
func (*PositionData) ProtoMessage()    {}
func (*PositionData) Descriptor() ([]byte, []int) {
	return fileDescriptor_db39efb7717b7d47, []int{2}
}

func (m *PositionData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PositionData.Unmarshal(m, b)
}
func (m *PositionData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PositionData.Marshal(b, m, deterministic)
}
func (m *PositionData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PositionData.Merge(m, src)
}
func (m *PositionData) XXX_Size() int {
	return xxx_messageInfo_PositionData.Size(m)
}
func (m *PositionData) XXX_DiscardUnknown() {
	xxx_messageInfo_PositionData.DiscardUnknown(m)
}

var xxx_messageInfo_PositionData proto.InternalMessageInfo

func (m *PositionData) GetCategory() Category {
	if m != nil {
		return m.Category
	}
	return Category_UNKNOWN
}

func (m *PositionData) GetTime() float64 {
	if m != nil {
		return m.Time
	}
	return 0
}

func (m *PositionData) GetPositionX() float32 {
	if m != nil {
		return m.PositionX
	}
	return 0
}

func (m *PositionData) GetPositionY() float32 {
	if m != nil {
		return m.PositionY
	}
	return 0
}

func (m *PositionData) GetPositionZ() float32 {
	if m != nil {
		return m.PositionZ
	}
	return 0
}

func (m *PositionData) GetRotationX() float32 {
	if m != nil {
		return m.RotationX
	}
	return 0
}

func (m *PositionData) GetRotationY() float32 {
	if m != nil {
		return m.RotationY
	}
	return 0
}

func (m *PositionData) GetRotationZ() float32 {
	if m != nil {
		return m.RotationZ
	}
	return 0
}

func (m *PositionData) GetRotationW() float32 {
	if m != nil {
		return m.RotationW
	}
	return 0
}

type ProfileData struct {
	Category             Category `protobuf:"varint,1,opt,name=category,proto3,enum=protocol.Category" json:"category,omitempty"`
	Time                 float64  `protobuf:"fixed64,2,opt,name=time,proto3" json:"time,omitempty"`
	AvatarType           string   `protobuf:"bytes,3,opt,name=avatar_type,json=avatarType,proto3" json:"avatar_type,omitempty"`
	DisplayName          string   `protobuf:"bytes,4,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	PublicKey            string   `protobuf:"bytes,5,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProfileData) Reset()         { *m = ProfileData{} }
func (m *ProfileData) String() string { return proto.CompactTextString(m) }
func (*ProfileData) ProtoMessage()    {}
func (*ProfileData) Descriptor() ([]byte, []int) {
	return fileDescriptor_db39efb7717b7d47, []int{3}
}

func (m *ProfileData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProfileData.Unmarshal(m, b)
}
func (m *ProfileData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProfileData.Marshal(b, m, deterministic)
}
func (m *ProfileData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProfileData.Merge(m, src)
}
func (m *ProfileData) XXX_Size() int {
	return xxx_messageInfo_ProfileData.Size(m)
}
func (m *ProfileData) XXX_DiscardUnknown() {
	xxx_messageInfo_ProfileData.DiscardUnknown(m)
}

var xxx_messageInfo_ProfileData proto.InternalMessageInfo

func (m *ProfileData) GetCategory() Category {
	if m != nil {
		return m.Category
	}
	return Category_UNKNOWN
}

func (m *ProfileData) GetTime() float64 {
	if m != nil {
		return m.Time
	}
	return 0
}

func (m *ProfileData) GetAvatarType() string {
	if m != nil {
		return m.AvatarType
	}
	return ""
}

func (m *ProfileData) GetDisplayName() string {
	if m != nil {
		return m.DisplayName
	}
	return ""
}

func (m *ProfileData) GetPublicKey() string {
	if m != nil {
		return m.PublicKey
	}
	return ""
}

type ChatData struct {
	Category             Category `protobuf:"varint,1,opt,name=category,proto3,enum=protocol.Category" json:"category,omitempty"`
	Time                 float64  `protobuf:"fixed64,2,opt,name=time,proto3" json:"time,omitempty"`
	MessageId            string   `protobuf:"bytes,3,opt,name=message_id,json=messageId,proto3" json:"message_id,omitempty"`
	Text                 string   `protobuf:"bytes,4,opt,name=text,proto3" json:"text,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ChatData) Reset()         { *m = ChatData{} }
func (m *ChatData) String() string { return proto.CompactTextString(m) }
func (*ChatData) ProtoMessage()    {}
func (*ChatData) Descriptor() ([]byte, []int) {
	return fileDescriptor_db39efb7717b7d47, []int{4}
}

func (m *ChatData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ChatData.Unmarshal(m, b)
}
func (m *ChatData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ChatData.Marshal(b, m, deterministic)
}
func (m *ChatData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ChatData.Merge(m, src)
}
func (m *ChatData) XXX_Size() int {
	return xxx_messageInfo_ChatData.Size(m)
}
func (m *ChatData) XXX_DiscardUnknown() {
	xxx_messageInfo_ChatData.DiscardUnknown(m)
}

var xxx_messageInfo_ChatData proto.InternalMessageInfo

func (m *ChatData) GetCategory() Category {
	if m != nil {
		return m.Category
	}
	return Category_UNKNOWN
}

func (m *ChatData) GetTime() float64 {
	if m != nil {
		return m.Time
	}
	return 0
}

func (m *ChatData) GetMessageId() string {
	if m != nil {
		return m.MessageId
	}
	return ""
}

func (m *ChatData) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

func init() {
	proto.RegisterEnum("protocol.Category", Category_name, Category_value)
	proto.RegisterType((*AuthData)(nil), "protocol.AuthData")
	proto.RegisterType((*DataHeader)(nil), "protocol.DataHeader")
	proto.RegisterType((*PositionData)(nil), "protocol.PositionData")
	proto.RegisterType((*ProfileData)(nil), "protocol.ProfileData")
	proto.RegisterType((*ChatData)(nil), "protocol.ChatData")
}

func init() { proto.RegisterFile("comms.proto", fileDescriptor_db39efb7717b7d47) }

var fileDescriptor_db39efb7717b7d47 = []byte{
	// 442 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x92, 0xc1, 0x6f, 0xd3, 0x30,
	0x14, 0xc6, 0x49, 0x57, 0x36, 0xe7, 0xa5, 0xa0, 0xe0, 0x53, 0x84, 0x40, 0x40, 0x4e, 0x88, 0x43,
	0x0f, 0x70, 0xe5, 0x52, 0x95, 0xc0, 0xaa, 0x41, 0x52, 0xa5, 0x45, 0xdb, 0xb8, 0x44, 0x5e, 0xf2,
	0xe8, 0xac, 0x35, 0x71, 0x14, 0xbb, 0x30, 0xef, 0xc6, 0x81, 0xbf, 0x86, 0x2b, 0x7f, 0x20, 0x8a,
	0x93, 0x14, 0x7c, 0x44, 0xea, 0x29, 0xce, 0xf7, 0xf3, 0xfb, 0xf4, 0xbd, 0xe7, 0x07, 0x5e, 0x2e,
	0xca, 0x52, 0x4e, 0xeb, 0x46, 0x28, 0x41, 0x89, 0xf9, 0xe4, 0x62, 0x1b, 0xfe, 0x74, 0x80, 0xcc,
	0x76, 0xea, 0xfa, 0x1d, 0x53, 0x8c, 0x3e, 0x01, 0x57, 0xf2, 0x4d, 0xc5, 0xd4, 0xae, 0xc1, 0xc0,
	0x79, 0xee, 0xbc, 0x74, 0xd3, 0xbf, 0x02, 0x7d, 0x0c, 0x84, 0x17, 0x58, 0x29, 0xae, 0x74, 0x30,
	0x32, 0x70, 0xff, 0xdf, 0x56, 0x2a, 0x5e, 0xa2, 0x54, 0xac, 0xac, 0x83, 0xa3, 0xae, 0x72, 0x2f,
	0xd0, 0x17, 0x30, 0x61, 0x79, 0x8e, 0x52, 0x66, 0x4a, 0xdc, 0x60, 0x15, 0x8c, 0xcd, 0x05, 0xaf,
	0xd3, 0xd6, 0xad, 0x14, 0xbe, 0x05, 0x68, 0x23, 0x9c, 0x22, 0x2b, 0xb0, 0xa1, 0x53, 0x20, 0x39,
	0x53, 0xb8, 0x11, 0x8d, 0x36, 0x39, 0x1e, 0xbe, 0xa6, 0xd3, 0x21, 0xf2, 0x74, 0xde, 0x93, 0x74,
	0x7f, 0x27, 0xfc, 0x35, 0x82, 0xc9, 0x52, 0x48, 0xae, 0xb8, 0xa8, 0x4c, 0x27, 0xff, 0x69, 0x40,
	0x29, 0x8c, 0xdb, 0xb8, 0xa6, 0x2f, 0x27, 0x35, 0x67, 0xfa, 0x14, 0xa0, 0xee, 0x3d, 0xb3, 0x5b,
	0xd3, 0xd4, 0x28, 0x75, 0x07, 0xe5, 0xc2, 0xc2, 0xda, 0xb4, 0xf4, 0x0f, 0xbe, 0xb4, 0xf0, 0x5d,
	0x70, 0xdf, 0xc6, 0x5f, 0x5a, 0xdc, 0x08, 0xc5, 0x7a, 0xf3, 0xe3, 0x0e, 0x0f, 0xca, 0x85, 0x85,
	0x75, 0x70, 0x62, 0xe3, 0x4b, 0x0b, 0xdf, 0x05, 0xc4, 0xc6, 0xb6, 0xf9, 0xf7, 0xc0, 0xb5, 0xf1,
	0x79, 0xf8, 0xdb, 0x01, 0x6f, 0xd9, 0x88, 0xaf, 0x7c, 0x8b, 0x07, 0x1b, 0xd6, 0x33, 0xf0, 0xd8,
	0x37, 0xa6, 0x58, 0x93, 0x29, 0x5d, 0x63, 0xbf, 0x02, 0xd0, 0x49, 0x6b, 0x5d, 0x63, 0xbb, 0x03,
	0x05, 0x97, 0xf5, 0x96, 0xe9, 0xac, 0x62, 0x25, 0x0e, 0x3b, 0xd0, 0x6b, 0x31, 0xeb, 0x07, 0xbe,
	0xbb, 0xda, 0xf2, 0x3c, 0xbb, 0x41, 0x6d, 0x46, 0xe6, 0xa6, 0x6e, 0xa7, 0x9c, 0xa1, 0x0e, 0x7f,
	0x38, 0x40, 0xe6, 0xd7, 0x4c, 0x1d, 0xf2, 0x81, 0x4b, 0x94, 0x92, 0x6d, 0x30, 0xe3, 0xc5, 0xb0,
	0xb5, 0xbd, 0xb2, 0x28, 0x4c, 0x09, 0xde, 0xaa, 0x3e, 0xa9, 0x39, 0xbf, 0x4a, 0x80, 0x0c, 0xe6,
	0xd4, 0x83, 0x93, 0xcf, 0xf1, 0x59, 0x9c, 0x9c, 0xc7, 0xfe, 0x3d, 0x3a, 0x01, 0xb2, 0x4c, 0x56,
	0x8b, 0xf5, 0x22, 0x89, 0x7d, 0xa7, 0x45, 0xcb, 0x34, 0x79, 0xbf, 0xf8, 0x18, 0xf9, 0x23, 0x4a,
	0x60, 0x3c, 0x3f, 0x9d, 0xad, 0xfd, 0x23, 0xfa, 0x08, 0x1e, 0xac, 0xe6, 0x51, 0x1c, 0x65, 0x9f,
	0xa2, 0xd5, 0x6a, 0xf6, 0x21, 0xf2, 0xc7, 0x57, 0xc7, 0x26, 0xf4, 0x9b, 0x3f, 0x01, 0x00, 0x00,
	0xff, 0xff, 0xb8, 0xef, 0xdc, 0x5b, 0x9f, 0x03, 0x00, 0x00,
}
