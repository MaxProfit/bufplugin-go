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
	"slices"
	"sort"

	checkv1beta1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/check/v1beta1"
	"github.com/bufbuild/bufplugin-go/internal/pkg/xslices"
)

// Rule is a single lint or breaking change rule.
//
// Rules have unique IDs. On the server-side (i.e. the plugin), Rules are created
// by RuleSpecs. Clients can list all available plugin Rules by calling ListRules.
type Rule interface {
	// ID is the ID of the Rule.
	//
	// This uniquely identifies the Rule.
	//
	// This must have at least four characters.
	// This must start and end with a capital letter from A-Z , and only consist
	// of capital letters from A-Z and underscores.
	ID() string
	// The categories that the Rule is a part of.
	//
	// Buf uses categories to include or exclude sets of rules via configuration.
	//
	// Optional.
	Categories() []Category
	// Whether or not the Rule is a default Rule.
	//
	// If a Rule is a default Rule, it will be called if a Request specifies no specific Rule IDs.
	//
	// A deprecated rule cannot be a default rule.
	IsDefault() bool
	// A user-displayable purpose of the rule.
	//
	// Always present.
	//
	// This should be a proper sentence that starts with a capital letter and ends in a period.
	Purpose() string
	// Type is the type of the Rule.
	Type() RuleType
	// Deprecated returns whether or not this Rule is deprecated.
	//
	// If the Rule is deprecated, it may be replaced by 0 or more Rules. These will be denoted
	// by ReplacementIDs.
	Deprecated() bool
	// ReplacementIDs returns the IDs of the Rules that replace this Rule, if this Rule is deprecated.
	//
	// This means that the combination of the Rules specified by ReplacementIDs replace this Rule entirely,
	// and this Rule is considered equivalent to the AND of the rules specified by ReplacementIDs.
	//
	// This will only be non-empty if Deprecated is true.
	//
	// It is not valid for a deprecated Rule to specfiy another deprecated Rule as a replacement.
	ReplacementIDs() []string

	toProto() *checkv1beta1.Rule

	isRule()
}

// *** PRIVATE ***

type rule struct {
	id             string
	categories     []Category
	isDefault      bool
	purpose        string
	ruleType       RuleType
	deprecated     bool
	replacementIDs []string
}

func newRule(
	id string,
	categories []Category,
	isDefault bool,
	purpose string,
	ruleType RuleType,
	deprecated bool,
	replacementIDs []string,
) *rule {
	return &rule{
		id:             id,
		categories:     categories,
		isDefault:      isDefault,
		purpose:        purpose,
		ruleType:       ruleType,
		deprecated:     deprecated,
		replacementIDs: replacementIDs,
	}
}

func (r *rule) ID() string {
	return r.id
}

func (r *rule) Categories() []Category {
	return slices.Clone(r.categories)
}

func (r *rule) IsDefault() bool {
	return r.isDefault
}

func (r *rule) Purpose() string {
	return r.purpose
}

func (r *rule) Type() RuleType {
	return r.ruleType
}

func (r *rule) Deprecated() bool {
	return r.deprecated
}

func (r *rule) ReplacementIDs() []string {
	return slices.Clone(r.replacementIDs)
}

func (r *rule) toProto() *checkv1beta1.Rule {
	if r == nil {
		return nil
	}
	protoRuleType := ruleTypeToProtoRuleType[r.ruleType]
	return &checkv1beta1.Rule{
		Id:             r.id,
		CategoryIds:    xslices.Map(r.categories, Category.ID),
		Default:        r.isDefault,
		Purpose:        r.purpose,
		Type:           protoRuleType,
		Deprecated:     r.deprecated,
		ReplacementIds: r.replacementIDs,
	}
}

func (*rule) isRule() {}

func ruleForProtoRule(protoRule *checkv1beta1.Rule, idToCategory map[string]Category) (Rule, error) {
	categories, err := xslices.MapError(
		protoRule.GetCategoryIds(),
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
	// TODO: We need to do some validation, even if we can't do full-on protovalidate (should we?)
	ruleType := protoRuleTypeToRuleType[protoRule.GetType()]
	return newRule(
		protoRule.GetId(),
		categories,
		protoRule.GetDefault(),
		protoRule.GetPurpose(),
		ruleType,
		protoRule.GetDeprecated(),
		protoRule.GetReplacementIds(),
	), nil
}

func sortRules(rules []Rule) {
	sort.Slice(rules, func(i int, j int) bool { return CompareRules(rules[i], rules[j]) < 0 })
}

func validateNoDuplicateRules(rules []Rule) error {
	return validateNoDuplicateRuleIDs(xslices.Map(rules, Rule.ID))
}

func validateNoDuplicateRuleIDs(ids []string) error {
	idToCount := make(map[string]int, len(ids))
	for _, id := range ids {
		idToCount[id]++
	}
	var duplicateIDs []string
	for id, count := range idToCount {
		if count > 1 {
			duplicateIDs = append(duplicateIDs, id)
		}
	}
	if len(duplicateIDs) > 0 {
		sort.Strings(duplicateIDs)
		return newDuplicateRuleIDError(duplicateIDs)
	}
	return nil
}

func validateNoDuplicateRuleOrCategoryIDs(ids []string) error {
	idToCount := make(map[string]int, len(ids))
	for _, id := range ids {
		idToCount[id]++
	}
	var duplicateIDs []string
	for id, count := range idToCount {
		if count > 1 {
			duplicateIDs = append(duplicateIDs, id)
		}
	}
	if len(duplicateIDs) > 0 {
		sort.Strings(duplicateIDs)
		return newDuplicateRuleOrCategoryIDError(duplicateIDs)
	}
	return nil
}
