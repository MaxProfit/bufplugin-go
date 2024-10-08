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

// Package timestampsuffix implements a simple plugin that checks that all
// google.protobuf.Timestamp fields end in a specific suffix".
//
// The default suffix is "_time", but this can be overridden with the
// "timestamp_suffix" option key.
//
// See cmd/buf-plugin-timestamp-suffix for the plugin main.
package timestampsuffix

import (
	"context"
	"strings"

	"github.com/bufbuild/bufplugin-go/check"
	"github.com/bufbuild/bufplugin-go/check/checkutil"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	// TimestampSuffixRuleID is the Rule ID of the timestamp suffix Rule.
	TimestampSuffixRuleID = "TIMESTAMP_SUFFIX"

	// TimestampSuffixOptionKey is the option key to override the default timestamp suffix.
	TimestampSuffixOptionKey = "timestamp_suffix"

	defaultTimestampSuffix = "_time"
)

var (
	// TimestampSuffixRuleSpec is the RuleSpec for the timestamp suffix Rule.
	TimestampSuffixRuleSpec = &check.RuleSpec{
		ID:        TimestampSuffixRuleID,
		IsDefault: true,
		Purpose:   `Checks that all google.protobuf.Timestamps end in a specific suffix (default is "_time").`,
		Type:      check.RuleTypeLint,
		Handler:   checkutil.NewFieldRuleHandler(checkTimestampSuffix),
	}

	// Spec is the Spec for the timestamp suffix plugin.
	Spec = &check.Spec{
		Rules: []*check.RuleSpec{
			TimestampSuffixRuleSpec,
		},
	}
)

func checkTimestampSuffix(
	_ context.Context,
	responseWriter check.ResponseWriter,
	request check.Request,
	fieldDescriptor protoreflect.FieldDescriptor,
) error {
	timestampSuffix := defaultTimestampSuffix
	timestampSuffixOptionValue, err := check.GetStringValue(request.Options(), TimestampSuffixOptionKey)
	if err != nil {
		return err
	}
	if timestampSuffixOptionValue != "" {
		timestampSuffix = timestampSuffixOptionValue
	}

	fieldDescriptorType := fieldDescriptor.Message()
	if fieldDescriptorType == nil {
		return nil
	}
	if string(fieldDescriptorType.FullName()) != "google.protobuf.Timestamp" {
		return nil
	}
	if !strings.HasSuffix(string(fieldDescriptor.Name()), timestampSuffix) {
		responseWriter.AddAnnotation(
			check.WithMessagef("Fields of type google.protobuf.Timestamp must end in %q but field name was %q.", timestampSuffix, string(fieldDescriptor.Name())),
			check.WithDescriptor(fieldDescriptor),
		)
	}
	return nil
}
