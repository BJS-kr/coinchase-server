// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.2
// source: protodef/game_map.proto

package protodef

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Cell struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Occupied bool   `protobuf:"varint,1,opt,name=occupied,proto3" json:"occupied,omitempty"`
	Owner    string `protobuf:"bytes,2,opt,name=owner,proto3" json:"owner,omitempty"`
}

func (x *Cell) Reset() {
	*x = Cell{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protodef_game_map_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Cell) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Cell) ProtoMessage() {}

func (x *Cell) ProtoReflect() protoreflect.Message {
	mi := &file_protodef_game_map_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Cell.ProtoReflect.Descriptor instead.
func (*Cell) Descriptor() ([]byte, []int) {
	return file_protodef_game_map_proto_rawDescGZIP(), []int{0}
}

func (x *Cell) GetOccupied() bool {
	if x != nil {
		return x.Occupied
	}
	return false
}

func (x *Cell) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

type Row struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cells []*Cell `protobuf:"bytes,1,rep,name=cells,proto3" json:"cells,omitempty"`
}

func (x *Row) Reset() {
	*x = Row{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protodef_game_map_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Row) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Row) ProtoMessage() {}

func (x *Row) ProtoReflect() protoreflect.Message {
	mi := &file_protodef_game_map_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Row.ProtoReflect.Descriptor instead.
func (*Row) Descriptor() ([]byte, []int) {
	return file_protodef_game_map_proto_rawDescGZIP(), []int{1}
}

func (x *Row) GetCells() []*Cell {
	if x != nil {
		return x.Cells
	}
	return nil
}

type GameMap struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Rows []*Row `protobuf:"bytes,1,rep,name=rows,proto3" json:"rows,omitempty"`
}

func (x *GameMap) Reset() {
	*x = GameMap{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protodef_game_map_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GameMap) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GameMap) ProtoMessage() {}

func (x *GameMap) ProtoReflect() protoreflect.Message {
	mi := &file_protodef_game_map_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GameMap.ProtoReflect.Descriptor instead.
func (*GameMap) Descriptor() ([]byte, []int) {
	return file_protodef_game_map_proto_rawDescGZIP(), []int{2}
}

func (x *GameMap) GetRows() []*Row {
	if x != nil {
		return x.Rows
	}
	return nil
}

type UserPositionedGameMap struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserPosition *Position `protobuf:"bytes,1,opt,name=user_position,json=userPosition,proto3" json:"user_position,omitempty"`
	GameMap      *GameMap  `protobuf:"bytes,2,opt,name=game_map,json=gameMap,proto3" json:"game_map,omitempty"`
}

func (x *UserPositionedGameMap) Reset() {
	*x = UserPositionedGameMap{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protodef_game_map_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserPositionedGameMap) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserPositionedGameMap) ProtoMessage() {}

