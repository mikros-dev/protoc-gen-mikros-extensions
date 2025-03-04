package extensions

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	descriptor "google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type DescriptorObject interface {
	GetOptions() *descriptor.MessageOptions
}

// HasExtension checks if a protobuf object is using a specific type of
// extension. Usually, this function should be used by addons to check if
// an object has an extension before trying to retrieve it.
func HasExtension[T DescriptorObject](msg T, options protoreflect.ExtensionTypeDescriptor) bool {
	if msg.GetOptions() == nil {
		return false
	}

	return msg.GetOptions().ProtoReflect().Has(options)
}

// RetrieveExtension extracts an extension from a protobuf message and
// fills target with it. It returns nil if the message does not have the
// extension.
func RetrieveExtension[T DescriptorObject](msg T, options protoreflect.ExtensionTypeDescriptor, target proto.Message) error {
	if msg.GetOptions() == nil {
		return nil
	}

	value := msg.GetOptions().ProtoReflect().Get(options)
	if !value.IsValid() {
		return nil
	}

	dynMsg, ok := value.Message().Interface().(*dynamicpb.Message)
	if !ok {
		return nil
	}

	data, err := proto.Marshal(dynMsg)
	if err != nil {
		return err
	}

	if err := proto.Unmarshal(data, target); err != nil {
		return err
	}

	return nil
}
