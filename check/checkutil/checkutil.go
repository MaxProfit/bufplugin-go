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
// This is a work in progress. The API may drastically changed, and this package
// is incomplete
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

// NewMessageRuleHandler returns a new RuleHandler that will call f for every message within Files.
//
// Imports are filtered. This is the standard case for lint rules.
func NewMessageRuleHandler(
	f func(context.Context, check.ResponseWriter, check.Request, protoreflect.MessageDescriptor) error,
) check.RuleHandler {
	return NewFileRuleHandler(
		func(
			ctx context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
			file check.File,
		) error {
			return forEachMessage(
				file.FileDescriptor().Messages(),
				func(messageDescriptor protoreflect.MessageDescriptor) error {
					return f(ctx, responseWriter, request, messageDescriptor)
				},
			)
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
	return NewMessageRuleHandler(
		func(
			ctx context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
			messageDescriptor protoreflect.MessageDescriptor,
		) error {
			fields := messageDescriptor.Fields()
			for i := range fields.Len() {
				if err := f(ctx, responseWriter, request, fields.Get(i)); err != nil {
					return err
				}
			}
			return nil
		},
	)
}

func forEachMessage(
	messages protoreflect.MessageDescriptors,
	f func(protoreflect.MessageDescriptor) error,
) error {
	for i := range messages.Len() {
		messageDescriptor := messages.Get(i)
		if err := f(messageDescriptor); err != nil {
			return err
		}
		// Nested messages.
		if err := forEachMessage(messageDescriptor.Messages(), f); err != nil {
			return err
		}
	}
	return nil
}
