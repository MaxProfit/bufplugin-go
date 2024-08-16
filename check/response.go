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

// Response is a response from a plugin for a check call.
type Response interface {
	// Annotations returns all of the Annotations.
	//
	// The returned annotations will be sorted.
	Annotations() []Annotation

	toProto() *checkv1beta1.CheckResponse

	isResponse()
}

// *** PRIVATE ***

type response struct {
	annotations []Annotation
}

func newResponse(annotations []Annotation) (*response, error) {
	sort.Slice(
		annotations,
		func(i int, j int) bool {
			return compareAnnotations(annotations[i], annotations[j]) < 0
		},
	)
	// TODO: validation? Leaving error for now
	return &response{
		annotations: annotations,
	}, nil
}

func (r *response) Annotations() []Annotation {
	return slices.Clone(r.annotations)
}

func (r *response) toProto() *checkv1beta1.CheckResponse {
	return &checkv1beta1.CheckResponse{
		Annotations: xslices.Map(r.annotations, Annotation.toProto),
	}
}

func (*response) isResponse() {}
