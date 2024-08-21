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
	"errors"
	"fmt"

	"github.com/bufbuild/protovalidate-go"
)

// CategorySpec is the spec for a Category.
//
// It is used to construct a Category on the server-side (i.e. within the plugin). It specifies the
// ID, purpose,  and a CategoryHandler to actually run the Category logic.
//
// Generally, these are provided to Main. This library will handle Check and ListCategories calls
// based on the provided CategorySpecs.
type CategorySpec struct {
	// Required.
	ID string
	// Required.
	Purpose        string
	Deprecated     bool
	ReplacementIDs []string
}

// *** PRIVATE ***

// Assumes that the CategorySpec is validated.
func categorySpecToCategory(categorySpec *CategorySpec) (Category, error) {
	return newCategory(
		categorySpec.ID,
		categorySpec.Purpose,
		categorySpec.Deprecated,
		categorySpec.ReplacementIDs,
	), nil
}

func validateCategorySpec(_ *protovalidate.Validator, categorySpec *CategorySpec) error {
	if categorySpec.ID == "" {
		return errors.New("CategorySpec.ID is empty")
	}
	if categorySpec.Purpose == "" {
		return fmt.Errorf("CategorySpec.Purpose is not set for ID %q", categorySpec.ID)
	}
	if len(categorySpec.ReplacementIDs) > 0 && !categorySpec.Deprecated {
		return fmt.Errorf("CategorySpec.ReplacementIDs had values %v but Deprecated was false", categorySpec.ReplacementIDs)
	}
	// We do this on the server-side only, this shouldn't be used client-side.
	// TODO: This isn't working
	return nil
	// return validator.Validate(categorySpecToCategory(categorySpec, emptyOptions).toProto())
}
