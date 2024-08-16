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
	"strings"
)

type duplicateRuleError struct {
	duplicateRuleIDs []string
}

func newDuplicateRuleError(duplicateRuleIDs []string) *duplicateRuleError {
	return &duplicateRuleError{
		duplicateRuleIDs: duplicateRuleIDs,
	}
}

func (d *duplicateRuleError) Error() string {
	if d == nil {
		return ""
	}
	if len(d.duplicateRuleIDs) == 0 {
		return ""
	}
	var sb strings.Builder
	_, _ = sb.WriteString("duplicate rule IDs: ")
	_, _ = sb.WriteString(strings.Join(d.duplicateRuleIDs, ", "))
	return sb.String()
}

type unexpectedOptionValueTypeError struct {
	key      string
	expected any
	actual   any
}

func newUnexpectedOptionValueError(key string, expected any, actual any) *unexpectedOptionValueTypeError {
	return &unexpectedOptionValueTypeError{
		key:      key,
		expected: expected,
		actual:   actual,
	}
}

func (u *unexpectedOptionValueTypeError) Error() string {
	if u == nil {
		return ""
	}
	var sb strings.Builder
	_, _ = sb.WriteString(`unexpected type for option value "`)
	_, _ = sb.WriteString(u.key)
	_, _ = sb.WriteString(fmt.Sprintf(`": expected %T, got %T`, u.expected, u.actual))
	return sb.String()
}
