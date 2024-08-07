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
	"slices"
	"sort"

	checkv1beta1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/check/v1beta1"
	"github.com/bufbuild/bufplugin-go/internal/pkg/xslices"
)

const checkRuleIDPageSize = 250

// Request is a request to a plugin to run checks.
type Request interface {
	// Files contains the files to check.
	//
	// Will never be nil or empty.
	Files() []File
	// AgainstFiles contains the files to check against, in the case of breaking change plugins.
	//
	// May be empty, including in the case where we did actually specify against files.
	// See the comment on the API for more details TODO this needs to be resolved.
	AgainstFiles() []File
	// Options contains any options passed to the plugin.
	//
	// Will never be nil, but may have no values.
	Options() Options
	// RuleIDs returns the specific IDs the of Rules to use.
	//
	// If empty, all Rules will be used.
	// The returned RuleIDs will be sorted.
	//
	// This may return more than 250 IDs; the underlying Client implemention is required to do
	// any necessary chunking.
	//
	// RuleHandlers can safely ignore this - the handling of RuleIDs will have already
	// been performed prior to the Request reaching the RuleHandler.
	RuleIDs() []string

	// toProtos converts the Request into one or more CheckRequests.
	//
	// If there are more than 250 Rule IDs, multiple CheckRequests will be produced by chunking up
	// the Rule IDs.
	toProtos() []*checkv1beta1.CheckRequest

	isRequest()
}

// NewRequest returns a new Request for the given Files.
//
// Files are always required. To set against Files or options, use
// WithAgainstFiles and WithOption.
func NewRequest(
	files []File,
	options ...RequestOption,
) (Request, error) {
	return newRequest(files, options...)
}

// RequestOption is an option for a new Request.
type RequestOption func(*requestOptions)

// WithAgainstFiles adds the given against Files to the Request.
func WithAgainstFiles(againstFiles []File) RequestOption {
	return func(requestOptions *requestOptions) {
		requestOptions.againstFiles = againstFiles
	}
}

// WithOption adds the given option to the Request.
func WithOption(key string, value []byte) RequestOption {
	// TODO: validate that duplicate keys are not specified
	return func(requestOptions *requestOptions) {
		requestOptions.keyToValue[key] = value
	}
}

// WithRuleIDs specifies that the given rule IDs should be used on the Request.
//
// Multiple calls to WithRuleIDs will result in the new rule IDs being appended.
func WithRuleIDs(ruleIDs ...string) RequestOption {
	return func(requestOptions *requestOptions) {
		for _, ruleID := range ruleIDs {
			requestOptions.ruleIDMap[ruleID] = struct{}{}
		}
	}
}

// RequestForProtoRequest returns a new Request for the given checkv1beta1.Request.
func RequestForProtoRequest(protoRequest *checkv1beta1.CheckRequest) (Request, error) {
	files, err := FilesForProtoFiles(protoRequest.GetFiles())
	if err != nil {
		return nil, err
	}
	againstFiles, err := FilesForProtoFiles(protoRequest.GetAgainstFiles())
	if err != nil {
		return nil, err
	}
	requestOptions := []RequestOption{
		WithAgainstFiles(againstFiles),
		WithRuleIDs(protoRequest.GetRuleIds()...),
	}
	for _, protoOption := range protoRequest.GetOptions() {
		requestOptions = append(requestOptions, WithOption(protoOption.GetKey(), protoOption.GetValue()))
	}
	return NewRequest(files, requestOptions...)
}

// *** PRIVATE ***

type request struct {
	files        []File
	againstFiles []File
	options      Options
	ruleIDs      []string
}

func newRequest(
	files []File,
	options ...RequestOption,
) (*request, error) {
	requestOptions := newRequestOptions()
	for _, option := range options {
		option(requestOptions)
	}
	optionsIface, err := newOptions(requestOptions.keyToValue)
	if err != nil {
		return nil, err
	}
	ruleIDs := make([]string, 0, len(requestOptions.ruleIDMap))
	for ruleID := range requestOptions.ruleIDMap {
		ruleIDs = append(ruleIDs, ruleID)
	}
	sort.Strings(ruleIDs)
	// TODO: need to validate Files and AgainstFiles per protovalidate specs
	return &request{
		files:        files,
		againstFiles: requestOptions.againstFiles,
		options:      optionsIface,
		ruleIDs:      ruleIDs,
	}, nil
}

func (r *request) Files() []File {
	return slices.Clone(r.files)
}

func (r *request) AgainstFiles() []File {
	return slices.Clone(r.againstFiles)
}

func (r *request) Options() Options {
	return r.options
}

func (r *request) RuleIDs() []string {
	return slices.Clone(r.ruleIDs)
}

func (r *request) toProtos() []*checkv1beta1.CheckRequest {
	if r == nil {
		return nil
	}
	protoFiles := xslices.Map(r.files, File.toProto)
	protoAgainstFiles := xslices.Map(r.againstFiles, File.toProto)
	protoOptions := r.options.toProto()
	if len(r.ruleIDs) == 0 {
		return []*checkv1beta1.CheckRequest{
			{
				Files:        protoFiles,
				AgainstFiles: protoAgainstFiles,
				Options:      protoOptions,
			},
		}
	}
	var checkRequests []*checkv1beta1.CheckRequest
	for i := 0; i < len(r.ruleIDs); i += checkRuleIDPageSize {
		start := i
		end := start + checkRuleIDPageSize
		if end > len(r.ruleIDs) {
			end = len(r.ruleIDs)
		}
		checkRequests = append(
			checkRequests,
			&checkv1beta1.CheckRequest{
				Files:        protoFiles,
				AgainstFiles: protoAgainstFiles,
				Options:      protoOptions,
				RuleIds:      r.ruleIDs[start:end],
			},
		)
	}
	return checkRequests
}

func (*request) isRequest() {}

type requestOptions struct {
	againstFiles []File
	keyToValue   map[string][]byte
	ruleIDMap    map[string]struct{}
}

func newRequestOptions() *requestOptions {
	return &requestOptions{
		keyToValue: make(map[string][]byte),
		ruleIDMap:  make(map[string]struct{}),
	}
}
