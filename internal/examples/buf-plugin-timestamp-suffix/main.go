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

package main

import (
	"context"
	"strings"

	"github.com/bufbuild/bufplugin-go/check"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	timestampSuffixID = "TIMESTAMP_SUFFIX"
)

var (
	timestampSuffixRuleSpec = &check.RuleSpec{
		ID:      timestampSuffixID,
		Purpose: check.NopPurpose("Checks that all google.protobuf.Timestamps end in _time."),
		Type:    check.RuleTypeLint,
		Handler: check.NopRuleHandler(check.RuleHandlerFunc(handleTimestampSuffix)),
	}

	ruleSpecs = []*check.RuleSpec{
		timestampSuffixRuleSpec,
	}
)

func main() {
	check.Main(ruleSpecs)
}

func handleTimestampSuffix(
	_ context.Context,
	responseWriter check.ResponseWriter,
	request check.Request,
) error {
	for _, file := range request.Files() {
		if file.IsImport() {
			continue
		}
		if err := forEachField(
			file.FileDescriptor(),
			func(fieldDescriptor protoreflect.FieldDescriptor) error {
				return handleFieldDescriptor(responseWriter, fieldDescriptor)
			},
		); err != nil {
			return err
		}
	}
	return nil
}

func handleFieldDescriptor(
	responseWriter check.ResponseWriter,
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
			check.WithMessage("Fields of type google.protobuf.Timestamp must end in _time."),
			check.WithDescriptor(fieldDescriptor),
		)
	}
	return nil
}

func forEachField(
	fileDescriptor protoreflect.FileDescriptor,
	f func(protoreflect.FieldDescriptor) error,
) error {
	messages := fileDescriptor.Messages()
	for i := 0; i < messages.Len(); i++ {
		if err := forEachFieldInMessage(messages.Get(i), f); err != nil {
			return err
		}
	}
	return nil
}

func forEachFieldInMessage(
	messageDescriptor protoreflect.MessageDescriptor,
	f func(protoreflect.FieldDescriptor) error,
) error {
	fields := messageDescriptor.Fields()
	for i := 0; i < fields.Len(); i++ {
		if err := f(fields.Get(i)); err != nil {
			return err
		}
	}
	messages := messageDescriptor.Messages()
	for i := 0; i < messages.Len(); i++ {
		if err := forEachFieldInMessage(messages.Get(i), f); err != nil {
			return err
		}
	}
	return nil
}
