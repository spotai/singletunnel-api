// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: coord.proto

package coord

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ForServersClient is the client API for ForServers service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ForServersClient interface {
	// Accepts a request containing a ServerInfo describing the server,
	// then returns a response containing a ServerConfig describing how
	// the server needs to configure its networking services to accept the clients
	// that are connecting/connected to it.
	GetServerConfig(ctx context.Context, in *GetServerConfigRequest, opts ...grpc.CallOption) (*GetServerConfigResponse, error)
	ShuttingDown(ctx context.Context, in *ShutdownNotification, opts ...grpc.CallOption) (*ShutdownResponse, error)
}

type forServersClient struct {
	cc grpc.ClientConnInterface
}

func NewForServersClient(cc grpc.ClientConnInterface) ForServersClient {
	return &forServersClient{cc}
}

func (c *forServersClient) GetServerConfig(ctx context.Context, in *GetServerConfigRequest, opts ...grpc.CallOption) (*GetServerConfigResponse, error) {
	out := new(GetServerConfigResponse)
	err := c.cc.Invoke(ctx, "/coord.ForServers/GetServerConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *forServersClient) ShuttingDown(ctx context.Context, in *ShutdownNotification, opts ...grpc.CallOption) (*ShutdownResponse, error) {
	out := new(ShutdownResponse)
	err := c.cc.Invoke(ctx, "/coord.ForServers/ShuttingDown", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ForServersServer is the server API for ForServers service.
// All implementations must embed UnimplementedForServersServer
// for forward compatibility
type ForServersServer interface {
	// Accepts a request containing a ServerInfo describing the server,
	// then returns a response containing a ServerConfig describing how
	// the server needs to configure its networking services to accept the clients
	// that are connecting/connected to it.
	GetServerConfig(context.Context, *GetServerConfigRequest) (*GetServerConfigResponse, error)
	ShuttingDown(context.Context, *ShutdownNotification) (*ShutdownResponse, error)
	mustEmbedUnimplementedForServersServer()
}

// UnimplementedForServersServer must be embedded to have forward compatible implementations.
type UnimplementedForServersServer struct {
}

func (UnimplementedForServersServer) GetServerConfig(context.Context, *GetServerConfigRequest) (*GetServerConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetServerConfig not implemented")
}
func (UnimplementedForServersServer) ShuttingDown(context.Context, *ShutdownNotification) (*ShutdownResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ShuttingDown not implemented")
}
func (UnimplementedForServersServer) mustEmbedUnimplementedForServersServer() {}

// UnsafeForServersServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ForServersServer will
// result in compilation errors.
type UnsafeForServersServer interface {
	mustEmbedUnimplementedForServersServer()
}

func RegisterForServersServer(s grpc.ServiceRegistrar, srv ForServersServer) {
	s.RegisterService(&ForServers_ServiceDesc, srv)
}

func _ForServers_GetServerConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetServerConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ForServersServer).GetServerConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/coord.ForServers/GetServerConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ForServersServer).GetServerConfig(ctx, req.(*GetServerConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ForServers_ShuttingDown_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ShutdownNotification)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ForServersServer).ShuttingDown(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/coord.ForServers/ShuttingDown",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ForServersServer).ShuttingDown(ctx, req.(*ShutdownNotification))
	}
	return interceptor(ctx, in, info, handler)
}

// ForServers_ServiceDesc is the grpc.ServiceDesc for ForServers service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ForServers_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "coord.ForServers",
	HandlerType: (*ForServersServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetServerConfig",
			Handler:    _ForServers_GetServerConfig_Handler,
		},
		{
			MethodName: "ShuttingDown",
			Handler:    _ForServers_ShuttingDown_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "coord.proto",
}

// ForClientsClient is the client API for ForClients service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ForClientsClient interface {
	ListServers(ctx context.Context, in *ClientInfo, opts ...grpc.CallOption) (*ServersInfo, error)
	Connect(ctx context.Context, in *ConnectionSpec, opts ...grpc.CallOption) (*ClientConfig, error)
	// Disconnects the client from any servers it's currently connected to.
	Disconnect(ctx context.Context, in *DisconnectRequest, opts ...grpc.CallOption) (*DisconnectResponse, error)
	WaitEvent(ctx context.Context, in *WaitEventSpec, opts ...grpc.CallOption) (*Event, error)
}

