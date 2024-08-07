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
	"slices"

	checkv1beta1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/check/v1beta1"
)

// Options are key/values that can control the behavior of a RuleHandler,
// and can control the value of the Purpose string of the Rule.
//
// For example, if you had a Rule that checked that the suffix of all Services was "API",
// you may want an option with key "service_suffix" that can override the suffix "API" to
// another suffix such as "Service". This would result in the behavior of the check changing,
// as well as result in the Purpose string potentially changing to specify that the
// expected suffix is "Service" instead of "API".
type Options interface {
	// Get gets the option value for the given key.
	//
	// It is not possible to set an option with an empty value. If you want to specify
	// set or not set, use a sentinel value such as "true" or "1".
	//
	// The value is a []byte to allow plugins to encode whatever information they wish, however
	// plugin authors are responsible for parsing. For example, users may want to specify a number,
	// in which case the plugin would be responsible for parsing this number.
	//
	// The key must have at least four characters.
	// The key must start and end with a lowercase letter from a-z, and only consist
	// of lowercase letters from a-z and underscores.
	Get(key string) []byte

	toProto() []*checkv1beta1.Option

	isOption()
}

// *** PRIVATE ***

type options struct {
	keyToValue map[string][]byte
}

func newOptions(keyToValue map[string][]byte) (*options, error) {
	if err := validateKeyToValue(keyToValue); err != nil {
		return nil, err
	}
	return newOptionsNoValidate(keyToValue), nil
}

func newOptionsNoValidate(keyToValue map[string][]byte) *options {
	if keyToValue == nil {
		keyToValue = make(map[string][]byte)
	}
	return &options{
		keyToValue: keyToValue,
	}
}

func (o *options) Get(key string) []byte {
	// Might be unnecessary, check docs for slices.Clone if nil input returns nil output.
	value, ok := o.keyToValue[key]
	if ok {
		return slices.Clone(value)
	}
	return nil
}

func (o *options) toProto() []*checkv1beta1.Option {
	if o == nil {
		return nil
	}
	protoOptions := make([]*checkv1beta1.Option, 0, len(o.keyToValue))
	for key, value := range o.keyToValue {
		// Assuming that we've validated that no values are empty.
		protoOptions = append(
			protoOptions,
			&checkv1beta1.Option{
				Key:   key,
				Value: value,
			},
		)
	}
	return protoOptions
}

func (*options) isOption() {}

func validateKeyToValue(keyToValue map[string][]byte) error {
	for key, value := range keyToValue {
		// This should all be validated via protovalidate, and the below doesn't
		// even encapsulate all the validation.
		if len(key) == 0 {
			return errors.New("option key is empty")
		}
		if len(value) == 0 {
			return fmt.Errorf("option value is empty for key %q", key)
		}
	}
	return nil
}
