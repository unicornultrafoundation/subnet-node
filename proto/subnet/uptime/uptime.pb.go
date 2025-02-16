// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/subnet/uptime/uptime.proto

package uptime

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Heartbeat defines the structure of the message sent via PubSub
type HeartbeatMsg struct {
	ProviderId int64 `protobuf:"varint,1,opt,name=provider_id,json=providerId,proto3" json:"provider_id,omitempty"`
	Timestamp  int64 `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (m *HeartbeatMsg) Reset()         { *m = HeartbeatMsg{} }
func (m *HeartbeatMsg) String() string { return proto.CompactTextString(m) }
func (*HeartbeatMsg) ProtoMessage()    {}
func (*HeartbeatMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd24e57098828256, []int{0}
}
func (m *HeartbeatMsg) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *HeartbeatMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_HeartbeatMsg.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *HeartbeatMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HeartbeatMsg.Merge(m, src)
}
func (m *HeartbeatMsg) XXX_Size() int {
	return m.Size()
}
func (m *HeartbeatMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_HeartbeatMsg.DiscardUnknown(m)
}

var xxx_messageInfo_HeartbeatMsg proto.InternalMessageInfo

func (m *HeartbeatMsg) GetProviderId() int64 {
	if m != nil {
		return m.ProviderId
	}
	return 0
}

func (m *HeartbeatMsg) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

// Proof represents an individual uptime proof
type Proof struct {
	ProviderId int64    `protobuf:"varint,1,opt,name=provider_id,json=providerId,proto3" json:"provider_id,omitempty"`
	Uptime     int64    `protobuf:"varint,2,opt,name=uptime,proto3" json:"uptime,omitempty"`
	Proof      []string `protobuf:"bytes,3,rep,name=proof,proto3" json:"proof,omitempty"`
}

func (m *Proof) Reset()         { *m = Proof{} }
func (m *Proof) String() string { return proto.CompactTextString(m) }
func (*Proof) ProtoMessage()    {}
func (*Proof) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd24e57098828256, []int{1}
}
func (m *Proof) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Proof) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Proof.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Proof) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Proof.Merge(m, src)
}
func (m *Proof) XXX_Size() int {
	return m.Size()
}
func (m *Proof) XXX_DiscardUnknown() {
	xxx_messageInfo_Proof.DiscardUnknown(m)
}

var xxx_messageInfo_Proof proto.InternalMessageInfo

func (m *Proof) GetProviderId() int64 {
	if m != nil {
		return m.ProviderId
	}
	return 0
}

func (m *Proof) GetUptime() int64 {
	if m != nil {
		return m.Uptime
	}
	return 0
}

func (m *Proof) GetProof() []string {
	if m != nil {
		return m.Proof
	}
	return nil
}

// MerkleProofs contains a collection of uptime proofs with a root hash
type MerkleProofMsg struct {
	Root   string   `protobuf:"bytes,1,opt,name=root,proto3" json:"root,omitempty"`
	Proofs []*Proof `protobuf:"bytes,2,rep,name=proofs,proto3" json:"proofs,omitempty"`
}

func (m *MerkleProofMsg) Reset()         { *m = MerkleProofMsg{} }
func (m *MerkleProofMsg) String() string { return proto.CompactTextString(m) }
func (*MerkleProofMsg) ProtoMessage()    {}
func (*MerkleProofMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd24e57098828256, []int{2}
}
func (m *MerkleProofMsg) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MerkleProofMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MerkleProofMsg.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MerkleProofMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MerkleProofMsg.Merge(m, src)
}
func (m *MerkleProofMsg) XXX_Size() int {
	return m.Size()
}
func (m *MerkleProofMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_MerkleProofMsg.DiscardUnknown(m)
}

var xxx_messageInfo_MerkleProofMsg proto.InternalMessageInfo

func (m *MerkleProofMsg) GetRoot() string {
	if m != nil {
		return m.Root
	}
	return ""
}

func (m *MerkleProofMsg) GetProofs() []*Proof {
	if m != nil {
		return m.Proofs
	}
	return nil
}

// UptimeProof represents the Merkle proof for a specific uptime record
type UptimeProof struct {
	Uptime int64    `protobuf:"varint,1,opt,name=uptime,proto3" json:"uptime,omitempty"`
	Proof  []string `protobuf:"bytes,2,rep,name=proof,proto3" json:"proof,omitempty"`
}

func (m *UptimeProof) Reset()         { *m = UptimeProof{} }
func (m *UptimeProof) String() string { return proto.CompactTextString(m) }
func (*UptimeProof) ProtoMessage()    {}
func (*UptimeProof) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd24e57098828256, []int{3}
}
func (m *UptimeProof) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *UptimeProof) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_UptimeProof.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *UptimeProof) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UptimeProof.Merge(m, src)
}
func (m *UptimeProof) XXX_Size() int {
	return m.Size()
}
func (m *UptimeProof) XXX_DiscardUnknown() {
	xxx_messageInfo_UptimeProof.DiscardUnknown(m)
}

var xxx_messageInfo_UptimeProof proto.InternalMessageInfo

func (m *UptimeProof) GetUptime() int64 {
	if m != nil {
		return m.Uptime
	}
	return 0
}

func (m *UptimeProof) GetProof() []string {
	if m != nil {
		return m.Proof
	}
	return nil
}

// UptimeRecord represents a peer's uptime data and proof
type UptimeRecord struct {
	ProviderId    int64        `protobuf:"varint,1,opt,name=provider_id,json=providerId,proto3" json:"provider_id,omitempty"`
	Uptime        int64        `protobuf:"varint,3,opt,name=uptime,proto3" json:"uptime,omitempty"`
	LastTimestamp int64        `protobuf:"varint,4,opt,name=last_timestamp,json=lastTimestamp,proto3" json:"last_timestamp,omitempty"`
	Proof         *UptimeProof `protobuf:"bytes,5,opt,name=proof,proto3" json:"proof,omitempty"`
	IsClaimed     bool         `protobuf:"varint,6,opt,name=is_claimed,json=isClaimed,proto3" json:"is_claimed,omitempty"`
}

func (m *UptimeRecord) Reset()         { *m = UptimeRecord{} }
func (m *UptimeRecord) String() string { return proto.CompactTextString(m) }
func (*UptimeRecord) ProtoMessage()    {}
func (*UptimeRecord) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd24e57098828256, []int{4}
}
func (m *UptimeRecord) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *UptimeRecord) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_UptimeRecord.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *UptimeRecord) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UptimeRecord.Merge(m, src)
}
func (m *UptimeRecord) XXX_Size() int {
	return m.Size()
}
func (m *UptimeRecord) XXX_DiscardUnknown() {
	xxx_messageInfo_UptimeRecord.DiscardUnknown(m)
}

var xxx_messageInfo_UptimeRecord proto.InternalMessageInfo

func (m *UptimeRecord) GetProviderId() int64 {
	if m != nil {
		return m.ProviderId
	}
	return 0
}

func (m *UptimeRecord) GetUptime() int64 {
	if m != nil {
		return m.Uptime
	}
	return 0
}

func (m *UptimeRecord) GetLastTimestamp() int64 {
	if m != nil {
		return m.LastTimestamp
	}
	return 0
}

func (m *UptimeRecord) GetProof() *UptimeProof {
	if m != nil {
		return m.Proof
	}
	return nil
}

func (m *UptimeRecord) GetIsClaimed() bool {
	if m != nil {
		return m.IsClaimed
	}
	return false
}

// Msg represents the data used for publishing uptime info
type Msg struct {
	// Types that are valid to be assigned to Payload:
	//
	//	*Msg_Heartbeat
	//	*Msg_MerkleProof
	Payload isMsg_Payload `protobuf_oneof:"payload"`
}

func (m *Msg) Reset()         { *m = Msg{} }
func (m *Msg) String() string { return proto.CompactTextString(m) }
func (*Msg) ProtoMessage()    {}
func (*Msg) Descriptor() ([]byte, []int) {
	return fileDescriptor_cd24e57098828256, []int{5}
}
func (m *Msg) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Msg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Msg.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Msg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Msg.Merge(m, src)
}
func (m *Msg) XXX_Size() int {
	return m.Size()
}
func (m *Msg) XXX_DiscardUnknown() {
	xxx_messageInfo_Msg.DiscardUnknown(m)
}

var xxx_messageInfo_Msg proto.InternalMessageInfo

type isMsg_Payload interface {
	isMsg_Payload()
	MarshalTo([]byte) (int, error)
	Size() int
}

type Msg_Heartbeat struct {
	Heartbeat *HeartbeatMsg `protobuf:"bytes,1,opt,name=heartbeat,proto3,oneof" json:"heartbeat,omitempty"`
}
type Msg_MerkleProof struct {
	MerkleProof *MerkleProofMsg `protobuf:"bytes,2,opt,name=merkleProof,proto3,oneof" json:"merkleProof,omitempty"`
}

func (*Msg_Heartbeat) isMsg_Payload()   {}
func (*Msg_MerkleProof) isMsg_Payload() {}

func (m *Msg) GetPayload() isMsg_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (m *Msg) GetHeartbeat() *HeartbeatMsg {
	if x, ok := m.GetPayload().(*Msg_Heartbeat); ok {
		return x.Heartbeat
	}
	return nil
}

func (m *Msg) GetMerkleProof() *MerkleProofMsg {
	if x, ok := m.GetPayload().(*Msg_MerkleProof); ok {
		return x.MerkleProof
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Msg) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Msg_Heartbeat)(nil),
		(*Msg_MerkleProof)(nil),
	}
}

func init() {
	proto.RegisterType((*HeartbeatMsg)(nil), "uptime.HeartbeatMsg")
	proto.RegisterType((*Proof)(nil), "uptime.Proof")
	proto.RegisterType((*MerkleProofMsg)(nil), "uptime.MerkleProofMsg")
	proto.RegisterType((*UptimeProof)(nil), "uptime.UptimeProof")
	proto.RegisterType((*UptimeRecord)(nil), "uptime.UptimeRecord")
	proto.RegisterType((*Msg)(nil), "uptime.Msg")
}

func init() { proto.RegisterFile("proto/subnet/uptime/uptime.proto", fileDescriptor_cd24e57098828256) }

var fileDescriptor_cd24e57098828256 = []byte{
	// 384 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0x41, 0x4b, 0xeb, 0x40,
	0x10, 0xc7, 0xb3, 0x4d, 0x9b, 0xf7, 0x32, 0x69, 0x7b, 0xd8, 0x57, 0x4a, 0x0e, 0xef, 0xe5, 0x85,
	0x40, 0x21, 0x5e, 0x5a, 0x88, 0x9e, 0xf4, 0x56, 0x2f, 0x15, 0x29, 0xc8, 0xa2, 0x5e, 0x4b, 0xda,
	0xac, 0x1a, 0x4c, 0xdc, 0xb0, 0xd9, 0x0a, 0x5e, 0xfd, 0x04, 0x7e, 0x1c, 0x3f, 0x82, 0xc7, 0x1e,
	0x3d, 0x4a, 0xfb, 0x45, 0x24, 0x9b, 0x2c, 0x49, 0x41, 0x10, 0x4f, 0xc9, 0xfc, 0x67, 0xe6, 0xcf,
	0x6f, 0x76, 0x06, 0xdc, 0x8c, 0x33, 0xc1, 0x26, 0xf9, 0x7a, 0xf9, 0x40, 0xc5, 0x64, 0x9d, 0x89,
	0x38, 0xa5, 0xd5, 0x67, 0x2c, 0x53, 0xd8, 0x28, 0x23, 0x6f, 0x0e, 0xdd, 0x19, 0x0d, 0xb9, 0x58,
	0xd2, 0x50, 0xcc, 0xf3, 0x5b, 0xfc, 0x1f, 0xac, 0x8c, 0xb3, 0xc7, 0x38, 0xa2, 0x7c, 0x11, 0x47,
	0x36, 0x72, 0x91, 0xaf, 0x13, 0x50, 0xd2, 0x59, 0x84, 0xff, 0x82, 0x59, 0x34, 0xe6, 0x22, 0x4c,
	0x33, 0xbb, 0x25, 0xd3, 0xb5, 0xe0, 0x5d, 0x43, 0xe7, 0x82, 0x33, 0x76, 0xf3, 0xbd, 0xcf, 0x10,
	0x2a, 0x84, 0xca, 0xa4, 0x8a, 0xf0, 0x00, 0x3a, 0x59, 0xe1, 0x60, 0xeb, 0xae, 0xee, 0x9b, 0xa4,
	0x0c, 0xbc, 0x73, 0xe8, 0xcf, 0x29, 0xbf, 0x4f, 0xa8, 0x74, 0x2f, 0x40, 0x31, 0xb4, 0x39, 0x63,
	0x42, 0x3a, 0x9b, 0x44, 0xfe, 0xe3, 0x11, 0x18, 0xb2, 0x3c, 0xb7, 0x5b, 0xae, 0xee, 0x5b, 0x41,
	0x6f, 0x5c, 0xcd, 0x2c, 0xbb, 0x48, 0x95, 0xf4, 0x4e, 0xc0, 0xba, 0x92, 0x7a, 0x89, 0x5a, 0x93,
	0xa0, 0xaf, 0x49, 0x5a, 0x4d, 0x92, 0x57, 0x04, 0xdd, 0xb2, 0x9b, 0xd0, 0x15, 0xe3, 0xd1, 0x4f,
	0x26, 0xd5, 0xf7, 0xfc, 0x47, 0xd0, 0x4f, 0xc2, 0x5c, 0x2c, 0xea, 0xe7, 0x6c, 0xcb, 0x7c, 0xaf,
	0x50, 0x2f, 0x95, 0x88, 0x0f, 0x14, 0x46, 0xc7, 0x45, 0xbe, 0x15, 0xfc, 0x51, 0x33, 0x35, 0x46,
	0xa8, 0xd8, 0xf0, 0x3f, 0x80, 0x38, 0x5f, 0xac, 0x92, 0x30, 0x4e, 0x69, 0x64, 0x1b, 0x2e, 0xf2,
	0x7f, 0x13, 0x33, 0xce, 0x4f, 0x4b, 0xc1, 0x7b, 0x46, 0xa0, 0x17, 0x4f, 0x77, 0x04, 0xe6, 0x9d,
	0xda, 0xb9, 0xe4, 0xb5, 0x82, 0x81, 0x72, 0x6d, 0x1e, 0xc3, 0x4c, 0x23, 0x75, 0x21, 0x3e, 0x06,
	0x2b, 0xad, 0x57, 0x20, 0xb7, 0x66, 0x05, 0x43, 0xd5, 0xb7, 0xbf, 0x9d, 0x99, 0x46, 0x9a, 0xc5,
	0x53, 0x13, 0x7e, 0x65, 0xe1, 0x53, 0xc2, 0xc2, 0x68, 0x6a, 0xbf, 0x6d, 0x1d, 0xb4, 0xd9, 0x3a,
	0xe8, 0x63, 0xeb, 0xa0, 0x97, 0x9d, 0xa3, 0x6d, 0x76, 0x8e, 0xf6, 0xbe, 0x73, 0xb4, 0xa5, 0x21,
	0x2f, 0xf3, 0xf0, 0x33, 0x00, 0x00, 0xff, 0xff, 0xf6, 0x9d, 0x12, 0x2a, 0xbd, 0x02, 0x00, 0x00,
}

func (m *HeartbeatMsg) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *HeartbeatMsg) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *HeartbeatMsg) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Timestamp != 0 {
		i = encodeVarintUptime(dAtA, i, uint64(m.Timestamp))
		i--
		dAtA[i] = 0x10
	}
	if m.ProviderId != 0 {
		i = encodeVarintUptime(dAtA, i, uint64(m.ProviderId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *Proof) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Proof) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Proof) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Proof) > 0 {
		for iNdEx := len(m.Proof) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Proof[iNdEx])
			copy(dAtA[i:], m.Proof[iNdEx])
			i = encodeVarintUptime(dAtA, i, uint64(len(m.Proof[iNdEx])))
			i--
			dAtA[i] = 0x1a
		}
	}
	if m.Uptime != 0 {
		i = encodeVarintUptime(dAtA, i, uint64(m.Uptime))
		i--
		dAtA[i] = 0x10
	}
	if m.ProviderId != 0 {
		i = encodeVarintUptime(dAtA, i, uint64(m.ProviderId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *MerkleProofMsg) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MerkleProofMsg) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MerkleProofMsg) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Proofs) > 0 {
		for iNdEx := len(m.Proofs) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Proofs[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintUptime(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.Root) > 0 {
		i -= len(m.Root)
		copy(dAtA[i:], m.Root)
		i = encodeVarintUptime(dAtA, i, uint64(len(m.Root)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *UptimeProof) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *UptimeProof) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *UptimeProof) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Proof) > 0 {
		for iNdEx := len(m.Proof) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Proof[iNdEx])
			copy(dAtA[i:], m.Proof[iNdEx])
			i = encodeVarintUptime(dAtA, i, uint64(len(m.Proof[iNdEx])))
			i--
			dAtA[i] = 0x12
		}
	}
	if m.Uptime != 0 {
		i = encodeVarintUptime(dAtA, i, uint64(m.Uptime))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *UptimeRecord) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *UptimeRecord) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *UptimeRecord) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.IsClaimed {
		i--
		if m.IsClaimed {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x30
	}
	if m.Proof != nil {
		{
			size, err := m.Proof.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintUptime(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	if m.LastTimestamp != 0 {
		i = encodeVarintUptime(dAtA, i, uint64(m.LastTimestamp))
		i--
		dAtA[i] = 0x20
	}
	if m.Uptime != 0 {
		i = encodeVarintUptime(dAtA, i, uint64(m.Uptime))
		i--
		dAtA[i] = 0x18
	}
	if m.ProviderId != 0 {
		i = encodeVarintUptime(dAtA, i, uint64(m.ProviderId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *Msg) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Msg) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Msg) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Payload != nil {
		{
			size := m.Payload.Size()
			i -= size
			if _, err := m.Payload.MarshalTo(dAtA[i:]); err != nil {
				return 0, err
			}
		}
	}
	return len(dAtA) - i, nil
}

func (m *Msg_Heartbeat) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Msg_Heartbeat) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.Heartbeat != nil {
		{
			size, err := m.Heartbeat.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintUptime(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}
func (m *Msg_MerkleProof) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Msg_MerkleProof) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.MerkleProof != nil {
		{
			size, err := m.MerkleProof.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintUptime(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	return len(dAtA) - i, nil
}
func encodeVarintUptime(dAtA []byte, offset int, v uint64) int {
	offset -= sovUptime(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *HeartbeatMsg) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ProviderId != 0 {
		n += 1 + sovUptime(uint64(m.ProviderId))
	}
	if m.Timestamp != 0 {
		n += 1 + sovUptime(uint64(m.Timestamp))
	}
	return n
}

func (m *Proof) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ProviderId != 0 {
		n += 1 + sovUptime(uint64(m.ProviderId))
	}
	if m.Uptime != 0 {
		n += 1 + sovUptime(uint64(m.Uptime))
	}
	if len(m.Proof) > 0 {
		for _, s := range m.Proof {
			l = len(s)
			n += 1 + l + sovUptime(uint64(l))
		}
	}
	return n
}

func (m *MerkleProofMsg) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Root)
	if l > 0 {
		n += 1 + l + sovUptime(uint64(l))
	}
	if len(m.Proofs) > 0 {
		for _, e := range m.Proofs {
			l = e.Size()
			n += 1 + l + sovUptime(uint64(l))
		}
	}
	return n
}

func (m *UptimeProof) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Uptime != 0 {
		n += 1 + sovUptime(uint64(m.Uptime))
	}
	if len(m.Proof) > 0 {
		for _, s := range m.Proof {
			l = len(s)
			n += 1 + l + sovUptime(uint64(l))
		}
	}
	return n
}

func (m *UptimeRecord) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ProviderId != 0 {
		n += 1 + sovUptime(uint64(m.ProviderId))
	}
	if m.Uptime != 0 {
		n += 1 + sovUptime(uint64(m.Uptime))
	}
	if m.LastTimestamp != 0 {
		n += 1 + sovUptime(uint64(m.LastTimestamp))
	}
	if m.Proof != nil {
		l = m.Proof.Size()
		n += 1 + l + sovUptime(uint64(l))
	}
	if m.IsClaimed {
		n += 2
	}
	return n
}

func (m *Msg) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Payload != nil {
		n += m.Payload.Size()
	}
	return n
}

func (m *Msg_Heartbeat) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Heartbeat != nil {
		l = m.Heartbeat.Size()
		n += 1 + l + sovUptime(uint64(l))
	}
	return n
}
func (m *Msg_MerkleProof) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.MerkleProof != nil {
		l = m.MerkleProof.Size()
		n += 1 + l + sovUptime(uint64(l))
	}
	return n
}

func sovUptime(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozUptime(x uint64) (n int) {
	return sovUptime(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *HeartbeatMsg) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowUptime
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: HeartbeatMsg: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: HeartbeatMsg: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProviderId", wireType)
			}
			m.ProviderId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ProviderId |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
			}
			m.Timestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Timestamp |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipUptime(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthUptime
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Proof) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowUptime
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Proof: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Proof: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProviderId", wireType)
			}
			m.ProviderId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ProviderId |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Uptime", wireType)
			}
			m.Uptime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Uptime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Proof", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthUptime
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthUptime
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Proof = append(m.Proof, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipUptime(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthUptime
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MerkleProofMsg) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowUptime
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MerkleProofMsg: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MerkleProofMsg: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Root", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthUptime
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthUptime
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Root = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Proofs", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthUptime
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthUptime
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Proofs = append(m.Proofs, &Proof{})
			if err := m.Proofs[len(m.Proofs)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipUptime(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthUptime
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *UptimeProof) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowUptime
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: UptimeProof: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: UptimeProof: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Uptime", wireType)
			}
			m.Uptime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Uptime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Proof", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthUptime
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthUptime
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Proof = append(m.Proof, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipUptime(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthUptime
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *UptimeRecord) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowUptime
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: UptimeRecord: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: UptimeRecord: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProviderId", wireType)
			}
			m.ProviderId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ProviderId |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Uptime", wireType)
			}
			m.Uptime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Uptime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastTimestamp", wireType)
			}
			m.LastTimestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LastTimestamp |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Proof", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthUptime
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthUptime
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Proof == nil {
				m.Proof = &UptimeProof{}
			}
			if err := m.Proof.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IsClaimed", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.IsClaimed = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipUptime(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthUptime
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Msg) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowUptime
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Msg: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Msg: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Heartbeat", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthUptime
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthUptime
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &HeartbeatMsg{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Payload = &Msg_Heartbeat{v}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MerkleProof", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthUptime
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthUptime
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &MerkleProofMsg{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Payload = &Msg_MerkleProof{v}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipUptime(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthUptime
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipUptime(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowUptime
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowUptime
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthUptime
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupUptime
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthUptime
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthUptime        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowUptime          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupUptime = fmt.Errorf("proto: unexpected end of group")
)
