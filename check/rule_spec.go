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
	"errors"
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

func validateRuleSpec(_ *protovalidate.Validator, ruleSpec *RuleSpec, categoryIDMap map[string]struct{}) error {
	if ruleSpec.ID == "" {
		return errors.New("RuleSpec.ID is empty")
	}
	for _, categoryID := range ruleSpec.CategoryIDs {
		if _, ok := categoryIDMap[categoryID]; !ok {
			return fmt.Errorf("no category for id %q", categoryID)
		}
	}
	if ruleSpec.Purpose == "" {
		return fmt.Errorf("RuleSpec.Purpose is not set for ID %q", ruleSpec.ID)
	}
	if ruleSpec.Type == 0 {
		return fmt.Errorf("RuleSpec.Type is not set for ID %q", ruleSpec.ID)
	}
	if _, ok := ruleTypeToProtoRuleType[ruleSpec.Type]; !ok {
		return fmt.Errorf("RuleSpec.Type is unknown: %q", ruleSpec.Type)
	}
	if ruleSpec.Handler == nil {
		return fmt.Errorf("RuleSpec.Handler is not set for ID %q", ruleSpec.ID)
	}
	if len(ruleSpec.ReplacementIDs) > 0 && !ruleSpec.Deprecated {
		return fmt.Errorf("RuleSpec.ReplacementIDs had values %v but Deprecated was false", ruleSpec.ReplacementIDs)
	}
	// We do this on the server-side only, this shouldn't be used client-side.
	// TODO: This isn't working
	return nil
	// return validator.Validate(ruleSpecToRule(ruleSpec, emptyOptions).toProto())
}
