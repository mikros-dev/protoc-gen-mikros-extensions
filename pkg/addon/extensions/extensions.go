package extensions

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type DescriptorObject interface {
	comparable
	ProtoReflect() protoreflect.Message
}

// HasExtension checks if a protobuf object is using a specific type of
// extension. Usually, this function should be used by addons to check if
// an object has an extension before trying to retrieve it.
// Remember to pass the object returned by the GetOption() call of the
// original object, otherwise the function will check the wrong one.
func HasExtension[T DescriptorObject](msg T, options protoreflect.ExtensionTypeDescriptor) bool {
	var zero T
	if msg == zero {
		return false
	}

	return msg.ProtoReflect().Has(options)
}

// RetrieveExtension extracts an extension from a protobuf message and
// fills target with it. It returns nil if the message does not have the
// extension.
func RetrieveExtension[T DescriptorObject](msg T, options protoreflect.ExtensionTypeDescriptor, target proto.Message) error {
	var zero T
	if msg == zero {
		return nil
	}

	value := msg.ProtoReflect().Get(options)
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
