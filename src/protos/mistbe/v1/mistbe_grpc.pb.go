// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: mistbe/v1/mistbe.proto

package mistbe

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	MistBEService_CreateAppserver_FullMethodName      = "/mistbe.v1.MistBEService/CreateAppserver"
	MistBEService_GetByIdAppserver_FullMethodName     = "/mistbe.v1.MistBEService/GetByIdAppserver"
	MistBEService_ListAppservers_FullMethodName       = "/mistbe.v1.MistBEService/ListAppservers"
	MistBEService_DeleteAppserver_FullMethodName      = "/mistbe.v1.MistBEService/DeleteAppserver"
	MistBEService_CreateAppserverSub_FullMethodName   = "/mistbe.v1.MistBEService/CreateAppserverSub"
	MistBEService_GetUserAppserverSubs_FullMethodName = "/mistbe.v1.MistBEService/GetUserAppserverSubs"
	MistBEService_DeleteAppserverSub_FullMethodName   = "/mistbe.v1.MistBEService/DeleteAppserverSub"
	MistBEService_CreateAppserverRole_FullMethodName  = "/mistbe.v1.MistBEService/CreateAppserverRole"
	MistBEService_GetAllAppserverRoles_FullMethodName = "/mistbe.v1.MistBEService/GetAllAppserverRoles"
	MistBEService_DeleteAppserverRole_FullMethodName  = "/mistbe.v1.MistBEService/DeleteAppserverRole"
	MistBEService_CreateChannel_FullMethodName        = "/mistbe.v1.MistBEService/CreateChannel"
	MistBEService_GetByIdChannel_FullMethodName       = "/mistbe.v1.MistBEService/GetByIdChannel"
	MistBEService_ListChannels_FullMethodName         = "/mistbe.v1.MistBEService/ListChannels"
	MistBEService_DeleteChannel_FullMethodName        = "/mistbe.v1.MistBEService/DeleteChannel"
)

// MistBEServiceClient is the client API for MistBEService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// The greeting service definition.
type MistBEServiceClient interface {
	// ----- APPSERVER ----
	CreateAppserver(ctx context.Context, in *CreateAppserverRequest, opts ...grpc.CallOption) (*CreateAppserverResponse, error)
	GetByIdAppserver(ctx context.Context, in *GetByIdAppserverRequest, opts ...grpc.CallOption) (*GetByIdAppserverResponse, error)
	ListAppservers(ctx context.Context, in *ListAppserversRequest, opts ...grpc.CallOption) (*ListAppserversResponse, error)
	DeleteAppserver(ctx context.Context, in *DeleteAppserverRequest, opts ...grpc.CallOption) (*DeleteAppserverResponse, error)
	// ----- APPSERVER SUB -----
	CreateAppserverSub(ctx context.Context, in *CreateAppserverSubRequest, opts ...grpc.CallOption) (*CreateAppserverSubResponse, error)
	GetUserAppserverSubs(ctx context.Context, in *GetUserAppserverSubsRequest, opts ...grpc.CallOption) (*GetUserAppserverSubsResponse, error)
	DeleteAppserverSub(ctx context.Context, in *DeleteAppserverSubRequest, opts ...grpc.CallOption) (*DeleteAppserverSubResponse, error)
	// ----- APPSERVER ROLE -----
	CreateAppserverRole(ctx context.Context, in *CreateAppserverRoleRequest, opts ...grpc.CallOption) (*CreateAppserverRoleResponse, error)
	GetAllAppserverRoles(ctx context.Context, in *GetAllAppserverRolesRequest, opts ...grpc.CallOption) (*GetAllAppserverRolesResponse, error)
	DeleteAppserverRole(ctx context.Context, in *DeleteAppserverRoleRequest, opts ...grpc.CallOption) (*DeleteAppserverRoleResponse, error)
	// ----- CHANNEL ----
	CreateChannel(ctx context.Context, in *CreateChannelRequest, opts ...grpc.CallOption) (*CreateChannelResponse, error)
	GetByIdChannel(ctx context.Context, in *GetByIdChannelRequest, opts ...grpc.CallOption) (*GetByIdChannelResponse, error)
	ListChannels(ctx context.Context, in *ListChannelsRequest, opts ...grpc.CallOption) (*ListChannelsResponse, error)
	DeleteChannel(ctx context.Context, in *DeleteChannelRequest, opts ...grpc.CallOption) (*DeleteChannelResponse, error)
}

type mistBEServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMistBEServiceClient(cc grpc.ClientConnInterface) MistBEServiceClient {
	return &mistBEServiceClient{cc}
}

