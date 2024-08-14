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

package check_test

import (
	"context"
	"testing"

	"github.com/bufbuild/bufplugin-go/check"
	"github.com/bufbuild/bufplugin-go/check/checktest"
	"github.com/bufbuild/bufplugin-go/check/internal/examples/fieldlowersnakecase"
	"github.com/bufbuild/bufplugin-go/check/internal/examples/timestampsuffix"
	"github.com/bufbuild/bufplugin-go/internal/pkg/xslices"
	"github.com/stretchr/testify/require"
)

func TestMultiClientSimple(t *testing.T) {
	t.Parallel()

	testMultiClientSimple(t, false)
}

func TestMultiClientSimpleCacheRules(t *testing.T) {
	t.Parallel()

	testMultiClientSimple(t, true)
}

func testMultiClientSimple(t *testing.T, cacheRules bool) {
	ctx := context.Background()

	requestSpec := &checktest.RequestSpec{
		Files: &checktest.ProtoFileSpec{
			DirPaths:  []string{"testdata/multi_client_simple"},
			FilePaths: []string{"simple.proto"},
		},
	}
	request, err := requestSpec.ToRequest(ctx)
	require.NoError(t, err)

	var clientOptions []check.ClientOption
	if cacheRules {
		clientOptions = append(clientOptions, check.ClientWithCacheRules())
	}
	fieldLowerSnakeCaseClient, err := check.NewClientForSpec(fieldlowersnakecase.Spec, clientOptions...)
	require.NoError(t, err)
	timestampSuffixClient, err := check.NewClientForSpec(timestampsuffix.Spec, clientOptions...)
	require.NoError(t, err)
	multiClient := check.NewMultiClient(
		[]check.Client{
			fieldLowerSnakeCaseClient,
			timestampSuffixClient,
		},
	)

	rules, err := multiClient.ListRules(ctx)
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			fieldlowersnakecase.FieldLowerSnakeCaseRuleID,
			timestampsuffix.TimestampSuffixRuleID,
		},
		xslices.Map(rules, check.Rule.ID),
	)
	response, err := multiClient.Check(ctx, request)
	require.NoError(t, err)
	checktest.AssertAnnotationsEqual(
		t,
		[]checktest.ExpectedAnnotation{
			{
				ID: fieldlowersnakecase.FieldLowerSnakeCaseRuleID,
				Location: &checktest.ExpectedLocation{
					FileName:    "simple.proto",
					StartLine:   10,
					StartColumn: 2,
					EndLine:     10,
					EndColumn:   23,
				},
			},
			{
				ID: timestampsuffix.TimestampSuffixRuleID,
				Location: &checktest.ExpectedLocation{
					FileName:    "simple.proto",
					StartLine:   9,
					StartColumn: 2,
					EndLine:     9,
					EndColumn:   50,
				},
			},
		},
		response.Annotations(),
	)
}
