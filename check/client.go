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
	listRulesPageSize      = 250
	listCategoriesPageSize = 250
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
	// ListCategories lists all available Categories from the plugin.
	//
	// The Categories will be sorted by Category ID.
	// Returns error if duplicate Category IDs were detected from the underlying source.
	ListCategories(ctx context.Context, options ...ListCategoriesCallOption) ([]Category, error)

	isClient()
}

// NewClient returns a new Client for the given pluginrpc.Client.
func NewClient(pluginrpcClient pluginrpc.Client, options ...ClientOption) Client {
	return newClient(pluginrpcClient, options...)
}

// ClientOption is an option for a new Client.
type ClientOption func(*clientOptions)

// ClientWithCacheRulesAndCategories returns a new ClientOption that will result in the Rules from
// ListRules and the Categories from ListCategories being cached.
//
// The default is to not cache Rules or Categories.
func ClientWithCacheRulesAndCategories() ClientOption {
	return func(clientOptions *clientOptions) {
		clientOptions.cacheRulesAndCategories = true
	}
}

// NewClientForSpec return a new Client that directly uses the given Spec.
//
// This should primarily be used for testing.
func NewClientForSpec(spec *Spec, options ...ClientOption) (Client, error) {
	checkServiceHandler, err := newCheckServiceHandler(spec, 0)
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

// ListCategoriesCallOption is an option for a Client.ListCategories call.
type ListCategoriesCallOption func(*listCategoriesCallOptions)

// *** PRIVATE ***

type client struct {
	pluginrpcClient pluginrpc.Client

	cacheRulesAndCategories bool

	cachedRules    []Rule
	cachedRulesErr error

	cachedCategories    []Category
	cachedCategoriesErr error

	// Lock ordering: rulesLock -> categoriesLock
	rulesLock      sync.RWMutex
	categoriesLock sync.RWMutex
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
		pluginrpcClient:         pluginrpcClient,
		cacheRulesAndCategories: clientOptions.cacheRulesAndCategories,
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
	if !c.cacheRulesAndCategories {
		return c.listRulesUncached(ctx)
	}
	c.rulesLock.RLock()
	if len(c.cachedRules) > 0 || c.cachedRulesErr != nil {
		c.rulesLock.RUnlock()
		return c.cachedRules, c.cachedRulesErr
	}
	c.rulesLock.RUnlock()

	c.rulesLock.Lock()
	defer c.rulesLock.Unlock()
	if len(c.cachedRules) == 0 && c.cachedRulesErr == nil {
		c.cachedRules, c.cachedRulesErr = c.listRulesUncached(ctx)
	}
	return c.cachedRules, c.cachedRulesErr
}

func (c *client) ListCategories(ctx context.Context, _ ...ListCategoriesCallOption) ([]Category, error) {
	if !c.cacheRulesAndCategories {
		return c.listCategoriesUncached(ctx)
	}
	c.categoriesLock.RLock()
	if len(c.cachedCategories) > 0 || c.cachedCategoriesErr != nil {
		c.categoriesLock.RUnlock()
		return c.cachedCategories, c.cachedCategoriesErr
	}
	c.categoriesLock.RUnlock()

	c.categoriesLock.Lock()
	defer c.categoriesLock.Unlock()
	if len(c.cachedCategories) == 0 && c.cachedCategoriesErr == nil {
		c.cachedCategories, c.cachedCategoriesErr = c.listCategoriesUncached(ctx)
	}
	return c.cachedCategories, c.cachedCategoriesErr
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

	// We acquire rulesLock before categoriesLock.
	categories, err := c.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	categoryIDToCategory := make(map[string]Category)
	for _, category := range categories {
		// We know there are no duplicate IDs from validation.
		categoryIDToCategory[category.ID()] = category
	}
	rules, err := xslices.MapError(
		protoRules,
		func(protoRule *checkv1beta1.Rule) (Rule, error) {
			return ruleForProtoRule(protoRule, categoryIDToCategory)
		},
	)
	if err != nil {
		return nil, err
	}
	if err := validateNoDuplicateRules(rules); err != nil {
		return nil, err
	}
	sortRules(rules)
	return rules, nil
}

func (c *client) listCategoriesUncached(ctx context.Context) ([]Category, error) {
	checkServiceClient, err := c.newCheckServiceClient()
	if err != nil {
		return nil, err
	}
	var protoCategories []*checkv1beta1.Category
	var pageToken string
	for {
		response, err := checkServiceClient.ListCategories(
			ctx,
			&checkv1beta1.ListCategoriesRequest{
				PageSize:  listCategoriesPageSize,
				PageToken: pageToken,
			},
		)
		if err != nil {
			return nil, err
		}
		protoCategories = append(protoCategories, response.GetCategories()...)
		pageToken = response.GetNextPageToken()
		if pageToken == "" {
			break
		}
	}
	categories, err := xslices.MapError(protoCategories, categoryForProtoCategory)
	if err != nil {
		return nil, err
	}
	if err := validateNoDuplicateCategories(categories); err != nil {
		return nil, err
	}
	sortCategories(categories)
	return categories, nil
}

func (c *client) newCheckServiceClient() (v1beta1pluginrpc.CheckServiceClient, error) {
	return v1beta1pluginrpc.NewCheckServiceClient(c.pluginrpcClient)
}

func (*client) isClient() {}

type clientOptions struct {
	cacheRulesAndCategories bool
}

func newClientOptions() *clientOptions {
	return &clientOptions{}
}

type checkCallOptions struct{}

type listRulesCallOptions struct{}

type listCategoriesCallOptions struct{}
