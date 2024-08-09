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

// Package main implements a simple plugin that checks that all google.protobuf.Timestamp
// fields end in "_time".
package main

import (
	"context"
	"strings"

	"github.com/bufbuild/bufplugin-go/check"
	"github.com/bufbuild/bufplugin-go/check/internal/checkutil"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	timestampSuffixID = "TIMESTAMP_SUFFIX"
)

var (
	timestampSuffixRuleSpec = &check.RuleSpec{
		ID:      timestampSuffixID,
		Purpose: "Checks that all google.protobuf.Timestamps end in _time.",
		Type:    check.RuleTypeLint,
		Handler: checkutil.NewFieldRuleHandler(checkTimestampSuffix),
	}

	spec = &check.Spec{
		Rules: []*check.RuleSpec{
			timestampSuffixRuleSpec,
		},
	}
)

func main() {
	check.Main(spec)
}

func checkTimestampSuffix(
	_ context.Context,
	responseWriter check.ResponseWriter,
	_ check.Request,
	fieldDescriptor protoreflect.FieldDescriptor,
) error {
	fieldDescriptorType := fieldDescriptor.Message()
	if fieldDescriptorType == nil {
		return nil
	}
	if string(fieldDescriptorType.FullName()) != "google.protobuf.Timestamp" {
		return nil
	}
	if !strings.HasSuffix(string(fieldDescriptor.Name()), "_time") {
		responseWriter.AddAnnotation(
			check.WithMessagef("Fields of type google.protobuf.Timestamp must end in _time but field name was %q.", string(fieldDescriptor.Name())),
			check.WithDescriptor(fieldDescriptor),
		)
	}
	return nil
}