func (x *UserPositionedGameMap) ProtoReflect() protoreflect.Message {
	mi := &file_protodef_game_map_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserPositionedGameMap.ProtoReflect.Descriptor instead.
func (*UserPositionedGameMap) Descriptor() ([]byte, []int) {
	return file_protodef_game_map_proto_rawDescGZIP(), []int{3}
}

func (x *UserPositionedGameMap) GetUserPosition() *Position {
	if x != nil {
		return x.UserPosition
	}
	return nil
}

func (x *UserPositionedGameMap) GetGameMap() *GameMap {
	if x != nil {
		return x.GameMap
	}
	return nil
}

var File_protodef_game_map_proto protoreflect.FileDescriptor

var file_protodef_game_map_proto_rawDesc = []byte{
	0x0a, 0x17, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x64, 0x65, 0x66, 0x2f, 0x67, 0x61, 0x6d, 0x65, 0x5f,
	0x6d, 0x61, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x67, 0x61, 0x6d, 0x65, 0x5f,
	0x6d, 0x61, 0x70, 0x1a, 0x15, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x64, 0x65, 0x66, 0x2f, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x38, 0x0a, 0x04, 0x43, 0x65,
	0x6c, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x6f, 0x63, 0x63, 0x75, 0x70, 0x69, 0x65, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x6f, 0x63, 0x63, 0x75, 0x70, 0x69, 0x65, 0x64, 0x12, 0x14,
	0x0a, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6f,
	0x77, 0x6e, 0x65, 0x72, 0x22, 0x2b, 0x0a, 0x03, 0x52, 0x6f, 0x77, 0x12, 0x24, 0x0a, 0x05, 0x63,
	0x65, 0x6c, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x67, 0x61, 0x6d,
	0x65, 0x5f, 0x6d, 0x61, 0x70, 0x2e, 0x43, 0x65, 0x6c, 0x6c, 0x52, 0x05, 0x63, 0x65, 0x6c, 0x6c,
	0x73, 0x22, 0x2c, 0x0a, 0x07, 0x47, 0x61, 0x6d, 0x65, 0x4d, 0x61, 0x70, 0x12, 0x21, 0x0a, 0x04,
	0x72, 0x6f, 0x77, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x67, 0x61, 0x6d,
	0x65, 0x5f, 0x6d, 0x61, 0x70, 0x2e, 0x52, 0x6f, 0x77, 0x52, 0x04, 0x72, 0x6f, 0x77, 0x73, 0x22,
	0x7c, 0x0a, 0x15, 0x55, 0x73, 0x65, 0x72, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x65,
	0x64, 0x47, 0x61, 0x6d, 0x65, 0x4d, 0x61, 0x70, 0x12, 0x35, 0x0a, 0x0d, 0x75, 0x73, 0x65, 0x72,
	0x5f, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x10, 0x2e, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x0c, 0x75, 0x73, 0x65, 0x72, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x2c, 0x0a, 0x08, 0x67, 0x61, 0x6d, 0x65, 0x5f, 0x6d, 0x61, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x11, 0x2e, 0x67, 0x61, 0x6d, 0x65, 0x5f, 0x6d, 0x61, 0x70, 0x2e, 0x47, 0x61, 0x6d,
	0x65, 0x4d, 0x61, 0x70, 0x52, 0x07, 0x67, 0x61, 0x6d, 0x65, 0x4d, 0x61, 0x70, 0x42, 0x0c, 0x5a,
	0x0a, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x64, 0x65, 0x66, 0x50, 0x00, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protodef_game_map_proto_rawDescOnce sync.Once
	file_protodef_game_map_proto_rawDescData = file_protodef_game_map_proto_rawDesc
)

func file_protodef_game_map_proto_rawDescGZIP() []byte {
	file_protodef_game_map_proto_rawDescOnce.Do(func() {
		file_protodef_game_map_proto_rawDescData = protoimpl.X.CompressGZIP(file_protodef_game_map_proto_rawDescData)
	})
	return file_protodef_game_map_proto_rawDescData
}

var file_protodef_game_map_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_protodef_game_map_proto_goTypes = []interface{}{
	(*Cell)(nil),                  // 0: game_map.Cell
	(*Row)(nil),                   // 1: game_map.Row
	(*GameMap)(nil),               // 2: game_map.GameMap
	(*UserPositionedGameMap)(nil), // 3: game_map.UserPositionedGameMap
	(*Position)(nil),              // 4: status.Position
}
var file_protodef_game_map_proto_depIdxs = []int32{
	0, // 0: game_map.Row.cells:type_name -> game_map.Cell
	1, // 1: game_map.GameMap.rows:type_name -> game_map.Row
	4, // 2: game_map.UserPositionedGameMap.user_position:type_name -> status.Position
	2, // 3: game_map.UserPositionedGameMap.game_map:type_name -> game_map.GameMap
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_protodef_game_map_proto_init() }
func file_protodef_game_map_proto_init() {
	if File_protodef_game_map_proto != nil {
		return
	}
	file_protodef_status_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_protodef_game_map_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Cell); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protodef_game_map_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Row); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protodef_game_map_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GameMap); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protodef_game_map_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserPositionedGameMap); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_protodef_game_map_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protodef_game_map_proto_goTypes,
		DependencyIndexes: file_protodef_game_map_proto_depIdxs,
		MessageInfos:      file_protodef_game_map_proto_msgTypes,
	}.Build()
	File_protodef_game_map_proto = out.File
	file_protodef_game_map_proto_rawDesc = nil
	file_protodef_game_map_proto_goTypes = nil
	file_protodef_game_map_proto_depIdxs = nil
}
