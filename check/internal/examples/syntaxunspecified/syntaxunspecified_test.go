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

package syntaxunspecified

import (
	"testing"

	"github.com/bufbuild/bufplugin-go/check/checktest"
)

func TestSimpleSuccess(t *testing.T) {
	t.Parallel()

	checktest.TestCase{
		Spec: Spec,
		Files: &checktest.ProtoFileSpec{
			DirPaths:  []string{"testdata/simple_success"},
			FilePaths: []string{"simple.proto"},
		},
	}.Run(t)
}

func TestSimpleFailure(t *testing.T) {
	t.Parallel()

	checktest.TestCase{
		Spec: Spec,
		Files: &checktest.ProtoFileSpec{
			DirPaths:  []string{"testdata/simple_failure"},
			FilePaths: []string{"simple.proto"},
		},
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				ID: SyntaxUnspecifiedRuleID,
				Location: &checktest.ExpectedLocation{
					FileName: "simple.proto",
				},
			},
		},
	}.Run(t)
}
