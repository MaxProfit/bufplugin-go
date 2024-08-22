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

	"github.com/bufbuild/protovalidate-go"
	"github.com/stretchr/testify/require"
)

func TestValidateSpec(t *testing.T) {
	t.Parallel()

	validator, err := protovalidate.New()
	require.NoError(t, err)

	validateRuleSpecError := &validateRuleSpecError{}
	validateCategorySpecError := &validateCategorySpecError{}
	validateSpecError := &validateSpecError{}

	// Simple spec that passes validation.
	spec := &Spec{
		Rules: []*RuleSpec{
			testNewSimpleLintRuleSpec("rule1", nil, true, false, nil),
			testNewSimpleLintRuleSpec("rule2", []string{"category1"}, true, false, nil),
			testNewSimpleLintRuleSpec("rule3", []string{"category1", "category2"}, true, false, nil),
		},
		Categories: []*CategorySpec{
			testNewSimpleCategorySpec("category1", false, nil),
			testNewSimpleCategorySpec("category2", false, nil),
		},
	}
	require.NoError(t, validateSpec(validator, spec))

	// More complicated spec with deprecated rules and categories that passes validation.
	spec = &Spec{
		Rules: []*RuleSpec{
			testNewSimpleLintRuleSpec("rule1", nil, true, false, nil),
			testNewSimpleLintRuleSpec("rule2", []string{"category1"}, true, false, nil),
			testNewSimpleLintRuleSpec("rule3", []string{"category1", "category2"}, true, false, nil),
			testNewSimpleLintRuleSpec("rule4", []string{"category1"}, false, true, []string{"rule1"}),
			testNewSimpleLintRuleSpec("rule5", []string{"category3", "category4"}, false, true, []string{"rule2", "rule3"}),
		},
		Categories: []*CategorySpec{
			testNewSimpleCategorySpec("category1", false, nil),
			testNewSimpleCategorySpec("category2", false, nil),
			testNewSimpleCategorySpec("category3", true, nil),
			testNewSimpleCategorySpec("category4", true, nil),
		},
	}
	require.NoError(t, validateSpec(validator, spec))

	// Spec that has rules with categories with no resulting category spec.
	spec = &Spec{
		Rules: []*RuleSpec{
			testNewSimpleLintRuleSpec("rule1", nil, true, false, nil),
			testNewSimpleLintRuleSpec("rule2", []string{"category1"}, true, false, nil),
			testNewSimpleLintRuleSpec("rule3", []string{"category1", "category2"}, true, false, nil),
		},
		Categories: []*CategorySpec{
			testNewSimpleCategorySpec("category1", false, nil),
		},
	}
	require.ErrorAs(t, validateSpec(validator, spec), &validateRuleSpecError)

	// Spec that has categories with no rules with those categories.
	spec = &Spec{
		Rules: []*RuleSpec{
			testNewSimpleLintRuleSpec("rule1", nil, true, false, nil),
			testNewSimpleLintRuleSpec("rule2", []string{"category1"}, true, false, nil),
			testNewSimpleLintRuleSpec("rule3", []string{"category1", "category2"}, true, false, nil),
		},
		Categories: []*CategorySpec{
			testNewSimpleCategorySpec("category1", false, nil),
			testNewSimpleCategorySpec("category2", false, nil),
			testNewSimpleCategorySpec("category3", false, nil),
			testNewSimpleCategorySpec("category4", false, nil),
		},
	}
	require.ErrorAs(t, validateSpec(validator, spec), &validateCategorySpecError)

	// Spec that has overlapping rules and categories.
	spec = &Spec{
		Rules: []*RuleSpec{
			testNewSimpleLintRuleSpec("rule1", nil, true, false, nil),
			testNewSimpleLintRuleSpec("rule2", []string{"category1"}, true, false, nil),
			testNewSimpleLintRuleSpec("rule3", []string{"category1", "category2"}, true, false, nil),
		},
		Categories: []*CategorySpec{
			testNewSimpleCategorySpec("category1", false, nil),
			testNewSimpleCategorySpec("category2", false, nil),
			testNewSimpleCategorySpec("rule3", false, nil),
		},
	}
	require.ErrorAs(t, validateSpec(validator, spec), &validateSpecError)
}

func testNewSimpleLintRuleSpec(
	id string,
	categoryIDs []string,
	isDefault bool,
	deprecated bool,
	replacementIDs []string,
) *RuleSpec {
	return &RuleSpec{
		ID:             id,
		CategoryIDs:    categoryIDs,
		IsDefault:      isDefault,
		Purpose:        "Checks " + id + ".",
		Type:           RuleTypeLint,
		Deprecated:     deprecated,
		ReplacementIDs: replacementIDs,
		Handler: RuleHandlerFunc(
			func(context.Context, ResponseWriter, Request) error {
				return nil
			},
		),
	}
}

func testNewSimpleCategorySpec(
	id string,
	deprecated bool,
	replacementIDs []string,
) *CategorySpec {
	return &CategorySpec{
		ID:             id,
		Purpose:        "Checks " + id + ".",
		Deprecated:     deprecated,
		ReplacementIDs: replacementIDs,
	}
}