func (c *mistBEServiceClient) CreateAppserver(ctx context.Context, in *CreateAppserverRequest, opts ...grpc.CallOption) (*CreateAppserverResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateAppserverResponse)
	err := c.cc.Invoke(ctx, MistBEService_CreateAppserver_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mistBEServiceClient) GetByIdAppserver(ctx context.Context, in *GetByIdAppserverRequest, opts ...grpc.CallOption) (*GetByIdAppserverResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetByIdAppserverResponse)
	err := c.cc.Invoke(ctx, MistBEService_GetByIdAppserver_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mistBEServiceClient) ListAppservers(ctx context.Context, in *ListAppserversRequest, opts ...grpc.CallOption) (*ListAppserversResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListAppserversResponse)
	err := c.cc.Invoke(ctx, MistBEService_ListAppservers_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mistBEServiceClient) DeleteAppserver(ctx context.Context, in *DeleteAppserverRequest, opts ...grpc.CallOption) (*DeleteAppserverResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteAppserverResponse)
	err := c.cc.Invoke(ctx, MistBEService_DeleteAppserver_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mistBEServiceClient) CreateAppserverSub(ctx context.Context, in *CreateAppserverSubRequest, opts ...grpc.CallOption) (*CreateAppserverSubResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateAppserverSubResponse)
	err := c.cc.Invoke(ctx, MistBEService_CreateAppserverSub_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mistBEServiceClient) GetUserAppserverSubs(ctx context.Context, in *GetUserAppserverSubsRequest, opts ...grpc.CallOption) (*GetUserAppserverSubsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetUserAppserverSubsResponse)
	err := c.cc.Invoke(ctx, MistBEService_GetUserAppserverSubs_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mistBEServiceClient) DeleteAppserverSub(ctx context.Context, in *DeleteAppserverSubRequest, opts ...grpc.CallOption) (*DeleteAppserverSubResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteAppserverSubResponse)
	err := c.cc.Invoke(ctx, MistBEService_DeleteAppserverSub_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mistBEServiceClient) CreateAppserverRole(ctx context.Context, in *CreateAppserverRoleRequest, opts ...grpc.CallOption) (*CreateAppserverRoleResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateAppserverRoleResponse)
	err := c.cc.Invoke(ctx, MistBEService_CreateAppserverRole_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mistBEServiceClient) GetAllAppserverRoles(ctx context.Context, in *GetAllAppserverRolesRequest, opts ...grpc.CallOption) (*GetAllAppserverRolesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetAllAppserverRolesResponse)
	err := c.cc.Invoke(ctx, MistBEService_GetAllAppserverRoles_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mistBEServiceClient) DeleteAppserverRole(ctx context.Context, in *DeleteAppserverRoleRequest, opts ...grpc.CallOption) (*DeleteAppserverRoleResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteAppserverRoleResponse)
	err := c.cc.Invoke(ctx, MistBEService_DeleteAppserverRole_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mistBEServiceClient) CreateChannel(ctx context.Context, in *CreateChannelRequest, opts ...grpc.CallOption) (*CreateChannelResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateChannelResponse)
	err := c.cc.Invoke(ctx, MistBEService_CreateChannel_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mistBEServiceClient) GetByIdChannel(ctx context.Context, in *GetByIdChannelRequest, opts ...grpc.CallOption) (*GetByIdChannelResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetByIdChannelResponse)
	err := c.cc.Invoke(ctx, MistBEService_GetByIdChannel_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mistBEServiceClient) ListChannels(ctx context.Context, in *ListChannelsRequest, opts ...grpc.CallOption) (*ListChannelsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListChannelsResponse)
	err := c.cc.Invoke(ctx, MistBEService_ListChannels_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mistBEServiceClient) DeleteChannel(ctx context.Context, in *DeleteChannelRequest, opts ...grpc.CallOption) (*DeleteChannelResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteChannelResponse)
	err := c.cc.Invoke(ctx, MistBEService_DeleteChannel_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MistBEServiceServer is the server API for MistBEService service.
// All implementations must embed UnimplementedMistBEServiceServer
// for forward compatibility.
//
// The greeting service definition.
type MistBEServiceServer interface {
	// ----- APPSERVER ----
	CreateAppserver(context.Context, *CreateAppserverRequest) (*CreateAppserverResponse, error)
	GetByIdAppserver(context.Context, *GetByIdAppserverRequest) (*GetByIdAppserverResponse, error)
	ListAppservers(context.Context, *ListAppserversRequest) (*ListAppserversResponse, error)
	DeleteAppserver(context.Context, *DeleteAppserverRequest) (*DeleteAppserverResponse, error)
	// ----- APPSERVER SUB -----
	CreateAppserverSub(context.Context, *CreateAppserverSubRequest) (*CreateAppserverSubResponse, error)
	GetUserAppserverSubs(context.Context, *GetUserAppserverSubsRequest) (*GetUserAppserverSubsResponse, error)
	DeleteAppserverSub(context.Context, *DeleteAppserverSubRequest) (*DeleteAppserverSubResponse, error)
	// ----- APPSERVER ROLE -----
	CreateAppserverRole(context.Context, *CreateAppserverRoleRequest) (*CreateAppserverRoleResponse, error)
	GetAllAppserverRoles(context.Context, *GetAllAppserverRolesRequest) (*GetAllAppserverRolesResponse, error)
	DeleteAppserverRole(context.Context, *DeleteAppserverRoleRequest) (*DeleteAppserverRoleResponse, error)
	// ----- CHANNEL ----
	CreateChannel(context.Context, *CreateChannelRequest) (*CreateChannelResponse, error)
	GetByIdChannel(context.Context, *GetByIdChannelRequest) (*GetByIdChannelResponse, error)
	ListChannels(context.Context, *ListChannelsRequest) (*ListChannelsResponse, error)
	DeleteChannel(context.Context, *DeleteChannelRequest) (*DeleteChannelResponse, error)
	mustEmbedUnimplementedMistBEServiceServer()
}

// UnimplementedMistBEServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedMistBEServiceServer struct{}

func (UnimplementedMistBEServiceServer) CreateAppserver(context.Context, *CreateAppserverRequest) (*CreateAppserverResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAppserver not implemented")
}
func (UnimplementedMistBEServiceServer) GetByIdAppserver(context.Context, *GetByIdAppserverRequest) (*GetByIdAppserverResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetByIdAppserver not implemented")
}
func (UnimplementedMistBEServiceServer) ListAppservers(context.Context, *ListAppserversRequest) (*ListAppserversResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAppservers not implemented")
}
func (UnimplementedMistBEServiceServer) DeleteAppserver(context.Context, *DeleteAppserverRequest) (*DeleteAppserverResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAppserver not implemented")
}
func (UnimplementedMistBEServiceServer) CreateAppserverSub(context.Context, *CreateAppserverSubRequest) (*CreateAppserverSubResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAppserverSub not implemented")
}
func (UnimplementedMistBEServiceServer) GetUserAppserverSubs(context.Context, *GetUserAppserverSubsRequest) (*GetUserAppserverSubsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserAppserverSubs not implemented")
}
func (UnimplementedMistBEServiceServer) DeleteAppserverSub(context.Context, *DeleteAppserverSubRequest) (*DeleteAppserverSubResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAppserverSub not implemented")
}
func (UnimplementedMistBEServiceServer) CreateAppserverRole(context.Context, *CreateAppserverRoleRequest) (*CreateAppserverRoleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAppserverRole not implemented")
}
func (UnimplementedMistBEServiceServer) GetAllAppserverRoles(context.Context, *GetAllAppserverRolesRequest) (*GetAllAppserverRolesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllAppserverRoles not implemented")
}
func (UnimplementedMistBEServiceServer) DeleteAppserverRole(context.Context, *DeleteAppserverRoleRequest) (*DeleteAppserverRoleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAppserverRole not implemented")
}
func (UnimplementedMistBEServiceServer) CreateChannel(context.Context, *CreateChannelRequest) (*CreateChannelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateChannel not implemented")
}
func (UnimplementedMistBEServiceServer) GetByIdChannel(context.Context, *GetByIdChannelRequest) (*GetByIdChannelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetByIdChannel not implemented")
}
func (UnimplementedMistBEServiceServer) ListChannels(context.Context, *ListChannelsRequest) (*ListChannelsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListChannels not implemented")
}
func (UnimplementedMistBEServiceServer) DeleteChannel(context.Context, *DeleteChannelRequest) (*DeleteChannelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteChannel not implemented")
}
func (UnimplementedMistBEServiceServer) mustEmbedUnimplementedMistBEServiceServer() {}
func (UnimplementedMistBEServiceServer) testEmbeddedByValue()                       {}

// UnsafeMistBEServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MistBEServiceServer will
// result in compilation errors.
type UnsafeMistBEServiceServer interface {
	mustEmbedUnimplementedMistBEServiceServer()
}

func RegisterMistBEServiceServer(s grpc.ServiceRegistrar, srv MistBEServiceServer) {
	// If the following call pancis, it indicates UnimplementedMistBEServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&MistBEService_ServiceDesc, srv)
}

func _MistBEService_CreateAppserver_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAppserverRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MistBEServiceServer).CreateAppserver(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MistBEService_CreateAppserver_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MistBEServiceServer).CreateAppserver(ctx, req.(*CreateAppserverRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MistBEService_GetByIdAppserver_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetByIdAppserverRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MistBEServiceServer).GetByIdAppserver(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MistBEService_GetByIdAppserver_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MistBEServiceServer).GetByIdAppserver(ctx, req.(*GetByIdAppserverRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MistBEService_ListAppservers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListAppserversRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MistBEServiceServer).ListAppservers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MistBEService_ListAppservers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MistBEServiceServer).ListAppservers(ctx, req.(*ListAppserversRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MistBEService_DeleteAppserver_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteAppserverRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MistBEServiceServer).DeleteAppserver(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MistBEService_DeleteAppserver_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MistBEServiceServer).DeleteAppserver(ctx, req.(*DeleteAppserverRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MistBEService_CreateAppserverSub_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAppserverSubRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MistBEServiceServer).CreateAppserverSub(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MistBEService_CreateAppserverSub_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MistBEServiceServer).CreateAppserverSub(ctx, req.(*CreateAppserverSubRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MistBEService_GetUserAppserverSubs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserAppserverSubsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MistBEServiceServer).GetUserAppserverSubs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MistBEService_GetUserAppserverSubs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MistBEServiceServer).GetUserAppserverSubs(ctx, req.(*GetUserAppserverSubsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MistBEService_DeleteAppserverSub_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteAppserverSubRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MistBEServiceServer).DeleteAppserverSub(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MistBEService_DeleteAppserverSub_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MistBEServiceServer).DeleteAppserverSub(ctx, req.(*DeleteAppserverSubRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MistBEService_CreateAppserverRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAppserverRoleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MistBEServiceServer).CreateAppserverRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MistBEService_CreateAppserverRole_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MistBEServiceServer).CreateAppserverRole(ctx, req.(*CreateAppserverRoleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MistBEService_GetAllAppserverRoles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAllAppserverRolesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MistBEServiceServer).GetAllAppserverRoles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MistBEService_GetAllAppserverRoles_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MistBEServiceServer).GetAllAppserverRoles(ctx, req.(*GetAllAppserverRolesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MistBEService_DeleteAppserverRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteAppserverRoleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MistBEServiceServer).DeleteAppserverRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MistBEService_DeleteAppserverRole_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MistBEServiceServer).DeleteAppserverRole(ctx, req.(*DeleteAppserverRoleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MistBEService_CreateChannel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateChannelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MistBEServiceServer).CreateChannel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MistBEService_CreateChannel_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MistBEServiceServer).CreateChannel(ctx, req.(*CreateChannelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MistBEService_GetByIdChannel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetByIdChannelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MistBEServiceServer).GetByIdChannel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MistBEService_GetByIdChannel_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MistBEServiceServer).GetByIdChannel(ctx, req.(*GetByIdChannelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MistBEService_ListChannels_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListChannelsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MistBEServiceServer).ListChannels(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MistBEService_ListChannels_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MistBEServiceServer).ListChannels(ctx, req.(*ListChannelsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MistBEService_DeleteChannel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteChannelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MistBEServiceServer).DeleteChannel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MistBEService_DeleteChannel_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MistBEServiceServer).DeleteChannel(ctx, req.(*DeleteChannelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MistBEService_ServiceDesc is the grpc.ServiceDesc for MistBEService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MistBEService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mistbe.v1.MistBEService",
	HandlerType: (*MistBEServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateAppserver",
			Handler:    _MistBEService_CreateAppserver_Handler,
		},
		{
			MethodName: "GetByIdAppserver",
			Handler:    _MistBEService_GetByIdAppserver_Handler,
		},
		{
			MethodName: "ListAppservers",
			Handler:    _MistBEService_ListAppservers_Handler,
		},
		{
			MethodName: "DeleteAppserver",
			Handler:    _MistBEService_DeleteAppserver_Handler,
		},
		{
			MethodName: "CreateAppserverSub",
			Handler:    _MistBEService_CreateAppserverSub_Handler,
		},
		{
			MethodName: "GetUserAppserverSubs",
			Handler:    _MistBEService_GetUserAppserverSubs_Handler,
		},
		{
			MethodName: "DeleteAppserverSub",
			Handler:    _MistBEService_DeleteAppserverSub_Handler,
		},
		{
			MethodName: "CreateAppserverRole",
			Handler:    _MistBEService_CreateAppserverRole_Handler,
		},
		{
			MethodName: "GetAllAppserverRoles",
			Handler:    _MistBEService_GetAllAppserverRoles_Handler,
		},
		{
			MethodName: "DeleteAppserverRole",
			Handler:    _MistBEService_DeleteAppserverRole_Handler,
		},
		{
			MethodName: "CreateChannel",
			Handler:    _MistBEService_CreateChannel_Handler,
		},
		{
			MethodName: "GetByIdChannel",
			Handler:    _MistBEService_GetByIdChannel_Handler,
		},
		{
			MethodName: "ListChannels",
			Handler:    _MistBEService_ListChannels_Handler,
		},
		{
			MethodName: "DeleteChannel",
			Handler:    _MistBEService_DeleteChannel_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mistbe/v1/mistbe.proto",
}
