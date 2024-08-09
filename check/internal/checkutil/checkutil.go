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

// Package checkutil implements helpers for the check package.
//
// This may eventually be made external to help others write plugins. For now, this
// is in early development and we do not want to commit to the API whatsoever.
// This only currently covers methods we need for examples and testing. If we
// expose this further, it will be more complete.
package checkutil

import (
	"context"

	"github.com/bufbuild/bufplugin-go/check"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// NewFileRuleHandler returns a new RuleHandler that will call f for every file within Files.
//
// Imports are filtered. This is the standard case for lint rules.
func NewFileRuleHandler(
	f func(context.Context, check.ResponseWriter, check.Request, check.File) error,
) check.RuleHandler {
	return check.RuleHandlerFunc(
		func(
			ctx context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
		) error {
			for _, file := range request.Files() {
				if file.IsImport() {
					continue
				}
				if err := f(ctx, responseWriter, request, file); err != nil {
					return err
				}
			}
			return nil
		},
	)
}

// NewFieldRuleHandler returns a new RuleHandler that will call f for every field in
// the messages within Files.
//
// Imports are filtered. This is the standard case for lint rules.
func NewFieldRuleHandler(
	f func(context.Context, check.ResponseWriter, check.Request, protoreflect.FieldDescriptor) error,
) check.RuleHandler {
	return NewFileRuleHandler(
		func(
			ctx context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
			file check.File,
		) error {
			return forEachField(
				file.FileDescriptor(),
				func(fieldDescriptor protoreflect.FieldDescriptor) error {
					return f(ctx, responseWriter, request, fieldDescriptor)
				},
			)
		},
	)
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
