// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: authz/v1/authz.proto

package pb

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

const (
	CertificateAuthority_Register_FullMethodName       = "/authz.v1.CertificateAuthority/Register"
	CertificateAuthority_GetCertificate_FullMethodName = "/authz.v1.CertificateAuthority/GetCertificate"
	CertificateAuthority_DeleteService_FullMethodName  = "/authz.v1.CertificateAuthority/DeleteService"
	CertificateAuthority_PublicKey_FullMethodName      = "/authz.v1.CertificateAuthority/PublicKey"
)

// CertificateAuthorityClient is the client API for CertificateAuthority service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CertificateAuthorityClient interface {
	Register(ctx context.Context, in *CertificateRequest, opts ...grpc.CallOption) (*CertificateResponse, error)
	GetCertificate(ctx context.Context, in *CertificateRequest, opts ...grpc.CallOption) (*CertificateResponse, error)
	DeleteService(ctx context.Context, in *DeletionRequest, opts ...grpc.CallOption) (*DeletionResponse, error)
	PublicKey(ctx context.Context, in *PublicKeyRequest, opts ...grpc.CallOption) (*PublicKeyResponse, error)
}

type certificateAuthorityClient struct {
	cc grpc.ClientConnInterface
}

func NewCertificateAuthorityClient(cc grpc.ClientConnInterface) CertificateAuthorityClient {
	return &certificateAuthorityClient{cc}
}

