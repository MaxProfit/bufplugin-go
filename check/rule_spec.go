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

	"github.com/bufbuild/protovalidate-go"
)

// RuleSpec is the spec for a Rule.
//
// It is used to construct a Rule on the server-side (i.e. within the plugin). It specifies the
// ID, categories, purpose, type, and a RuleHandler to actually run the Rule logic. The purpose
// and RuleHandler can change depending on the options provided to the plugin at runtime.
//
// Generally, these are provided to Main. This library will handle Check and ListRules calls
// based on the provided RuleSpecs.
type RuleSpec struct {
	// Required.
	ID         string
	Categories []string
	// Required.
	Purpose func(options Options) string
	// Required.
	Type           RuleType
	Deprecated     bool
	ReplacementIDs []string
	// Required.
	Handler func(options Options) RuleHandler
}

// NopPurpose is a convenience function that returns a static purpose value that does not change
// depending on options provided at runtime.
//
//	ruleSpec := &check.RuleSpec{
//		ID: "FOO",
//		Purpose: check.NopPurpose("Checks foo."),
//		Type: check.RuleTypeLint,
//		Handler: check.NopRuleHandler(ruleHandler),
//	}
func NopPurpose(purpose string) func(Options) string {
	return func(Options) string { return purpose }
}

// NopRuleHandler is a convenience function that returns a static RuleHandler that does not change
// depending on options provided at runtime.
//
//	ruleSpec := &check.RuleSpec{
//		ID: "FOO",
//		Purpose: check.NopPurpose("Checks foo."),
//		Type: check.RuleTypeLint,
//		Handler: check.NopRuleHandler(ruleHandler),
//	}
func NopRuleHandler(ruleHandler RuleHandler) func(Options) RuleHandler {
	return func(Options) RuleHandler { return ruleHandler }
}

// *** PRIVATE ***

// Assumes that the RuleSpec is validated.
func ruleSpecToRule(ruleSpec *RuleSpec, options Options) Rule {
	return newRule(
		ruleSpec.ID,
		ruleSpec.Categories,
		ruleSpec.Purpose(options),
		ruleSpec.Type,
		ruleSpec.Deprecated,
		ruleSpec.ReplacementIDs,
	)
}

func validateRuleSpec(_ *protovalidate.Validator, ruleSpec *RuleSpec) error {
	if ruleSpec.ID == "" {
		return errors.New("RuleSpec.ID is empty")
	}
	if ruleSpec.Purpose == nil {
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
