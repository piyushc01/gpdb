// Code generated by protoc-gen-go. DO NOT EDIT.
// source: agent.proto

package idl

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type GetHostNameReply struct {
	Hostname             string   `protobuf:"bytes,1,opt,name=hostname,proto3" json:"hostname,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetHostNameReply) Reset()         { *m = GetHostNameReply{} }
func (m *GetHostNameReply) String() string { return proto.CompactTextString(m) }
func (*GetHostNameReply) ProtoMessage()    {}
func (*GetHostNameReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_56ede974c0020f77, []int{0}
}

func (m *GetHostNameReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetHostNameReply.Unmarshal(m, b)
}
func (m *GetHostNameReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetHostNameReply.Marshal(b, m, deterministic)
}
func (m *GetHostNameReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetHostNameReply.Merge(m, src)
}
func (m *GetHostNameReply) XXX_Size() int {
	return xxx_messageInfo_GetHostNameReply.Size(m)
}
func (m *GetHostNameReply) XXX_DiscardUnknown() {
	xxx_messageInfo_GetHostNameReply.DiscardUnknown(m)
}

var xxx_messageInfo_GetHostNameReply proto.InternalMessageInfo

func (m *GetHostNameReply) GetHostname() string {
	if m != nil {
		return m.Hostname
	}
	return ""
}

type GetHostNameRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetHostNameRequest) Reset()         { *m = GetHostNameRequest{} }
func (m *GetHostNameRequest) String() string { return proto.CompactTextString(m) }
func (*GetHostNameRequest) ProtoMessage()    {}
func (*GetHostNameRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_56ede974c0020f77, []int{1}
}

func (m *GetHostNameRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetHostNameRequest.Unmarshal(m, b)
}
func (m *GetHostNameRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetHostNameRequest.Marshal(b, m, deterministic)
}
func (m *GetHostNameRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetHostNameRequest.Merge(m, src)
}
func (m *GetHostNameRequest) XXX_Size() int {
	return xxx_messageInfo_GetHostNameRequest.Size(m)
}
func (m *GetHostNameRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetHostNameRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetHostNameRequest proto.InternalMessageInfo

type StartSegmentRequest struct {
	DataDir              string   `protobuf:"bytes,1,opt,name=dataDir,proto3" json:"dataDir,omitempty"`
	Wait                 bool     `protobuf:"varint,2,opt,name=wait,proto3" json:"wait,omitempty"`
	Timeout              int32    `protobuf:"varint,4,opt,name=timeout,proto3" json:"timeout,omitempty"`
	Options              string   `protobuf:"bytes,5,opt,name=options,proto3" json:"options,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StartSegmentRequest) Reset()         { *m = StartSegmentRequest{} }
func (m *StartSegmentRequest) String() string { return proto.CompactTextString(m) }
func (*StartSegmentRequest) ProtoMessage()    {}
func (*StartSegmentRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_56ede974c0020f77, []int{2}
}

func (m *StartSegmentRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StartSegmentRequest.Unmarshal(m, b)
}
func (m *StartSegmentRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StartSegmentRequest.Marshal(b, m, deterministic)
}
func (m *StartSegmentRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StartSegmentRequest.Merge(m, src)
}
func (m *StartSegmentRequest) XXX_Size() int {
	return xxx_messageInfo_StartSegmentRequest.Size(m)
}
func (m *StartSegmentRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_StartSegmentRequest.DiscardUnknown(m)
}

var xxx_messageInfo_StartSegmentRequest proto.InternalMessageInfo

func (m *StartSegmentRequest) GetDataDir() string {
	if m != nil {
		return m.DataDir
	}
	return ""
}

func (m *StartSegmentRequest) GetWait() bool {
	if m != nil {
		return m.Wait
	}
	return false
}

func (m *StartSegmentRequest) GetTimeout() int32 {
	if m != nil {
		return m.Timeout
	}
	return 0
}

func (m *StartSegmentRequest) GetOptions() string {
	if m != nil {
		return m.Options
	}
	return ""
}

type StartSegmentReply struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StartSegmentReply) Reset()         { *m = StartSegmentReply{} }
func (m *StartSegmentReply) String() string { return proto.CompactTextString(m) }
func (*StartSegmentReply) ProtoMessage()    {}
func (*StartSegmentReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_56ede974c0020f77, []int{3}
}

