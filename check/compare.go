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

// joinCompares returns the first compare value that is non-zero, or zero otherwise.
func joinCompares(compares ...int) int {
	for _, compare := range compares {
		if compare != 0 {
			return compare
		}
	}
	return 0
}

func intCompare(one int, two int) int {
	if one < two {
		return -1
	}
	if one > two {
		return 1
	}
	return 0
}

func nilCompare(one any, two any) int {
	if one == nil && two != nil {
		return -1
	}
	if one != nil && two == nil {
		return 1
	}
	return 0
}
