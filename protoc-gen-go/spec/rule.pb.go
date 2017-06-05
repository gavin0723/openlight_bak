// Code generated by protoc-gen-go. DO NOT EDIT.
// source: spec/rule.proto

package spec

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// The parsed rule files
type RuleFiles struct {
	// Build rule file
	Build *BuildFile `protobuf:"bytes,1,opt,name=build" json:"build,omitempty"`
	// Run rule file
	Run *RunFile `protobuf:"bytes,2,opt,name=run" json:"run,omitempty"`
}

func (m *RuleFiles) Reset()                    { *m = RuleFiles{} }
func (m *RuleFiles) String() string            { return proto.CompactTextString(m) }
func (*RuleFiles) ProtoMessage()               {}
func (*RuleFiles) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

func (m *RuleFiles) GetBuild() *BuildFile {
	if m != nil {
		return m.Build
	}
	return nil
}

func (m *RuleFiles) GetRun() *RunFile {
	if m != nil {
		return m.Run
	}
	return nil
}

func init() {
	proto.RegisterType((*RuleFiles)(nil), "openlight.spec.RuleFiles")
}

func init() { proto.RegisterFile("spec/rule.proto", fileDescriptor2) }

var fileDescriptor2 = []byte{
	// 170 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2f, 0x2e, 0x48, 0x4d,
	0xd6, 0x2f, 0x2a, 0xcd, 0x49, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0xcb, 0x2f, 0x48,
	0xcd, 0xcb, 0xc9, 0x4c, 0xcf, 0x28, 0xd1, 0x03, 0x49, 0x49, 0x09, 0x80, 0x15, 0x24, 0x95, 0x66,
	0xe6, 0xa4, 0x40, 0x54, 0x48, 0x09, 0x42, 0xb5, 0xe4, 0xe5, 0xa5, 0x16, 0x41, 0x84, 0x94, 0xd2,
	0xb9, 0x38, 0x83, 0x4a, 0x73, 0x52, 0xdd, 0x32, 0x73, 0x52, 0x8b, 0x85, 0xf4, 0xb9, 0x58, 0xc1,
	0xca, 0x25, 0x18, 0x15, 0x18, 0x35, 0xb8, 0x8d, 0x24, 0xf5, 0x50, 0x4d, 0xd4, 0x73, 0x02, 0x49,
	0x82, 0x94, 0x06, 0x41, 0xd4, 0x09, 0x69, 0x72, 0x31, 0x17, 0x95, 0xe6, 0x49, 0x30, 0x81, 0x95,
	0x8b, 0xa3, 0x2b, 0x0f, 0x2a, 0xcd, 0x03, 0x2b, 0x06, 0xa9, 0x71, 0xd2, 0xe3, 0x92, 0x4c, 0xce,
	0xcf, 0xd5, 0x4b, 0xcf, 0x2c, 0xc9, 0x28, 0x4d, 0x42, 0x53, 0xe9, 0x04, 0x76, 0x43, 0x00, 0xc8,
	0x41, 0x01, 0x8c, 0x51, 0x2c, 0x20, 0xa1, 0x24, 0x36, 0xb0, 0xfb, 0x8c, 0x01, 0x01, 0x00, 0x00,
	0xff, 0xff, 0xaf, 0x25, 0xbd, 0x71, 0xe7, 0x00, 0x00, 0x00,
}