func (c *certificateAuthorityClient) Register(ctx context.Context, in *CertificateRequest, opts ...grpc.CallOption) (*CertificateResponse, error) {
	out := new(CertificateResponse)
	err := c.cc.Invoke(ctx, CertificateAuthority_Register_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *certificateAuthorityClient) GetCertificate(ctx context.Context, in *CertificateRequest, opts ...grpc.CallOption) (*CertificateResponse, error) {
	out := new(CertificateResponse)
	err := c.cc.Invoke(ctx, CertificateAuthority_GetCertificate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *certificateAuthorityClient) DeleteService(ctx context.Context, in *DeletionRequest, opts ...grpc.CallOption) (*DeletionResponse, error) {
	out := new(DeletionResponse)
	err := c.cc.Invoke(ctx, CertificateAuthority_DeleteService_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *certificateAuthorityClient) PublicKey(ctx context.Context, in *PublicKeyRequest, opts ...grpc.CallOption) (*PublicKeyResponse, error) {
	out := new(PublicKeyResponse)
	err := c.cc.Invoke(ctx, CertificateAuthority_PublicKey_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CertificateAuthorityServer is the server API for CertificateAuthority service.
// All implementations must embed UnimplementedCertificateAuthorityServer
// for forward compatibility
type CertificateAuthorityServer interface {
	Register(context.Context, *CertificateRequest) (*CertificateResponse, error)
	GetCertificate(context.Context, *CertificateRequest) (*CertificateResponse, error)
	DeleteService(context.Context, *DeletionRequest) (*DeletionResponse, error)
	PublicKey(context.Context, *PublicKeyRequest) (*PublicKeyResponse, error)
	mustEmbedUnimplementedCertificateAuthorityServer()
}

// UnimplementedCertificateAuthorityServer must be embedded to have forward compatible implementations.
type UnimplementedCertificateAuthorityServer struct {
}

func (UnimplementedCertificateAuthorityServer) Register(context.Context, *CertificateRequest) (*CertificateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedCertificateAuthorityServer) GetCertificate(context.Context, *CertificateRequest) (*CertificateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCertificate not implemented")
}
func (UnimplementedCertificateAuthorityServer) DeleteService(context.Context, *DeletionRequest) (*DeletionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteService not implemented")
}
func (UnimplementedCertificateAuthorityServer) PublicKey(context.Context, *PublicKeyRequest) (*PublicKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PublicKey not implemented")
}
func (UnimplementedCertificateAuthorityServer) mustEmbedUnimplementedCertificateAuthorityServer() {}

// UnsafeCertificateAuthorityServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CertificateAuthorityServer will
// result in compilation errors.
type UnsafeCertificateAuthorityServer interface {
	mustEmbedUnimplementedCertificateAuthorityServer()
}

func RegisterCertificateAuthorityServer(s grpc.ServiceRegistrar, srv CertificateAuthorityServer) {
	s.RegisterService(&CertificateAuthority_ServiceDesc, srv)
}

func _CertificateAuthority_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CertificateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertificateAuthorityServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CertificateAuthority_Register_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertificateAuthorityServer).Register(ctx, req.(*CertificateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CertificateAuthority_GetCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CertificateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertificateAuthorityServer).GetCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CertificateAuthority_GetCertificate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertificateAuthorityServer).GetCertificate(ctx, req.(*CertificateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CertificateAuthority_DeleteService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeletionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertificateAuthorityServer).DeleteService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CertificateAuthority_DeleteService_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertificateAuthorityServer).DeleteService(ctx, req.(*DeletionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CertificateAuthority_PublicKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PublicKeyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CertificateAuthorityServer).PublicKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CertificateAuthority_PublicKey_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CertificateAuthorityServer).PublicKey(ctx, req.(*PublicKeyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CertificateAuthority_ServiceDesc is the grpc.ServiceDesc for CertificateAuthority service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CertificateAuthority_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "authz.v1.CertificateAuthority",
	HandlerType: (*CertificateAuthorityServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _CertificateAuthority_Register_Handler,
		},
		{
			MethodName: "GetCertificate",
			Handler:    _CertificateAuthority_GetCertificate_Handler,
		},
		{
			MethodName: "DeleteService",
			Handler:    _CertificateAuthority_DeleteService_Handler,
		},
		{
			MethodName: "PublicKey",
			Handler:    _CertificateAuthority_PublicKey_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "authz/v1/authz.proto",
}

const (
	Authz_Register_FullMethodName = "/authz.v1.Authz/Register"
	Authz_Login_FullMethodName    = "/authz.v1.Authz/Login"
	Authz_GetToken_FullMethodName = "/authz.v1.Authz/GetToken"
)

// AuthzClient is the client API for Authz service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuthzClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
	GetToken(ctx context.Context, in *TokenRequest, opts ...grpc.CallOption) (*TokenResponse, error)
}

type authzClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthzClient(cc grpc.ClientConnInterface) AuthzClient {
	return &authzClient{cc}
}

func (c *authzClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, Authz_Register_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authzClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	out := new(LoginResponse)
	err := c.cc.Invoke(ctx, Authz_Login_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authzClient) GetToken(ctx context.Context, in *TokenRequest, opts ...grpc.CallOption) (*TokenResponse, error) {
	out := new(TokenResponse)
	err := c.cc.Invoke(ctx, Authz_GetToken_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthzServer is the server API for Authz service.
// All implementations must embed UnimplementedAuthzServer
// for forward compatibility
type AuthzServer interface {
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	Login(context.Context, *LoginRequest) (*LoginResponse, error)
	GetToken(context.Context, *TokenRequest) (*TokenResponse, error)
	mustEmbedUnimplementedAuthzServer()
}

// UnimplementedAuthzServer must be embedded to have forward compatible implementations.
type UnimplementedAuthzServer struct {
}

func (UnimplementedAuthzServer) Register(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedAuthzServer) Login(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedAuthzServer) GetToken(context.Context, *TokenRequest) (*TokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetToken not implemented")
}
func (UnimplementedAuthzServer) mustEmbedUnimplementedAuthzServer() {}

// UnsafeAuthzServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuthzServer will
// result in compilation errors.
type UnsafeAuthzServer interface {
	mustEmbedUnimplementedAuthzServer()
}

func RegisterAuthzServer(s grpc.ServiceRegistrar, srv AuthzServer) {
	s.RegisterService(&Authz_ServiceDesc, srv)
}

func _Authz_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthzServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Authz_Register_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthzServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Authz_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthzServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Authz_Login_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthzServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Authz_GetToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthzServer).GetToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Authz_GetToken_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthzServer).GetToken(ctx, req.(*TokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Authz_ServiceDesc is the grpc.ServiceDesc for Authz service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Authz_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "authz.v1.Authz",
	HandlerType: (*AuthzServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _Authz_Register_Handler,
		},
		{
			MethodName: "Login",
			Handler:    _Authz_Login_Handler,
		},
		{
			MethodName: "GetToken",
			Handler:    _Authz_GetToken_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "authz/v1/authz.proto",
}