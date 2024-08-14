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

// Package fieldlowersnakecase implements a simple plugin that checks that all
// field names are lower_snake_case.
//
// Note that buf implements this check by default, but this is just for example.
//
// See cmd/buf-plugin-field-lower-snake-case for the plugin main.
package fieldlowersnakecase

import (
	"context"
	"strings"
	"unicode"

	"github.com/bufbuild/bufplugin-go/check"
	"github.com/bufbuild/bufplugin-go/check/internal/checkutil"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// FieldLowerSnakeCaseRuleID is the Rule ID of the timestamp suffix Rule.
const FieldLowerSnakeCaseRuleID = "FIELD_LOWER_SNAKE_CASE"

var (
	// FieldLowerSnakeCaseRuleSpec is the RuleSpec for the timestamp suffix Rule.
	FieldLowerSnakeCaseRuleSpec = &check.RuleSpec{
		ID:      FieldLowerSnakeCaseRuleID,
		Purpose: "Checks that all field names are lower_snake_case.",
		Type:    check.RuleTypeLint,
		Handler: checkutil.NewFieldRuleHandler(checkFieldLowerSnakeCase),
	}

	// Spec is the Spec for the timestamp suffix plugin.
	Spec = &check.Spec{
		Rules: []*check.RuleSpec{
			FieldLowerSnakeCaseRuleSpec,
		},
	}
)

func checkFieldLowerSnakeCase(
	_ context.Context,
	responseWriter check.ResponseWriter,
	_ check.Request,
	fieldDescriptor protoreflect.FieldDescriptor,
) error {
	fieldName := string(fieldDescriptor.Name())
	fieldNameToLowerSnakeCase := toLowerSnakeCase(fieldName)
	if fieldName != fieldNameToLowerSnakeCase {
		responseWriter.AddAnnotation(
			check.WithMessagef("Field name %q should be lower_snake_case, such as %q.", fieldName, fieldNameToLowerSnakeCase),
			check.WithDescriptor(fieldDescriptor),
		)
	}
	return nil
}

func toLowerSnakeCase(s string) string {
	return strings.ToLower(toSnakeCase(s))
}

func toSnakeCase(s string) string {
	output := ""
	s = strings.TrimFunc(s, isDelimiter)
	for i, c := range s {
		if isDelimiter(c) {
			c = '_'
		}
		switch {
		case i == 0:
			output += string(c)
		case isSnakeCaseNewWord(c, false) &&
			output[len(output)-1] != '_' &&
			((i < len(s)-1 && !isSnakeCaseNewWord(rune(s[i+1]), true) && !isDelimiter(rune(s[i+1]))) ||
				(unicode.IsLower(rune(s[i-1])))):
			output += "_" + string(c)
		case !(isDelimiter(c) && output[len(output)-1] == '_'):
			output += string(c)
		}
	}
	return output
}

func isSnakeCaseNewWord(r rune, newWordOnDigits bool) bool {
	if newWordOnDigits {
		return unicode.IsUpper(r) || unicode.IsDigit(r)
	}
	return unicode.IsUpper(r)
}

func isDelimiter(r rune) bool {
	return r == '.' || r == '-' || r == '_' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
}
