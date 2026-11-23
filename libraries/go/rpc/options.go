package rpc

import (
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	optionsv1 "github.com/zachnoble/laurel/rpc/generated/go/options/v1"
)

// IsPublic reports whether the procedure is marked `option (options.v1.public) = true`.
func IsPublic(spec connect.Spec) bool {
	method, ok := spec.Schema.(protoreflect.MethodDescriptor)
	if !ok {
		return false
	}
	public, _ := proto.GetExtension(method.Options(), optionsv1.E_Public).(bool)
	return public
}

// isInternal reports whether the procedure's service is marked
// `option (options.v1.internal) = true`.
func isInternal(spec connect.Spec) bool {
	method, ok := spec.Schema.(protoreflect.MethodDescriptor)
	if !ok {
		return false
	}
	service, ok := method.Parent().(protoreflect.ServiceDescriptor)
	if !ok {
		return false
	}
	internal, _ := proto.GetExtension(service.Options(), optionsv1.E_Internal).(bool)
	return internal
}
