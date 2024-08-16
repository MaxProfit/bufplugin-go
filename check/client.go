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
	"sync"

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
	//
	// The Rules will be sorted by Rule ID.
	// Returns error if duplicate Rule IDs were detected from the underlying source.
	ListRules(ctx context.Context, options ...ListRulesCallOption) ([]Rule, error)

	isClient()
}

// NewClient returns a new Client for the given pluginrpc.Client.
func NewClient(pluginrpcClient pluginrpc.Client, options ...ClientOption) Client {
	return newClient(pluginrpcClient, options...)
}

// ClientOption is an option for a new Client.
type ClientOption func(*clientOptions)

// ClientWithCacheRules returns a new ClientOption that will result in the Rules from
// ListRules being cached.
//
// The default is to not cache Rules.
func ClientWithCacheRules() ClientOption {
	return func(clientOptions *clientOptions) {
		clientOptions.cacheRules = true
	}
}

// NewClientForSpec return a new Client that directly uses the given Spec.
//
// This should primarily be used for testing.
func NewClientForSpec(spec *Spec, options ...ClientOption) (Client, error) {
	checkServiceHandler, err := newCheckServiceHandler(spec)
	if err != nil {
		return nil, err
	}
	checkServer, err := newCheckServer(checkServiceHandler)
	if err != nil {
		return nil, err
	}
	return newClient(pluginrpc.NewClient(pluginrpc.NewServerRunner(checkServer)), options...), nil
}

// CheckCallOption is an option for a Client.Check call.
type CheckCallOption func(*checkCallOptions)

// ListRulesCallOption is an option for a Client.ListRules call.
type ListRulesCallOption func(*listRulesCallOptions)

// *** PRIVATE ***

type client struct {
	pluginrpcClient pluginrpc.Client

	cacheRules     bool
	cachedRules    []Rule
	cachedRulesErr error
	lock           sync.RWMutex
}

func newClient(
	pluginrpcClient pluginrpc.Client,
	options ...ClientOption,
) *client {
	clientOptions := newClientOptions()
	for _, option := range options {
		option(clientOptions)
	}
	return &client{
		pluginrpcClient: pluginrpcClient,
		cacheRules:      clientOptions.cacheRules,
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
	protoRequests, err := request.toProtos()
	if err != nil {
		return nil, err
	}
	for _, protoRequest := range protoRequests {
		protoResponse, err := checkServiceClient.Check(ctx, protoRequest)
		if err != nil {
			return nil, err
		}
		for _, protoAnnotation := range protoResponse.GetAnnotations() {
			multiResponseWriter.addAnnotation(
				protoAnnotation.GetRuleId(),
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
	if !c.cacheRules {
		return c.listRulesUncached(ctx)
	}
	c.lock.RLock()
	if len(c.cachedRules) > 0 || c.cachedRulesErr != nil {
		c.lock.RUnlock()
		return c.cachedRules, c.cachedRulesErr
	}
	c.lock.RUnlock()

	c.lock.Lock()
	defer c.lock.Unlock()
	if len(c.cachedRules) == 0 && c.cachedRulesErr == nil {
		c.cachedRules, c.cachedRulesErr = c.listRulesUncached(ctx)
	}
	return c.cachedRules, c.cachedRulesErr
}

func (c *client) listRulesUncached(ctx context.Context) ([]Rule, error) {
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
	rules, err := xslices.MapError(protoRules, ruleForProtoRule)
	if err != nil {
		return nil, err
	}
	if err := validateNoDuplicateRules(rules); err != nil {
		return nil, err
	}
	sortRules(rules)
	return rules, nil
}

func (c *client) newCheckServiceClient() (v1beta1pluginrpc.CheckServiceClient, error) {
	return v1beta1pluginrpc.NewCheckServiceClient(c.pluginrpcClient)
}

func (*client) isClient() {}

type clientOptions struct {
	cacheRules bool
}

func newClientOptions() *clientOptions {
	return &clientOptions{}
}

type checkCallOptions struct{}

type listRulesCallOptions struct{}
