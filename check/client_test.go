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
	"testing"

	"github.com/bufbuild/bufplugin-go/internal/pkg/xslices"
	"github.com/stretchr/testify/require"
)

func TestClientListRulesCategories(t *testing.T) {
	t.Parallel()

	testClientListRulesCategories(t)
	testClientListRulesCategories(t, ClientWithCacheRulesAndCategories())
}

func testClientListRulesCategories(t *testing.T, options ...ClientOption) {
	ctx := context.Background()
	client, err := NewClientForSpec(
		&Spec{
			Rules: []*RuleSpec{
				{
					ID:      "rule1",
					Purpose: "Test rule1.",
					Type:    RuleTypeLint,
					Handler: RuleHandlerFunc(func(context.Context, ResponseWriter, Request) error { return nil }),
				},
				{
					ID: "rule2",
					CategoryIDs: []string{
						"category1",
					},
					Purpose: "Test rule2.",
					Type:    RuleTypeLint,
					Handler: RuleHandlerFunc(func(context.Context, ResponseWriter, Request) error { return nil }),
				},
				{
					ID: "rule3",
					CategoryIDs: []string{
						"category1",
						"category2",
					},
					Purpose: "Test rule3.",
					Type:    RuleTypeLint,
					Handler: RuleHandlerFunc(func(context.Context, ResponseWriter, Request) error { return nil }),
				},
			},
			Categories: []*CategorySpec{
				{
					ID:      "category1",
					Purpose: "Test category1.",
				},
				{
					ID:      "category2",
					Purpose: "Test category2.",
				},
			},
		},
		options...,
	)
	require.NoError(t, err)
	rules, err := client.ListRules(ctx)
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"rule1",
			"rule2",
			"rule3",
		},
		xslices.Map(rules, Rule.ID),
	)
	categories, err := client.ListCategories(ctx)
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"category1",
			"category2",
		},
		xslices.Map(categories, Category.ID),
	)
	categories = rules[0].Categories()
	require.Empty(t, categories)
	categories = rules[1].Categories()
	require.Equal(
		t,
		[]string{
			"category1",
		},
		xslices.Map(categories, Category.ID),
	)
	categories = rules[2].Categories()
	require.Equal(
		t,
		[]string{
			"category1",
			"category2",
		},
		xslices.Map(categories, Category.ID),
	)
}
