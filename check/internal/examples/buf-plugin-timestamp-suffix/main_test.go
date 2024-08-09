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
	"testing"

	"github.com/bufbuild/bufplugin-go/check/checktest"
)

func TestSimpleSuccess(t *testing.T) {
	t.Parallel()

	checktest.TestCase{
		Spec: spec,
		Files: &checktest.ProtoFileSpec{
			DirPaths:  []string{"testdata/simple_success"},
			FilePaths: []string{"simple.proto"},
		},
	}.Run(t)
}

func TestSimpleFailure(t *testing.T) {
	t.Parallel()

	checktest.TestCase{
		Spec: spec,
		Files: &checktest.ProtoFileSpec{
			DirPaths:  []string{"testdata/simple_failure"},
			FilePaths: []string{"simple.proto"},
		},
		// This linter only has a single Rule, so this has no effect in this
		// test, however this is how you scope a test to a single Rule.
		RuleIDs: []string{timestampSuffixID},
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				ID: timestampSuffixID,
				Location: &checktest.ExpectedLocation{
					FileName:    "simple.proto",
					StartLine:   8,
					StartColumn: 2,
					EndLine:     8,
					EndColumn:   50,
				},
			},
		},
	}.Run(t)
}
