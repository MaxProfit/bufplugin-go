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
	"github.com/bufbuild/pluginrpc-go"
)

// Main is the main entrypoint for a plugin that implements the given RuleSpecs.
//
// A plugin just needs to provide RuleSpecs, and then call this function within main.
//
//	func main() {
//		check.Main(
//			[]*check.RuleSpec{
//				{
//					ID:      "TIMESTAMP_SUFFIX",
//					Purpose: check.NopPurpose("Checks that all google.protobuf.Timestamps end in _time."),
//					Type:    check.RuleTypeLint,
//					Handler: check.NopRuleHandler(check.RuleHandlerFunc(handleTimestampSuffix)),
//				},
//			},
//		)
//	}
func Main(ruleSpecs []*RuleSpec, _ ...MainOption) {
	pluginrpc.Main(
		func() (pluginrpc.Server, error) {
			checkServiceHandler, err := newCheckServiceHandler(ruleSpecs)
			if err != nil {
				return nil, err
			}
			return newCheckServer(checkServiceHandler)
		},
	)
}

// MainOption is an option for Main.
type MainOption func(*mainOptions)

// *** PRIVATE ***

type mainOptions struct{}
