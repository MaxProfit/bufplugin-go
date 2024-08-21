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

type duplicateRuleIDError struct {
	duplicateIDs []string
}

func newDuplicateRuleIDError(duplicateIDs []string) *duplicateRuleIDError {
	return &duplicateRuleIDError{
		duplicateIDs: duplicateIDs,
	}
}

func (r *duplicateRuleIDError) Error() string {
	if r == nil {
		return ""
	}
	if len(r.duplicateIDs) == 0 {
		return ""
	}
	var sb strings.Builder
	_, _ = sb.WriteString("duplicate rule IDs: ")
	_, _ = sb.WriteString(strings.Join(r.duplicateIDs, ", "))
	return sb.String()
}

type duplicateCategoryIDError struct {
	duplicateIDs []string
}

func newDuplicateCategoryIDError(duplicateIDs []string) *duplicateCategoryIDError {
	return &duplicateCategoryIDError{
		duplicateIDs: duplicateIDs,
	}
}

func (c *duplicateCategoryIDError) Error() string {
	if c == nil {
		return ""
	}
	if len(c.duplicateIDs) == 0 {
		return ""
	}
	var sb strings.Builder
	_, _ = sb.WriteString("duplicate category IDs: ")
	_, _ = sb.WriteString(strings.Join(c.duplicateIDs, ", "))
	return sb.String()
}

type duplicateRuleOrCategoryIDError struct {
	duplicateIDs []string
}

func newDuplicateRuleOrCategoryIDError(duplicateIDs []string) *duplicateRuleOrCategoryIDError {
	return &duplicateRuleOrCategoryIDError{
		duplicateIDs: duplicateIDs,
	}
}

func (o *duplicateRuleOrCategoryIDError) Error() string {
	if o == nil {
		return ""
	}
	if len(o.duplicateIDs) == 0 {
		return ""
	}
	var sb strings.Builder
	_, _ = sb.WriteString("duplicate rule or category IDs: ")
	_, _ = sb.WriteString(strings.Join(o.duplicateIDs, ", "))
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