func (m *StartSegmentReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StartSegmentReply.Unmarshal(m, b)
}
func (m *StartSegmentReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StartSegmentReply.Marshal(b, m, deterministic)
}
func (m *StartSegmentReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StartSegmentReply.Merge(m, src)
}
func (m *StartSegmentReply) XXX_Size() int {
	return xxx_messageInfo_StartSegmentReply.Size(m)
}
func (m *StartSegmentReply) XXX_DiscardUnknown() {
	xxx_messageInfo_StartSegmentReply.DiscardUnknown(m)
}

var xxx_messageInfo_StartSegmentReply proto.InternalMessageInfo

type StopAgentRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StopAgentRequest) Reset()         { *m = StopAgentRequest{} }
func (m *StopAgentRequest) String() string { return proto.CompactTextString(m) }
func (*StopAgentRequest) ProtoMessage()    {}
func (*StopAgentRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_56ede974c0020f77, []int{4}
}

func (m *StopAgentRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StopAgentRequest.Unmarshal(m, b)
}
func (m *StopAgentRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StopAgentRequest.Marshal(b, m, deterministic)
}
func (m *StopAgentRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StopAgentRequest.Merge(m, src)
}
func (m *StopAgentRequest) XXX_Size() int {
	return xxx_messageInfo_StopAgentRequest.Size(m)
}
func (m *StopAgentRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_StopAgentRequest.DiscardUnknown(m)
}

var xxx_messageInfo_StopAgentRequest proto.InternalMessageInfo

type StopAgentReply struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StopAgentReply) Reset()         { *m = StopAgentReply{} }
func (m *StopAgentReply) String() string { return proto.CompactTextString(m) }
func (*StopAgentReply) ProtoMessage()    {}
func (*StopAgentReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_56ede974c0020f77, []int{5}
}

func (m *StopAgentReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StopAgentReply.Unmarshal(m, b)
}
func (m *StopAgentReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StopAgentReply.Marshal(b, m, deterministic)
}
func (m *StopAgentReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StopAgentReply.Merge(m, src)
}
func (m *StopAgentReply) XXX_Size() int {
	return xxx_messageInfo_StopAgentReply.Size(m)
}
func (m *StopAgentReply) XXX_DiscardUnknown() {
	xxx_messageInfo_StopAgentReply.DiscardUnknown(m)
}

var xxx_messageInfo_StopAgentReply proto.InternalMessageInfo

type StatusAgentRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StatusAgentRequest) Reset()         { *m = StatusAgentRequest{} }
func (m *StatusAgentRequest) String() string { return proto.CompactTextString(m) }
func (*StatusAgentRequest) ProtoMessage()    {}
func (*StatusAgentRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_56ede974c0020f77, []int{6}
}

func (m *StatusAgentRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StatusAgentRequest.Unmarshal(m, b)
}
func (m *StatusAgentRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StatusAgentRequest.Marshal(b, m, deterministic)
}
func (m *StatusAgentRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StatusAgentRequest.Merge(m, src)
}
func (m *StatusAgentRequest) XXX_Size() int {
	return xxx_messageInfo_StatusAgentRequest.Size(m)
}
func (m *StatusAgentRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_StatusAgentRequest.DiscardUnknown(m)
}

var xxx_messageInfo_StatusAgentRequest proto.InternalMessageInfo

