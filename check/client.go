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

package check

import (
	"context"

	checkv1beta1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/check/v1beta1"
	"github.com/bufbuild/bufplugin-go/internal/gen/buf/plugin/check/v1beta1/v1beta1pluginrpc"
	"github.com/bufbuild/bufplugin-go/internal/pkg/xslices"
	"github.com/bufbuild/pluginrpc-go"
)

const (
	listRulesPageSize = 250
)

// Client is a client for a custom lint or breaking change plugin.
type Client interface {
	// Check invokes a check using the plugin..
	Check(ctx context.Context, request Request, options ...CheckCallOption) (Response, error)
	// ListRules lists all available Rules from the plugin.
	ListRules(ctx context.Context, options ...ListRulesCallOption) ([]Rule, error)
}

// NewClient returns a new Client for the given pluginrpc.Client.
func NewClient(pluginrpcClient pluginrpc.Client) Client {
	return newClient(pluginrpcClient)
}

// NewClientForSpec return a new Client that directly uses the given Spec.
//
// This should primarily be used for testing.
func NewClientForSpec(spec *Spec) (Client, error) {
	checkServiceHandler, err := newCheckServiceHandler(spec)
	if err != nil {
		return nil, err
	}
	checkServer, err := newCheckServer(checkServiceHandler)
	if err != nil {
		return nil, err
	}
	return newClient(pluginrpc.NewClient(pluginrpc.NewServerRunner(checkServer))), nil
}

// CheckCallOption is an option for a Client.Check call.
type CheckCallOption func(*checkCallOptions)

// ListRulesCallOption is an option for a Client.ListRules call.
type ListRulesCallOption func(*listRulesCallOptions)

// *** PRIVATE ***

type client struct {
	pluginrpcClient pluginrpc.Client
}

func newClient(
	pluginrpcClient pluginrpc.Client,
) *client {
	return &client{
		pluginrpcClient: pluginrpcClient,
	}
}

func (c *client) Check(ctx context.Context, request Request, _ ...CheckCallOption) (Response, error) {
	checkServiceClient, err := c.newCheckServiceClient()
	if err != nil {
		return nil, err
	}
	multiResponseWriter, err := newMultiResponseWriter(request)
	if err != nil {
		return nil, err
	}
	for _, protoRequest := range request.toProtos() {
		protoResponse, err := checkServiceClient.Check(ctx, protoRequest)
		if err != nil {
			return nil, err
		}
		for _, protoAnnotation := range protoResponse.GetAnnotations() {
			multiResponseWriter.addAnnotation(
				protoAnnotation.GetId(),
				WithMessage(protoAnnotation.GetMessage()),
				WithFileName(protoAnnotation.GetLocation().GetFileName()),
				WithSourcePath(protoAnnotation.GetLocation().GetSourcePath()),
				WithAgainstFileName(protoAnnotation.GetAgainstLocation().GetFileName()),
				WithAgainstSourcePath(protoAnnotation.GetAgainstLocation().GetSourcePath()),
			)
		}
	}
	return multiResponseWriter.toResponse()
}

func (c *client) ListRules(ctx context.Context, _ ...ListRulesCallOption) ([]Rule, error) {
	checkServiceClient, err := c.newCheckServiceClient()
	if err != nil {
		return nil, err
	}
	var protoRules []*checkv1beta1.Rule
	var pageToken string
	for {
		response, err := checkServiceClient.ListRules(
			ctx,
			&checkv1beta1.ListRulesRequest{
				PageSize:  listRulesPageSize,
				PageToken: pageToken,
			},
		)
		if err != nil {
			return nil, err
		}
		protoRules = append(protoRules, response.GetRules()...)
		pageToken = response.GetNextPageToken()
		if pageToken == "" {
			break
		}
	}
	return xslices.MapError(protoRules, ruleForProtoRule)
}

func (c *client) newCheckServiceClient() (v1beta1pluginrpc.CheckServiceClient, error) {
	return v1beta1pluginrpc.NewCheckServiceClient(c.pluginrpcClient)
}

type checkCallOptions struct{}

type listRulesCallOptions struct{}
