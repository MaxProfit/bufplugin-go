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

	"github.com/bufbuild/bufplugin-go/internal/pkg/xslices"
)

// NewMultiClient returns a new Client that is the union of the given Clients.
//
// The Clients must not have overlapping Rule IDs, that is ListRules must
// return unique IDs for each Client.
func NewMultiClient(clients []Client, _ ...MultiClientOption) Client {
	return newMultiClient(clients)
}

// MultiClientOption is an option for a new multi Client.
type MultiClientOption func(*multiClientOptions)

type multiClient struct {
	clients []Client
}

func newMultiClient(clients []Client) *multiClient {
	return &multiClient{
		clients: clients,
	}
}

func (c *multiClient) Check(ctx context.Context, request Request, _ ...CheckCallOption) (Response, error) {
	switch len(c.clients) {
	case 0:
		return newResponse(nil)
	case 1:
		return c.clients[0].Check(ctx, request)
	default:
		allRules, chunkedRuleIDs, err := c.getRulesAndChunkedRuleIDs(ctx)
		if err != nil {
			return nil, err
		}
		// These are the specific ruleIDs that were requested.
		requestRuleIDs := request.RuleIDs()
		if len(requestRuleIDs) == 0 {
			// If we didn't have specific ruleIDs, the requested ruleIDs are all ruleIDs.
			requestRuleIDs = xslices.Map(allRules, Rule.ID)
		}
		// This is a map of the requested ruleIDs.
		requestRuleIDMap := make(map[string]struct{})
		for _, requestRuleID := range requestRuleIDs {
			requestRuleIDMap[requestRuleID] = struct{}{}
		}

		var allAnnotations []Annotation
		for i, delegate := range c.clients {
			// This is all ruleIDs for this client.
			allDelegateRuleIDs := chunkedRuleIDs[i]
			// This is the specific requested ruleIDs for this client
			requestDelegateRuleIDs := make([]string, 0, len(allDelegateRuleIDs))
			for _, delegateRuleID := range allDelegateRuleIDs {
				// If this ruleID was requested, add it to requestDelegateRuleIDs.
				// This will result it being part of the delegate Request.
				if _, ok := requestRuleIDMap[delegateRuleID]; ok {
					requestDelegateRuleIDs = append(requestDelegateRuleIDs, delegateRuleID)
				}
			}

			delegateRequest, err := NewRequest(
				request.Files(),
				WithAgainstFiles(request.AgainstFiles()),
				WithOptions(request.Options()),
				WithRuleIDs(requestDelegateRuleIDs...),
			)
			if err != nil {
				return nil, err
			}
			delegateResponse, err := delegate.Check(ctx, delegateRequest)
			if err != nil {
				return nil, err
			}
			allAnnotations = append(allAnnotations, delegateResponse.Annotations()...)
		}

		return newResponse(allAnnotations)
	}
}

func (c *multiClient) ListRules(ctx context.Context, _ ...ListRulesCallOption) ([]Rule, error) {
	switch len(c.clients) {
	case 0:
		return nil, nil
	case 1:
		return c.clients[0].ListRules(ctx)
	default:
		rules, _, err := c.getRulesAndChunkedRuleIDs(ctx)
		if err != nil {
			return nil, err
		}
		return rules, nil
	}
}

// Each []string within the returned [][]string is a slice of ruleIDs that corresponds
// to the client at the same index.
//
// For example, chunkedRuleIDs[1] corresponds to the ruleIDs for c.clients[1].
func (c *multiClient) getRulesAndChunkedRuleIDs(ctx context.Context) ([]Rule, [][]string, error) {
	var rules []Rule
	chunkedRuleIDs := make([][]string, len(c.clients))
	for i, delegate := range c.clients {
		delegateRules, err := delegate.ListRules(ctx)
		if err != nil {
			return nil, nil, err
		}
		rules = append(rules, delegateRules...)
		chunkedRuleIDs[i] = xslices.Map(delegateRules, Rule.ID)
	}
	if err := validateNoDuplicateRules(rules); err != nil {
		return nil, nil, err
	}
	sortRules(rules)
	return rules, chunkedRuleIDs, nil
}

func (*multiClient) isClient() {}

type multiClientOptions struct{}