type forClientsClient struct {
	cc grpc.ClientConnInterface
}

func NewForClientsClient(cc grpc.ClientConnInterface) ForClientsClient {
	return &forClientsClient{cc}
}

func (c *forClientsClient) ListServers(ctx context.Context, in *ClientInfo, opts ...grpc.CallOption) (*ServersInfo, error) {
	out := new(ServersInfo)
	err := c.cc.Invoke(ctx, "/coord.ForClients/ListServers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *forClientsClient) Connect(ctx context.Context, in *ConnectionSpec, opts ...grpc.CallOption) (*ClientConfig, error) {
	out := new(ClientConfig)
	err := c.cc.Invoke(ctx, "/coord.ForClients/Connect", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *forClientsClient) Disconnect(ctx context.Context, in *DisconnectRequest, opts ...grpc.CallOption) (*DisconnectResponse, error) {
	out := new(DisconnectResponse)
	err := c.cc.Invoke(ctx, "/coord.ForClients/Disconnect", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *forClientsClient) WaitEvent(ctx context.Context, in *WaitEventSpec, opts ...grpc.CallOption) (*Event, error) {
	out := new(Event)
	err := c.cc.Invoke(ctx, "/coord.ForClients/WaitEvent", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ForClientsServer is the server API for ForClients service.
// All implementations must embed UnimplementedForClientsServer
// for forward compatibility
type ForClientsServer interface {
	ListServers(context.Context, *ClientInfo) (*ServersInfo, error)
	Connect(context.Context, *ConnectionSpec) (*ClientConfig, error)
	// Disconnects the client from any servers it's currently connected to.
	Disconnect(context.Context, *DisconnectRequest) (*DisconnectResponse, error)
	WaitEvent(context.Context, *WaitEventSpec) (*Event, error)
	mustEmbedUnimplementedForClientsServer()
}

// UnimplementedForClientsServer must be embedded to have forward compatible implementations.
type UnimplementedForClientsServer struct {
}

func (UnimplementedForClientsServer) ListServers(context.Context, *ClientInfo) (*ServersInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListServers not implemented")
}
func (UnimplementedForClientsServer) Connect(context.Context, *ConnectionSpec) (*ClientConfig, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Connect not implemented")
}
func (UnimplementedForClientsServer) Disconnect(context.Context, *DisconnectRequest) (*DisconnectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Disconnect not implemented")
}
func (UnimplementedForClientsServer) WaitEvent(context.Context, *WaitEventSpec) (*Event, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WaitEvent not implemented")
}
func (UnimplementedForClientsServer) mustEmbedUnimplementedForClientsServer() {}

// UnsafeForClientsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ForClientsServer will
// result in compilation errors.
type UnsafeForClientsServer interface {
	mustEmbedUnimplementedForClientsServer()
}

func RegisterForClientsServer(s grpc.ServiceRegistrar, srv ForClientsServer) {
	s.RegisterService(&ForClients_ServiceDesc, srv)
}

func _ForClients_ListServers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClientInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ForClientsServer).ListServers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/coord.ForClients/ListServers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ForClientsServer).ListServers(ctx, req.(*ClientInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _ForClients_Connect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConnectionSpec)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ForClientsServer).Connect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/coord.ForClients/Connect",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ForClientsServer).Connect(ctx, req.(*ConnectionSpec))
	}
	return interceptor(ctx, in, info, handler)
}

func _ForClients_Disconnect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DisconnectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ForClientsServer).Disconnect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/coord.ForClients/Disconnect",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ForClientsServer).Disconnect(ctx, req.(*DisconnectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ForClients_WaitEvent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WaitEventSpec)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ForClientsServer).WaitEvent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/coord.ForClients/WaitEvent",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ForClientsServer).WaitEvent(ctx, req.(*WaitEventSpec))
	}
	return interceptor(ctx, in, info, handler)
}

// ForClients_ServiceDesc is the grpc.ServiceDesc for ForClients service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ForClients_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "coord.ForClients",
	HandlerType: (*ForClientsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListServers",
			Handler:    _ForClients_ListServers_Handler,
		},
		{
			MethodName: "Connect",
			Handler:    _ForClients_Connect_Handler,
		},
		{
			MethodName: "Disconnect",
			Handler:    _ForClients_Disconnect_Handler,
		},
		{
			MethodName: "WaitEvent",
			Handler:    _ForClients_WaitEvent_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "coord.proto",
}