type StatusAgentReply struct {
	Status               string   `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Uptime               string   `protobuf:"bytes,2,opt,name=uptime,proto3" json:"uptime,omitempty"`
	Pid                  uint32   `protobuf:"varint,3,opt,name=pid,proto3" json:"pid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StatusAgentReply) Reset()         { *m = StatusAgentReply{} }
func (m *StatusAgentReply) String() string { return proto.CompactTextString(m) }
func (*StatusAgentReply) ProtoMessage()    {}
func (*StatusAgentReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_56ede974c0020f77, []int{7}
}

func (m *StatusAgentReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StatusAgentReply.Unmarshal(m, b)
}
func (m *StatusAgentReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StatusAgentReply.Marshal(b, m, deterministic)
}
func (m *StatusAgentReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StatusAgentReply.Merge(m, src)
}
func (m *StatusAgentReply) XXX_Size() int {
	return xxx_messageInfo_StatusAgentReply.Size(m)
}
func (m *StatusAgentReply) XXX_DiscardUnknown() {
	xxx_messageInfo_StatusAgentReply.DiscardUnknown(m)
}

var xxx_messageInfo_StatusAgentReply proto.InternalMessageInfo

func (m *StatusAgentReply) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

func (m *StatusAgentReply) GetUptime() string {
	if m != nil {
		return m.Uptime
	}
	return ""
}

func (m *StatusAgentReply) GetPid() uint32 {
	if m != nil {
		return m.Pid
	}
	return 0
}

type ValidateHostEnvRequest struct {
	HostAddressList      []string `protobuf:"bytes,1,rep,name=hostAddressList,proto3" json:"hostAddressList,omitempty"`
	DirectoryList        []string `protobuf:"bytes,2,rep,name=DirectoryList,proto3" json:"DirectoryList,omitempty"`
	PortList             []string `protobuf:"bytes,3,rep,name=portList,proto3" json:"portList,omitempty"`
	Locale               *Locale  `protobuf:"bytes,4,opt,name=locale,proto3" json:"locale,omitempty"`
	GpVersion            string   `protobuf:"bytes,5,opt,name=gpVersion,proto3" json:"gpVersion,omitempty"`
	Forced               bool     `protobuf:"varint,6,opt,name=forced,proto3" json:"forced,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValidateHostEnvRequest) Reset()         { *m = ValidateHostEnvRequest{} }
func (m *ValidateHostEnvRequest) String() string { return proto.CompactTextString(m) }
func (*ValidateHostEnvRequest) ProtoMessage()    {}
func (*ValidateHostEnvRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_56ede974c0020f77, []int{8}
}

func (m *ValidateHostEnvRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ValidateHostEnvRequest.Unmarshal(m, b)
}
func (m *ValidateHostEnvRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ValidateHostEnvRequest.Marshal(b, m, deterministic)
}
func (m *ValidateHostEnvRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidateHostEnvRequest.Merge(m, src)
}
func (m *ValidateHostEnvRequest) XXX_Size() int {
	return xxx_messageInfo_ValidateHostEnvRequest.Size(m)
}
func (m *ValidateHostEnvRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidateHostEnvRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ValidateHostEnvRequest proto.InternalMessageInfo

func (m *ValidateHostEnvRequest) GetHostAddressList() []string {
	if m != nil {
		return m.HostAddressList
	}
	return nil
}

func (m *ValidateHostEnvRequest) GetDirectoryList() []string {
	if m != nil {
		return m.DirectoryList
	}
	return nil
}

func (m *ValidateHostEnvRequest) GetPortList() []string {
	if m != nil {
		return m.PortList
	}
	return nil
}

func (m *ValidateHostEnvRequest) GetLocale() *Locale {
	if m != nil {
		return m.Locale
	}
	return nil
}

func (m *ValidateHostEnvRequest) GetGpVersion() string {
	if m != nil {
		return m.GpVersion
	}
	return ""
}

func (m *ValidateHostEnvRequest) GetForced() bool {
	if m != nil {
		return m.Forced
	}
	return false
}

type ValidateHostEnvReply struct {
	Messages             []*LogMessage `protobuf:"bytes,1,rep,name=messages,proto3" json:"messages,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *ValidateHostEnvReply) Reset()         { *m = ValidateHostEnvReply{} }
func (m *ValidateHostEnvReply) String() string { return proto.CompactTextString(m) }
func (*ValidateHostEnvReply) ProtoMessage()    {}
func (*ValidateHostEnvReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_56ede974c0020f77, []int{9}
}

func (m *ValidateHostEnvReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ValidateHostEnvReply.Unmarshal(m, b)
}
func (m *ValidateHostEnvReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ValidateHostEnvReply.Marshal(b, m, deterministic)
}
func (m *ValidateHostEnvReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidateHostEnvReply.Merge(m, src)
}
func (m *ValidateHostEnvReply) XXX_Size() int {
	return xxx_messageInfo_ValidateHostEnvReply.Size(m)
}
func (m *ValidateHostEnvReply) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidateHostEnvReply.DiscardUnknown(m)
}

var xxx_messageInfo_ValidateHostEnvReply proto.InternalMessageInfo

func (m *ValidateHostEnvReply) GetMessages() []*LogMessage {
	if m != nil {
		return m.Messages
	}
	return nil
}

type MakeSegmentRequest struct {
	Segment              *Segment          `protobuf:"bytes,1,opt,name=segment,proto3" json:"segment,omitempty"`
	Locale               *Locale           `protobuf:"bytes,2,opt,name=locale,proto3" json:"locale,omitempty"`
	Encoding             string            `protobuf:"bytes,3,opt,name=Encoding,proto3" json:"Encoding,omitempty"`
	SegConfig            map[string]string `protobuf:"bytes,4,rep,name=SegConfig,proto3" json:"SegConfig,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	CoordinatorAddrs     []string          `protobuf:"bytes,5,rep,name=coordinatorAddrs,proto3" json:"coordinatorAddrs,omitempty"`
	HbaHostNames         bool              `protobuf:"varint,6,opt,name=hbaHostNames,proto3" json:"hbaHostNames,omitempty"`
	DataChecksums        bool              `protobuf:"varint,7,opt,name=dataChecksums,proto3" json:"dataChecksums,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *MakeSegmentRequest) Reset()         { *m = MakeSegmentRequest{} }
func (m *MakeSegmentRequest) String() string { return proto.CompactTextString(m) }
func (*MakeSegmentRequest) ProtoMessage()    {}
func (*MakeSegmentRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_56ede974c0020f77, []int{10}
}

func (m *MakeSegmentRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MakeSegmentRequest.Unmarshal(m, b)
}
func (m *MakeSegmentRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MakeSegmentRequest.Marshal(b, m, deterministic)
}
func (m *MakeSegmentRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MakeSegmentRequest.Merge(m, src)
}
func (m *MakeSegmentRequest) XXX_Size() int {
	return xxx_messageInfo_MakeSegmentRequest.Size(m)
}
func (m *MakeSegmentRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_MakeSegmentRequest.DiscardUnknown(m)
}

var xxx_messageInfo_MakeSegmentRequest proto.InternalMessageInfo

func (m *MakeSegmentRequest) GetSegment() *Segment {
	if m != nil {
		return m.Segment
	}
	return nil
}

func (m *MakeSegmentRequest) GetLocale() *Locale {
	if m != nil {
		return m.Locale
	}
	return nil
}

func (m *MakeSegmentRequest) GetEncoding() string {
	if m != nil {
		return m.Encoding
	}
	return ""
}

func (m *MakeSegmentRequest) GetSegConfig() map[string]string {
	if m != nil {
		return m.SegConfig
	}
	return nil
}

func (m *MakeSegmentRequest) GetCoordinatorAddrs() []string {
	if m != nil {
		return m.CoordinatorAddrs
	}
	return nil
}

func (m *MakeSegmentRequest) GetHbaHostNames() bool {
	if m != nil {
		return m.HbaHostNames
	}
	return false
}

func (m *MakeSegmentRequest) GetDataChecksums() bool {
	if m != nil {
		return m.DataChecksums
	}
	return false
}

type MakeSegmentReply struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MakeSegmentReply) Reset()         { *m = MakeSegmentReply{} }
func (m *MakeSegmentReply) String() string { return proto.CompactTextString(m) }
func (*MakeSegmentReply) ProtoMessage()    {}
func (*MakeSegmentReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_56ede974c0020f77, []int{11}
}

func (m *MakeSegmentReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MakeSegmentReply.Unmarshal(m, b)
}
func (m *MakeSegmentReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MakeSegmentReply.Marshal(b, m, deterministic)
}
func (m *MakeSegmentReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MakeSegmentReply.Merge(m, src)
}
func (m *MakeSegmentReply) XXX_Size() int {
	return xxx_messageInfo_MakeSegmentReply.Size(m)
}
func (m *MakeSegmentReply) XXX_DiscardUnknown() {
	xxx_messageInfo_MakeSegmentReply.DiscardUnknown(m)
}

var xxx_messageInfo_MakeSegmentReply proto.InternalMessageInfo

func init() {
	proto.RegisterType((*GetHostNameReply)(nil), "idl.GetHostNameReply")
	proto.RegisterType((*GetHostNameRequest)(nil), "idl.GetHostNameRequest")
	proto.RegisterType((*StartSegmentRequest)(nil), "idl.StartSegmentRequest")
	proto.RegisterType((*StartSegmentReply)(nil), "idl.StartSegmentReply")
	proto.RegisterType((*StopAgentRequest)(nil), "idl.StopAgentRequest")
	proto.RegisterType((*StopAgentReply)(nil), "idl.StopAgentReply")
	proto.RegisterType((*StatusAgentRequest)(nil), "idl.StatusAgentRequest")
	proto.RegisterType((*StatusAgentReply)(nil), "idl.StatusAgentReply")
	proto.RegisterType((*ValidateHostEnvRequest)(nil), "idl.ValidateHostEnvRequest")
	proto.RegisterType((*ValidateHostEnvReply)(nil), "idl.ValidateHostEnvReply")
	proto.RegisterType((*MakeSegmentRequest)(nil), "idl.MakeSegmentRequest")
	proto.RegisterMapType((map[string]string)(nil), "idl.MakeSegmentRequest.SegConfigEntry")
	proto.RegisterType((*MakeSegmentReply)(nil), "idl.MakeSegmentReply")
}

func init() { proto.RegisterFile("agent.proto", fileDescriptor_56ede974c0020f77) }

var fileDescriptor_56ede974c0020f77 = []byte{
	// 678 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x54, 0x4d, 0x6f, 0xd3, 0x4c,
	0x10, 0xae, 0xf3, 0xd5, 0x78, 0xd2, 0x8f, 0xbc, 0xd3, 0xb4, 0xaf, 0x31, 0x1c, 0x22, 0x83, 0xaa,
	0x08, 0x24, 0x23, 0x05, 0x0e, 0xa8, 0x42, 0x42, 0xfd, 0x12, 0x48, 0xb4, 0x1c, 0x1c, 0xd4, 0x03,
	0xb7, 0x6d, 0xbc, 0x75, 0x56, 0x75, 0xbc, 0xc6, 0xbb, 0x6e, 0x95, 0xff, 0xc0, 0x7f, 0xe4, 0x67,
	0x70, 0x45, 0xbb, 0x5e, 0xa7, 0xb1, 0x13, 0x6e, 0x9e, 0x67, 0x1e, 0xcf, 0xce, 0x33, 0x5f, 0xd0,
	0x23, 0x11, 0x4d, 0xa4, 0x9f, 0x66, 0x5c, 0x72, 0x6c, 0xb2, 0x30, 0x76, 0xed, 0x59, 0x7e, 0x5b,
	0xd8, 0x9e, 0x0f, 0xfd, 0xcf, 0x54, 0x7e, 0xe1, 0x42, 0x7e, 0x23, 0x73, 0x1a, 0xd0, 0x34, 0x5e,
	0xa0, 0x0b, 0xdd, 0x19, 0x17, 0x32, 0x21, 0x73, 0xea, 0x58, 0x43, 0x6b, 0x64, 0x07, 0x4b, 0xdb,
	0x1b, 0x00, 0x56, 0xf8, 0x3f, 0x73, 0x2a, 0xa4, 0xf7, 0x08, 0x07, 0x13, 0x49, 0x32, 0x39, 0xa1,
	0xd1, 0x9c, 0x26, 0xd2, 0xc0, 0xe8, 0xc0, 0x76, 0x48, 0x24, 0xb9, 0x60, 0x99, 0x89, 0x53, 0x9a,
	0x88, 0xd0, 0x7a, 0x24, 0x4c, 0x3a, 0x8d, 0xa1, 0x35, 0xea, 0x06, 0xfa, 0x5b, 0xb1, 0x25, 0x9b,
	0x53, 0x9e, 0x4b, 0xa7, 0x35, 0xb4, 0x46, 0xed, 0xa0, 0x34, 0x95, 0x87, 0xa7, 0x92, 0xf1, 0x44,
	0x38, 0xed, 0x22, 0x8e, 0x31, 0xbd, 0x03, 0xf8, 0xaf, 0xfa, 0x70, 0x1a, 0x2f, 0x3c, 0x84, 0xfe,
	0x44, 0xf2, 0xf4, 0x34, 0x7a, 0x4a, 0xc5, 0xeb, 0xc3, 0xde, 0x0a, 0xa6, 0x58, 0x03, 0xc0, 0x89,
	0x24, 0x32, 0x17, 0x15, 0xde, 0x77, 0xf5, 0xef, 0x0a, 0xaa, 0xea, 0x71, 0x04, 0x1d, 0xa1, 0x31,
	0xa3, 0xc2, 0x58, 0x0a, 0xcf, 0x53, 0x95, 0xa3, 0x96, 0x61, 0x07, 0xc6, 0xc2, 0x3e, 0x34, 0x53,
	0x16, 0x3a, 0xcd, 0xa1, 0x35, 0xda, 0x0d, 0xd4, 0xa7, 0xf7, 0xdb, 0x82, 0xa3, 0x1b, 0x12, 0xb3,
	0x90, 0x48, 0xaa, 0x6a, 0x77, 0x99, 0x3c, 0x94, 0x35, 0x1a, 0xc1, 0xbe, 0x2a, 0xee, 0x69, 0x18,
	0x66, 0x54, 0x88, 0x2b, 0x26, 0xa4, 0x63, 0x0d, 0x9b, 0x23, 0x3b, 0xa8, 0xc3, 0xf8, 0x0a, 0x76,
	0x2f, 0x58, 0x46, 0xa7, 0x92, 0x67, 0x0b, 0xcd, 0x6b, 0x68, 0x5e, 0x15, 0x54, 0xcd, 0x4b, 0x79,
	0x26, 0x35, 0xa1, 0xa9, 0x09, 0x4b, 0x1b, 0x5f, 0x42, 0x27, 0xe6, 0x53, 0x12, 0x53, 0x5d, 0xe0,
	0xde, 0xb8, 0xe7, 0xb3, 0x30, 0xf6, 0xaf, 0x34, 0x14, 0x18, 0x17, 0xbe, 0x00, 0x3b, 0x4a, 0x6f,
	0x68, 0x26, 0x18, 0x4f, 0x4c, 0xb9, 0x9f, 0x00, 0xa5, 0xf9, 0x8e, 0x67, 0x53, 0x1a, 0x3a, 0x1d,
	0xdd, 0x3a, 0x63, 0x79, 0xe7, 0x30, 0x58, 0x13, 0xa8, 0x6a, 0xf7, 0x06, 0xba, 0x73, 0x2a, 0x04,
	0x89, 0xa8, 0xd0, 0xba, 0x7a, 0xe3, 0x7d, 0xf3, 0x68, 0x74, 0x5d, 0xe0, 0xc1, 0x92, 0xe0, 0xfd,
	0x69, 0x00, 0x5e, 0x93, 0x7b, 0x5a, 0x1b, 0xa3, 0x63, 0xd8, 0x16, 0x05, 0xa2, 0x1b, 0xd0, 0x1b,
	0xef, 0xe8, 0x10, 0x25, 0xab, 0x74, 0xae, 0xc8, 0x6b, 0xfc, 0x5b, 0x9e, 0x0b, 0xdd, 0xcb, 0x64,
	0xca, 0x43, 0x96, 0x44, 0xba, 0x43, 0x76, 0xb0, 0xb4, 0xf1, 0x02, 0xec, 0x09, 0x8d, 0xce, 0x79,
	0x72, 0xc7, 0x22, 0xa7, 0xa5, 0xb3, 0x3d, 0xd6, 0x31, 0xd6, 0x93, 0xf2, 0x97, 0xc4, 0xcb, 0x44,
	0x66, 0x8b, 0xe0, 0xe9, 0x47, 0x7c, 0x0d, 0xfd, 0x29, 0xe7, 0x59, 0xc8, 0x12, 0x22, 0x79, 0xa6,
	0x3a, 0xa8, 0xc6, 0x56, 0x75, 0x62, 0x0d, 0x47, 0x0f, 0x76, 0x66, 0xb7, 0xa4, 0x5c, 0x27, 0x61,
	0x8a, 0x5a, 0xc1, 0x54, 0xdf, 0xd5, 0xda, 0x9c, 0xcf, 0xe8, 0xf4, 0x5e, 0xe4, 0x73, 0xe1, 0x6c,
	0x6b, 0x52, 0x15, 0x74, 0x3f, 0xc2, 0x5e, 0x35, 0x25, 0x35, 0x86, 0xf7, 0x74, 0x61, 0x66, 0x56,
	0x7d, 0xe2, 0x00, 0xda, 0x0f, 0x24, 0xce, 0xcb, 0x79, 0x2d, 0x8c, 0x93, 0xc6, 0x07, 0x4b, 0xad,
	0x4c, 0x45, 0x63, 0x1a, 0x2f, 0xc6, 0xbf, 0x9a, 0xd0, 0xd6, 0x5b, 0x80, 0xef, 0xa1, 0xa5, 0x96,
	0x07, 0x0f, 0x8b, 0xba, 0xd7, 0x76, 0xcb, 0x3d, 0xa8, 0xc3, 0x6a, 0xbd, 0xb6, 0xf0, 0x04, 0x3a,
	0xc5, 0x2a, 0xe1, 0xff, 0x86, 0x50, 0xdf, 0x36, 0xf7, 0x70, 0xdd, 0x51, 0xfc, 0xfb, 0x09, 0x7a,
	0x2b, 0xf9, 0x98, 0x00, 0xeb, 0x5d, 0x30, 0x01, 0xea, 0xa9, 0x7b, 0x5b, 0x78, 0x06, 0x3b, 0xab,
	0x87, 0x01, 0x9d, 0xf2, 0xa5, 0xfa, 0x91, 0x72, 0x8f, 0x36, 0x78, 0x8a, 0x18, 0x5f, 0x61, 0xbf,
	0x36, 0xd3, 0xf8, 0x5c, 0x93, 0x37, 0xaf, 0xb2, 0xfb, 0x6c, 0xb3, 0x73, 0xa9, 0x68, 0xe5, 0x70,
	0x1a, 0x45, 0xeb, 0xa7, 0xd4, 0x28, 0xaa, 0xdf, 0x64, 0x6f, 0xeb, 0xac, 0xfb, 0xa3, 0xe3, 0xfb,
	0x6f, 0x59, 0x18, 0xdf, 0x76, 0xf4, 0xe9, 0x7e, 0xf7, 0x37, 0x00, 0x00, 0xff, 0xff, 0xb6, 0xe0,
	0x8d, 0x56, 0xd9, 0x05, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// AgentClient is the client API for Agent service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AgentClient interface {
	Stop(ctx context.Context, in *StopAgentRequest, opts ...grpc.CallOption) (*StopAgentReply, error)
	Status(ctx context.Context, in *StatusAgentRequest, opts ...grpc.CallOption) (*StatusAgentReply, error)
	MakeSegment(ctx context.Context, in *MakeSegmentRequest, opts ...grpc.CallOption) (*MakeSegmentReply, error)
	StartSegment(ctx context.Context, in *StartSegmentRequest, opts ...grpc.CallOption) (*StartSegmentReply, error)
	ValidateHostEnv(ctx context.Context, in *ValidateHostEnvRequest, opts ...grpc.CallOption) (*ValidateHostEnvReply, error)
	GetHostName(ctx context.Context, in *GetHostNameRequest, opts ...grpc.CallOption) (*GetHostNameReply, error)
}

type agentClient struct {
	cc *grpc.ClientConn
}

func NewAgentClient(cc *grpc.ClientConn) AgentClient {
	return &agentClient{cc}
}

func (c *agentClient) Stop(ctx context.Context, in *StopAgentRequest, opts ...grpc.CallOption) (*StopAgentReply, error) {
	out := new(StopAgentReply)
	err := c.cc.Invoke(ctx, "/idl.Agent/Stop", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) Status(ctx context.Context, in *StatusAgentRequest, opts ...grpc.CallOption) (*StatusAgentReply, error) {
	out := new(StatusAgentReply)
	err := c.cc.Invoke(ctx, "/idl.Agent/Status", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) MakeSegment(ctx context.Context, in *MakeSegmentRequest, opts ...grpc.CallOption) (*MakeSegmentReply, error) {
	out := new(MakeSegmentReply)
	err := c.cc.Invoke(ctx, "/idl.Agent/MakeSegment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) StartSegment(ctx context.Context, in *StartSegmentRequest, opts ...grpc.CallOption) (*StartSegmentReply, error) {
	out := new(StartSegmentReply)
	err := c.cc.Invoke(ctx, "/idl.Agent/StartSegment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) ValidateHostEnv(ctx context.Context, in *ValidateHostEnvRequest, opts ...grpc.CallOption) (*ValidateHostEnvReply, error) {
	out := new(ValidateHostEnvReply)
	err := c.cc.Invoke(ctx, "/idl.Agent/ValidateHostEnv", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentClient) GetHostName(ctx context.Context, in *GetHostNameRequest, opts ...grpc.CallOption) (*GetHostNameReply, error) {
	out := new(GetHostNameReply)
	err := c.cc.Invoke(ctx, "/idl.Agent/GetHostName", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AgentServer is the server API for Agent service.
type AgentServer interface {
	Stop(context.Context, *StopAgentRequest) (*StopAgentReply, error)
	Status(context.Context, *StatusAgentRequest) (*StatusAgentReply, error)
	MakeSegment(context.Context, *MakeSegmentRequest) (*MakeSegmentReply, error)
	StartSegment(context.Context, *StartSegmentRequest) (*StartSegmentReply, error)
	ValidateHostEnv(context.Context, *ValidateHostEnvRequest) (*ValidateHostEnvReply, error)
	GetHostName(context.Context, *GetHostNameRequest) (*GetHostNameReply, error)
}

// UnimplementedAgentServer can be embedded to have forward compatible implementations.
type UnimplementedAgentServer struct {
}

func (*UnimplementedAgentServer) Stop(ctx context.Context, req *StopAgentRequest) (*StopAgentReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stop not implemented")
}
func (*UnimplementedAgentServer) Status(ctx context.Context, req *StatusAgentRequest) (*StatusAgentReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Status not implemented")
}
func (*UnimplementedAgentServer) MakeSegment(ctx context.Context, req *MakeSegmentRequest) (*MakeSegmentReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MakeSegment not implemented")
}
func (*UnimplementedAgentServer) StartSegment(ctx context.Context, req *StartSegmentRequest) (*StartSegmentReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartSegment not implemented")
}
func (*UnimplementedAgentServer) ValidateHostEnv(ctx context.Context, req *ValidateHostEnvRequest) (*ValidateHostEnvReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateHostEnv not implemented")
}
func (*UnimplementedAgentServer) GetHostName(ctx context.Context, req *GetHostNameRequest) (*GetHostNameReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetHostName not implemented")
}

func RegisterAgentServer(s *grpc.Server, srv AgentServer) {
	s.RegisterService(&_Agent_serviceDesc, srv)
}

func _Agent_Stop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopAgentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).Stop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/Stop",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).Stop(ctx, req.(*StopAgentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusAgentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).Status(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/Status",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).Status(ctx, req.(*StatusAgentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_MakeSegment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MakeSegmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).MakeSegment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/MakeSegment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).MakeSegment(ctx, req.(*MakeSegmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_StartSegment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartSegmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).StartSegment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/StartSegment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).StartSegment(ctx, req.(*StartSegmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_ValidateHostEnv_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateHostEnvRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).ValidateHostEnv(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/ValidateHostEnv",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).ValidateHostEnv(ctx, req.(*ValidateHostEnvRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Agent_GetHostName_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetHostNameRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServer).GetHostName(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/idl.Agent/GetHostName",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServer).GetHostName(ctx, req.(*GetHostNameRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Agent_serviceDesc = grpc.ServiceDesc{
	ServiceName: "idl.Agent",
	HandlerType: (*AgentServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Stop",
			Handler:    _Agent_Stop_Handler,
		},
		{
			MethodName: "Status",
			Handler:    _Agent_Status_Handler,
		},
		{
			MethodName: "MakeSegment",
			Handler:    _Agent_MakeSegment_Handler,
		},
		{
			MethodName: "StartSegment",
			Handler:    _Agent_StartSegment_Handler,
		},
		{
			MethodName: "ValidateHostEnv",
			Handler:    _Agent_ValidateHostEnv_Handler,
		},
		{
			MethodName: "GetHostName",
			Handler:    _Agent_GetHostName_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "agent.proto",
}
