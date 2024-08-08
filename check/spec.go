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
	Rules []*RuleSpec
}

// *** PRIVATE ***

func validateSpec(validator *protovalidate.Validator, spec *Spec) error {
	if len(spec.Rules) == 0 {
		return errors.New("Spec.Rules is empty")
	}
	for _, ruleSpec := range spec.Rules {
		if err := validateRuleSpec(validator, ruleSpec); err != nil {
			return err
		}
	}
	return nil
}
