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
	"sync"

	checkv1beta1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/check/v1beta1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Location is a reference to a File or to a location within a File.
//
// A Location always has a file name.
type Location interface {
	// FileName returns the name of the File.
	//
	// Always present.
	//
	// This matches the name field within the corresponding FileDescriptorProto.
	FileName() string

	// StartLine returns the zero-indexed start line, if known.
	StartLine() int
	// StartColumn returns the zero-indexed start column, if known.
	StartColumn() int
	// EndLine returns the zero-indexed end line, if known.
	EndLine() int
	// EndColumn returns the zero-indexed end column, if known.
	EndColumn() int
	// LeadingComments returns any leading comments, if known.
	LeadingComments() string
	// TrailingComments returns any trailing comments, if known.
	TrailingComments() string
	// LeadingDetachedComments returns any leading detached comments, if known.
	LeadingDetachedComments() []string

	// SourcePath returns the path within the FileDescriptorProto of the location.
	SourcePath() protoreflect.SourcePath

	toProto() *checkv1beta1.Location

	isLocation()
}

// *** PRIVATE ***

func locationForDescriptor(descriptor protoreflect.Descriptor) Location {
	if fileDescriptor := descriptor.ParentFile(); fileDescriptor != nil {
		return newLocation(
			fileDescriptor.Path(),
			func() protoreflect.SourceLocation { return sourceLocationForDescriptor(descriptor) },
		)
	}
	return nil
}

func locationForFileNameAndSourceLocation(fileName string, sourceLocation protoreflect.SourceLocation) Location {
	return newLocation(
		fileName,
		func() protoreflect.SourceLocation { return sourceLocation },
	)
}

type location struct {
	fileName          string
	getSourceLocation func() protoreflect.SourceLocation
}

func newLocation(
	fileName string,
	getSourceLocation func() protoreflect.SourceLocation,
) *location {
	return &location{
		fileName:          fileName,
		getSourceLocation: sync.OnceValue(getSourceLocation),
	}
}

func (l *location) FileName() string {
	return l.fileName
}

func (l *location) StartLine() int {
	return l.getSourceLocation().StartLine
}

func (l *location) StartColumn() int {
	return l.getSourceLocation().StartColumn
}

func (l *location) EndLine() int {
	return l.getSourceLocation().EndLine
}

func (l *location) EndColumn() int {
	return l.getSourceLocation().EndColumn
}

func (l *location) LeadingComments() string {
	return l.getSourceLocation().LeadingComments
}

func (l *location) TrailingComments() string {
	return l.getSourceLocation().TrailingComments
}

func (l *location) LeadingDetachedComments() []string {
	return slices.Clone(l.getSourceLocation().LeadingDetachedComments)
}

func (l *location) SourcePath() protoreflect.SourcePath {
	return slices.Clone(l.getSourceLocation().Path)
}

func (l *location) toProto() *checkv1beta1.Location {
	if l == nil {
		return nil
	}
	return &checkv1beta1.Location{
		FileName:   l.FileName(),
		SourcePath: l.getSourceLocation().Path,
	}
}

func (*location) isLocation() {}

func sourceLocationForDescriptor(descriptor protoreflect.Descriptor) protoreflect.SourceLocation {
	if descriptor == nil {
		return protoreflect.SourceLocation{}
	}
	if fileDescriptor := descriptor.ParentFile(); fileDescriptor != nil {
		return fileDescriptor.SourceLocations().ByDescriptor(descriptor)
	}
	return protoreflect.SourceLocation{}
}
