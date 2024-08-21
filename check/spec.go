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
	"errors"
	"fmt"
	"slices"

	"github.com/bufbuild/bufplugin-go/internal/pkg/xslices"
	"github.com/bufbuild/protovalidate-go"
)

// Spec is the spec for a plugin.
//
// It is used to construct a plugin on the server-side (i.e. within the plugin).
//
// Generally, this is provided to Main. This library will handle Check and ListRules calls
// based on the provided RuleSpecs.
type Spec struct {
	// Required.
	//
	// All RuleSpecs must have Category IDs that match a CategorySpec within Categories.
	//
	// No IDs can overlap with Category IDs in Categories.
	Rules []*RuleSpec
	// Required if any RuleSpec specifies a category.
	//
	// All CategorySpecs must have an ID that matches at least one Category ID on a
	// RuleSpec within Rules.
	//
	// No IDs can overlap with Rule IDs in Rules.
	Categories []*CategorySpec

	// Before is a function that will be executed before any RuleHandlers are
	// invoked that returns a new Context and Request. This new Context and
	// Request will be passed to the RuleHandlers. This allows for any
	// pre-processing that needs to occur.
	Before func(ctx context.Context, request Request) (context.Context, Request, error)
}

// *** PRIVATE ***

func validateSpec(validator *protovalidate.Validator, spec *Spec) error {
	if len(spec.Rules) == 0 {
		return errors.New("Spec.Rules is empty")
	}
	ruleIDs := xslices.Map(spec.Rules, func(ruleSpec *RuleSpec) string { return ruleSpec.ID })
	if err := validateNoDuplicateRuleIDs(ruleIDs); err != nil {
		return err
	}
	categoryIDs := xslices.Map(spec.Categories, func(categorySpec *CategorySpec) string { return categorySpec.ID })
	if err := validateNoDuplicateCategoryIDs(categoryIDs); err != nil {
		return err
	}
	ruleAndCategoryIDs := append(slices.Clone(ruleIDs), categoryIDs...)
	if err := validateNoDuplicateRuleOrCategoryIDs(ruleAndCategoryIDs); err != nil {
		return err
	}
	categoryIDMap := xslices.ToStructMap(categoryIDs)
	categoryIDForRulesMap := make(map[string]struct{})
	for _, ruleSpec := range spec.Rules {
		if err := validateRuleSpec(validator, ruleSpec, categoryIDMap); err != nil {
			return err
		}
		for _, categoryID := range ruleSpec.CategoryIDs {
			categoryIDForRulesMap[categoryID] = struct{}{}
		}
	}
	for _, categorySpec := range spec.Categories {
		if err := validateCategorySpec(validator, categorySpec); err != nil {
			return err
		}
		if _, ok := categoryIDForRulesMap[categorySpec.ID]; !ok {
			return fmt.Errorf("no Rule has a Category ID of %q", categorySpec.ID)
		}
	}
	return nil
}
