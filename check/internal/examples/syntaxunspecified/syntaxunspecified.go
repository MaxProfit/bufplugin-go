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

// Package syntaxunspecified implements a simple plugin that checks that syntax
// is unspecified in every file.
//
// This is just demonstrating the additional functionality that check.Files have
// over FileDescriptorProtos. We have no idea why you'd actually want to lint this.
//
// See cmd/buf-plugin-syntax-unspecified for the plugin main.
package syntaxunspecified

import (
	"context"

	"github.com/bufbuild/bufplugin-go/check"
	"github.com/bufbuild/bufplugin-go/check/internal/checkutil"
)

// SyntaxUnspecifiedRuleID is the Rule ID of the syntax unspecified Rule.
const SyntaxUnspecifiedRuleID = "SYNTAX_UNSPECIFIED"

var (
	// SyntaxUnspecifiedRuleSpec is the RuleSpec for the syntax unspecified Rule.
	SyntaxUnspecifiedRuleSpec = &check.RuleSpec{
		ID:        SyntaxUnspecifiedRuleID,
		IsDefault: true,
		Purpose:   "Checks that syntax is never specified.",
		Type:      check.RuleTypeLint,
		Handler:   checkutil.NewFileRuleHandler(checkSyntaxUnspecified),
	}

	// Spec is the Spec for the syntax unspecified plugin.
	Spec = &check.Spec{
		Rules: []*check.RuleSpec{
			SyntaxUnspecifiedRuleSpec,
		},
	}
)

func checkSyntaxUnspecified(
	_ context.Context,
	responseWriter check.ResponseWriter,
	_ check.Request,
	file check.File,
) error {
	if !file.IsSyntaxUnspecified() {
		syntax := file.FileDescriptorProto().GetSyntax()
		responseWriter.AddAnnotation(
			check.WithMessagef("Syntax should not be specified but was %q.", syntax),
			check.WithDescriptor(file.FileDescriptor()),
		)
	}
	return nil
}
