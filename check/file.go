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
	"fmt"

	checkv1beta1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/check/v1beta1"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// File is an invidual file that should be checked.
//
// Both the protoreflect FileDescriptor and the raw FileDescriptorProto interacves
// are provided.
//
// Files also have the property of being imports or non-imports.
type File interface {
	// FileDescriptor returns the protoreflect FileDescriptor representing this File.
	//
	// This will always contain SourceCodeInfo.
	FileDescriptor() protoreflect.FileDescriptor
	// FileDescriptorProto returns the FileDescriptorProto representing this File.
	//
	// This is not a copy - do not modify!
	FileDescriptorProto() *descriptorpb.FileDescriptorProto
	// IsImport returns true if the File is an import.
	//
	// An import is a file that is either:
	//
	//   - A Well-Known Type included from the compiler and imported by a targeted file.
	//   - A file that was included from a Buf module dependency and imported by a targeted file.
	//   - A file that was not targeted, but was imported by a targeted file.
	//
	// We use "import" as this matches with the protoc concept of --include_imports, however
	// import is a bit of an overloaded term.
	IsImport() bool

	toProto() *checkv1beta1.File

	isFile()
}

// FilesForProtoFiles returns a new slice of Files for the given checkv1beta1.Files.
func FilesForProtoFiles(protoFiles []*checkv1beta1.File) ([]File, error) {
	fileNameToProtoFile := make(map[string]*checkv1beta1.File, len(protoFiles))
	fileDescriptorProtos := make([]*descriptorpb.FileDescriptorProto, len(protoFiles))
	for i, protoFile := range protoFiles {
		fileDescriptorProto := protoFile.GetFileDescriptorProto()
		fileName := fileDescriptorProto.GetName()
		if _, ok := fileNameToProtoFile[fileName]; ok {
			//  This should have been validated via protovalidate.
			return nil, fmt.Errorf("duplicate file name: %q", fileName)
		}
		fileDescriptorProtos[i] = fileDescriptorProto
		fileNameToProtoFile[fileName] = protoFile
	}

	protoregistryFiles, err := protodesc.NewFiles(
		&descriptorpb.FileDescriptorSet{
			File: fileDescriptorProtos,
		},
	)
	if err != nil {
		return nil, err
	}

	files := make([]File, 0, len(protoFiles))
	protoregistryFiles.RangeFiles(
		func(fileDescriptor protoreflect.FileDescriptor) bool {
			protoFile, ok := fileNameToProtoFile[fileDescriptor.Path()]
			if !ok {
				// If the protoreflect API is sane, this should never happen.
				// However, the protoreflect API is not sane.
				err = fmt.Errorf("unknown file: %q", fileDescriptor.Path())
				return false
			}
			files = append(
				files,
				newFile(
					fileDescriptor,
					protoFile.GetFileDescriptorProto(),
					protoFile.GetIsImport(),
				),
			)
			return true
		},
	)
	if err != nil {
		return nil, err
	}
	if len(files) != len(protoFiles) {
		// If the protoreflect API is sane, this should never happen.
		// However, the protoreflect API is not sane.
		return nil, fmt.Errorf("expected %d files from protoregistry, got %d", len(protoFiles), len(files))
	}
	return files, nil
}

// *** PRIVATE ***

type file struct {
	fileDescriptor      protoreflect.FileDescriptor
	fileDescriptorProto *descriptorpb.FileDescriptorProto
	isImport            bool
}

func newFile(
	fileDescriptor protoreflect.FileDescriptor,
	fileDescriptorProto *descriptorpb.FileDescriptorProto,
	isImport bool,
) *file {
	return &file{
		fileDescriptor:      fileDescriptor,
		fileDescriptorProto: fileDescriptorProto,
		isImport:            isImport,
	}
}

func (f *file) FileDescriptor() protoreflect.FileDescriptor {
	return f.fileDescriptor
}

func (f *file) FileDescriptorProto() *descriptorpb.FileDescriptorProto {
	return f.fileDescriptorProto
}

func (f *file) IsImport() bool {
	return f.isImport
}

func (f *file) toProto() *checkv1beta1.File {
	return &checkv1beta1.File{
		FileDescriptorProto: f.fileDescriptorProto,
		IsImport:            f.isImport,
	}
}

func (*file) isFile() {}

func fileNameToFileForFiles(files []File) (map[string]File, error) {
	fileNameToFile := make(map[string]File, len(files))
	for _, file := range files {
		fileName := file.FileDescriptor().Path()
		if _, ok := fileNameToFile[fileName]; ok {
			return nil, fmt.Errorf("duplicate file name: %q", fileName)
		}
		fileNameToFile[fileName] = file
	}
	return fileNameToFile, nil
}
