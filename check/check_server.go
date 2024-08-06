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
	"github.com/bufbuild/bufplugin-go/internal/gen/buf/plugin/check/v1beta1/v1beta1pluginrpc"
	"github.com/bufbuild/pluginrpc-go"
)

// *** PRIVATE ***

func newCheckServer(checkServiceHandler v1beta1pluginrpc.CheckServiceHandler) (pluginrpc.Server, error) {
	spec, err := v1beta1pluginrpc.CheckServiceSpecBuilder{
		Check:     []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("check")},
		ListRules: []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("list-rules")},
	}.Build()
	if err != nil {
		return nil, err
	}
	serverRegistrar := pluginrpc.NewServerRegistrar()
	checkServiceServer := v1beta1pluginrpc.NewCheckServiceServer(pluginrpc.NewHandler(), checkServiceHandler)
	v1beta1pluginrpc.RegisterCheckServiceServer(serverRegistrar, checkServiceServer)
	return pluginrpc.NewServer(spec, serverRegistrar)
}
