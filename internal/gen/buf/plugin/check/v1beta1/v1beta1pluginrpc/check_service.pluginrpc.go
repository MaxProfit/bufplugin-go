// Copyright 2024 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-pluginrpc-go. DO NOT EDIT.
//
// Source: buf/plugin/check/v1beta1/check_service.proto

package v1beta1pluginrpc

import (
	v1beta1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/check/v1beta1"
	context "context"
	fmt "fmt"
	pluginrpc_go "github.com/bufbuild/pluginrpc-go"
)

// This is a compile-time assertion to ensure that this generated file and the pluginrpc package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of pluginrpc newer than the one compiled into your binary. You can fix
// the problem by either regenerating this code with an older version of pluginrpc or updating the
// pluginrpc version compiled into your binary.
const _ = pluginrpc_go.IsAtLeastVersion0_1_0

const (
	// CheckServiceCheckPath is the path of the CheckService's Check RPC.
	CheckServiceCheckPath = "/buf.plugin.check.v1beta1.CheckService/Check"
	// CheckServiceListRulesPath is the path of the CheckService's ListRules RPC.
	CheckServiceListRulesPath = "/buf.plugin.check.v1beta1.CheckService/ListRules"
	// CheckServiceListCategoriesPath is the path of the CheckService's ListCategories RPC.
	CheckServiceListCategoriesPath = "/buf.plugin.check.v1beta1.CheckService/ListCategories"
)

// CheckServiceSpecBuilder builds a Spec for the buf.plugin.check.v1beta1.CheckService service.
type CheckServiceSpecBuilder struct {
	Check          []pluginrpc_go.ProcedureOption
	ListRules      []pluginrpc_go.ProcedureOption
	ListCategories []pluginrpc_go.ProcedureOption
}

// Build builds a Spec for the buf.plugin.check.v1beta1.CheckService service.
func (s CheckServiceSpecBuilder) Build() (pluginrpc_go.Spec, error) {
	procedures := make([]pluginrpc_go.Procedure, 0, 3)
	procedure, err := pluginrpc_go.NewProcedure(CheckServiceCheckPath, s.Check...)
	if err != nil {
		return nil, err
	}
	procedures = append(procedures, procedure)
	procedure, err = pluginrpc_go.NewProcedure(CheckServiceListRulesPath, s.ListRules...)
	if err != nil {
		return nil, err
	}
	procedures = append(procedures, procedure)
	procedure, err = pluginrpc_go.NewProcedure(CheckServiceListCategoriesPath, s.ListCategories...)
	if err != nil {
		return nil, err
	}
	procedures = append(procedures, procedure)
	return pluginrpc_go.NewSpec(procedures)
}

// CheckServiceClient is a client for the buf.plugin.check.v1beta1.CheckService service.
type CheckServiceClient interface {
	// Check a set of Files for failures.
	//
	// All Annotations returned will have an ID that is contained within a Rule listed by ListRules.
	Check(context.Context, *v1beta1.CheckRequest, ...pluginrpc_go.CallOption) (*v1beta1.CheckResponse, error)
	// List all rules that this service implements.
	ListRules(context.Context, *v1beta1.ListRulesRequest, ...pluginrpc_go.CallOption) (*v1beta1.ListRulesResponse, error)
	// List all categories that this service implements.
	ListCategories(context.Context, *v1beta1.ListCategoriesRequest, ...pluginrpc_go.CallOption) (*v1beta1.ListCategoriesResponse, error)
}

// NewCheckServiceClient constructs a client for the buf.plugin.check.v1beta1.CheckService service.
func NewCheckServiceClient(client pluginrpc_go.Client) (CheckServiceClient, error) {
	return &checkServiceClient{
		client: client,
	}, nil
}

// CheckServiceHandler is an implementation of the buf.plugin.check.v1beta1.CheckService service.
type CheckServiceHandler interface {
	// Check a set of Files for failures.
	//
	// All Annotations returned will have an ID that is contained within a Rule listed by ListRules.
	Check(context.Context, *v1beta1.CheckRequest) (*v1beta1.CheckResponse, error)
	// List all rules that this service implements.
	ListRules(context.Context, *v1beta1.ListRulesRequest) (*v1beta1.ListRulesResponse, error)
	// List all categories that this service implements.
	ListCategories(context.Context, *v1beta1.ListCategoriesRequest) (*v1beta1.ListCategoriesResponse, error)
}

// CheckServiceServer serves the buf.plugin.check.v1beta1.CheckService service.
type CheckServiceServer interface {
	// Check a set of Files for failures.
	//
	// All Annotations returned will have an ID that is contained within a Rule listed by ListRules.
	Check(context.Context, pluginrpc_go.HandleEnv, ...pluginrpc_go.HandleOption) error
	// List all rules that this service implements.
	ListRules(context.Context, pluginrpc_go.HandleEnv, ...pluginrpc_go.HandleOption) error
	// List all categories that this service implements.
	ListCategories(context.Context, pluginrpc_go.HandleEnv, ...pluginrpc_go.HandleOption) error
}

