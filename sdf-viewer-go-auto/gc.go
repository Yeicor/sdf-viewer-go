//go:build !tinygo

package sdf_viewer_go_auto

import (
	"reflect"
	"strconv"
	"unsafe"
)

// Generic reflect implementation for the standard go compilers

// interfaceAndImplementsHint returns the interface{} behind a reflect.Value, and may also return if it implements a
// reflect.Type (of an interface). Optimization: it will only return the interface if the type cannot be checked or it was ok.
func interfaceAndImplementsHint(value reflect.Value, kind reflect.Type) (interface{}, *bool) {
	hint := false
	if value.Type().Implements(kind) {
		value = makeInterfaceWorkHack(value)
		return value.Interface(), &hint
	}
	return nil, nil
}

func makeInterfaceWorkHack(value reflect.Value) reflect.Value {
	if !value.CanInterface() {
		// HACK: Read-only access to unexported value (Interface() is not allowed due to possible write operations?)
		value = getUnexportedField(value, value.Addr().UnsafePointer())
	}
	return value
}

func getUnexportedField(field reflect.Value, unsafeAddr unsafe.Pointer) reflect.Value {
	return reflect.NewAt(field.Type(), unsafeAddr).Elem()
}

var nameToNextIndex = map[string]int{}

func nameOfType(kind reflect.Value, _ uintptr) string {
	tp := kind.Type().String()
	// Unique name of the type (avoid collisions)
	if val, ok := nameToNextIndex[tp]; ok {
		nameToNextIndex[tp] = val + 1
		tp += "_" + strconv.Itoa(val)
	} else {
		nameToNextIndex[tp] = 1
	}
	return tp
}
