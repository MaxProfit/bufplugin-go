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
	"fmt"

	"github.com/bufbuild/bufplugin-go/internal/pkg/xslices"
	"github.com/bufbuild/protovalidate-go"
)

// RuleSpec is the spec for a Rule.
//
// It is used to construct a Rule on the server-side (i.e. within the plugin). It specifies the
// ID, categories, purpose, type, and a RuleHandler to actually run the Rule logic.
//
// Generally, these are provided to Main. This library will handle Check and ListRules calls
// based on the provided RuleSpecs.
type RuleSpec struct {
	// Required.
	ID          string
	CategoryIDs []string
	IsDefault   bool
	// Required.
	Purpose string
	// Required.
	Type           RuleType
	Deprecated     bool
	ReplacementIDs []string
	// Required.
	Handler RuleHandler
}

// *** PRIVATE ***

// Assumes that the RuleSpec is validated.
func ruleSpecToRule(ruleSpec *RuleSpec, idToCategory map[string]Category) (Rule, error) {
	categories, err := xslices.MapError(
		ruleSpec.CategoryIDs,
		func(id string) (Category, error) {
			category, ok := idToCategory[id]
			if !ok {
				return nil, fmt.Errorf("no category for id %q", id)
			}
			return category, nil
		},
	)
	if err != nil {
		return nil, err
	}
	return newRule(
		ruleSpec.ID,
		categories,
		ruleSpec.IsDefault,
		ruleSpec.Purpose,
		ruleSpec.Type,
		ruleSpec.Deprecated,
		ruleSpec.ReplacementIDs,
	), nil
}

func validateRuleSpecs(
	validator *protovalidate.Validator,
	ruleSpecs []*RuleSpec,
	categoryIDMap map[string]struct{},
) error {
	ruleIDs := xslices.Map(ruleSpecs, func(ruleSpec *RuleSpec) string { return ruleSpec.ID })
	if err := validateNoDuplicateRuleIDs(ruleIDs); err != nil {
		return err
	}
	ruleIDToRuleSpec := make(map[string]*RuleSpec)
	for _, ruleSpec := range ruleSpecs {
		if ruleSpec.ID == "" {
			return newValidateRuleSpecError("ID is empty")
		}
		ruleIDToRuleSpec[ruleSpec.ID] = ruleSpec
	}
	for _, ruleSpec := range ruleSpecs {
		if err := validateRuleSpec(validator, ruleSpec, ruleIDToRuleSpec, categoryIDMap); err != nil {
			return err
		}
	}
	return nil
}

func validateRuleSpec(
	_ *protovalidate.Validator,
	ruleSpec *RuleSpec,
	ruleIDToRuleSpec map[string]*RuleSpec,
	categoryIDMap map[string]struct{},
) error {
	for _, categoryID := range ruleSpec.CategoryIDs {
		if _, ok := categoryIDMap[categoryID]; !ok {
			return newValidateRuleSpecErrorf("no category for ID %q", categoryID)
		}
	}
	if ruleSpec.Purpose == "" {
		return newValidateRuleSpecErrorf("Purpose is not set for ID %q", ruleSpec.ID)
	}
	if ruleSpec.Type == 0 {
		return newValidateRuleSpecErrorf("Type is not set for ID %q", ruleSpec.ID)
	}
	if _, ok := ruleTypeToProtoRuleType[ruleSpec.Type]; !ok {
		return newValidateRuleSpecErrorf("Type is unknown: %q", ruleSpec.Type)
	}
	if ruleSpec.Handler == nil {
		return newValidateRuleSpecErrorf("Handler is not set for ID %q", ruleSpec.ID)
	}
	if len(ruleSpec.ReplacementIDs) > 0 && !ruleSpec.Deprecated {
		return newValidateRuleSpecErrorf("ID %q had ReplacementIDs but Deprecated was false", ruleSpec.ID)
	}
	for _, replacementID := range ruleSpec.ReplacementIDs {
		replacementRuleSpec, ok := ruleIDToRuleSpec[replacementID]
		if !ok {
			return newValidateRuleSpecErrorf("ID %q specified replacement ID %q which was not found", ruleSpec.ID, replacementID)
		}
		if replacementRuleSpec.Deprecated {
			return newValidateRuleSpecErrorf("Deprecated ID %q specified replacement ID %q which also deprecated", ruleSpec.ID, replacementID)
		}
	}
	// We do this on the server-side only, this shouldn't be used client-side.
	// TODO: This isn't working
	return nil
	// return validator.Validate(ruleSpecToRule(ruleSpec, emptyOptions).toProto())
}