// NewCheckServiceServer constructs a server for the buf.plugin.check.v1beta1.CheckService service.
func NewCheckServiceServer(handler pluginrpc_go.Handler, checkServiceHandler CheckServiceHandler) CheckServiceServer {
	return &checkServiceServer{
		handler:             handler,
		checkServiceHandler: checkServiceHandler,
	}
}

// RegisterCheckServiceServer registers the server for the buf.plugin.check.v1beta1.CheckService
// service.
func RegisterCheckServiceServer(serverRegistrar pluginrpc_go.ServerRegistrar, checkServiceServer CheckServiceServer) {
	serverRegistrar.Register(CheckServiceCheckPath, checkServiceServer.Check)
	serverRegistrar.Register(CheckServiceListRulesPath, checkServiceServer.ListRules)
	serverRegistrar.Register(CheckServiceListCategoriesPath, checkServiceServer.ListCategories)
}

// *** PRIVATE ***

// checkServiceClient implements CheckServiceClient.
type checkServiceClient struct {
	client pluginrpc_go.Client
}

// Check calls buf.plugin.check.v1beta1.CheckService.Check.
func (c *checkServiceClient) Check(ctx context.Context, req *v1beta1.CheckRequest, opts ...pluginrpc_go.CallOption) (*v1beta1.CheckResponse, error) {
	res := &v1beta1.CheckResponse{}
	if err := c.client.Call(ctx, CheckServiceCheckPath, req, res, opts...); err != nil {
		return nil, err
	}
	return res, nil
}

// ListRules calls buf.plugin.check.v1beta1.CheckService.ListRules.
func (c *checkServiceClient) ListRules(ctx context.Context, req *v1beta1.ListRulesRequest, opts ...pluginrpc_go.CallOption) (*v1beta1.ListRulesResponse, error) {
	res := &v1beta1.ListRulesResponse{}
	if err := c.client.Call(ctx, CheckServiceListRulesPath, req, res, opts...); err != nil {
		return nil, err
	}
	return res, nil
}

// ListCategories calls buf.plugin.check.v1beta1.CheckService.ListCategories.
func (c *checkServiceClient) ListCategories(ctx context.Context, req *v1beta1.ListCategoriesRequest, opts ...pluginrpc_go.CallOption) (*v1beta1.ListCategoriesResponse, error) {
	res := &v1beta1.ListCategoriesResponse{}
	if err := c.client.Call(ctx, CheckServiceListCategoriesPath, req, res, opts...); err != nil {
		return nil, err
	}
	return res, nil
}

// checkServiceServer implements CheckServiceServer.
type checkServiceServer struct {
	handler             pluginrpc_go.Handler
	checkServiceHandler CheckServiceHandler
}

// Check calls buf.plugin.check.v1beta1.CheckService.Check.
func (c *checkServiceServer) Check(ctx context.Context, handleEnv pluginrpc_go.HandleEnv, options ...pluginrpc_go.HandleOption) error {
	return c.handler.Handle(
		ctx,
		handleEnv,
		&v1beta1.CheckRequest{},
		func(ctx context.Context, anyReq any) (any, error) {
			req, ok := anyReq.(*v1beta1.CheckRequest)
			if !ok {
				return nil, fmt.Errorf("could not cast %T to a *v1beta1.CheckRequest", anyReq)
			}
			return c.checkServiceHandler.Check(ctx, req)
		},
		options...,
	)
}

// ListRules calls buf.plugin.check.v1beta1.CheckService.ListRules.
func (c *checkServiceServer) ListRules(ctx context.Context, handleEnv pluginrpc_go.HandleEnv, options ...pluginrpc_go.HandleOption) error {
	return c.handler.Handle(
		ctx,
		handleEnv,
		&v1beta1.ListRulesRequest{},
		func(ctx context.Context, anyReq any) (any, error) {
			req, ok := anyReq.(*v1beta1.ListRulesRequest)
			if !ok {
				return nil, fmt.Errorf("could not cast %T to a *v1beta1.ListRulesRequest", anyReq)
			}
			return c.checkServiceHandler.ListRules(ctx, req)
		},
		options...,
	)
}

// ListCategories calls buf.plugin.check.v1beta1.CheckService.ListCategories.
func (c *checkServiceServer) ListCategories(ctx context.Context, handleEnv pluginrpc_go.HandleEnv, options ...pluginrpc_go.HandleOption) error {
	return c.handler.Handle(
		ctx,
		handleEnv,
		&v1beta1.ListCategoriesRequest{},
		func(ctx context.Context, anyReq any) (any, error) {
			req, ok := anyReq.(*v1beta1.ListCategoriesRequest)
			if !ok {
				return nil, fmt.Errorf("could not cast %T to a *v1beta1.ListCategoriesRequest", anyReq)
			}
			return c.checkServiceHandler.ListCategories(ctx, req)
		},
		options...,
	)
}
