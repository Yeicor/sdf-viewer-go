//go:build tinygo

package sdf_viewer_go_auto

import (
	"reflect"
	"strconv"
	"unsafe"
)

// Reflection hacks specific to the tinygo compiler, as it does not support some features of the go standard library.
// HACK: TinyGo has very limited support for reflect (e.g. no Implements() or Interface()), which we must work around.

// HACK: Internal type
type Value struct {
	typecode uintptr
	value    unsafe.Pointer
	flags    uint8
}

// HACK: Internal function
//
//go:linkname valueInterfaceUnsafe reflect.valueInterfaceUnsafe
func valueInterfaceUnsafe(v reflect.Value) interface{}

//go:linkname composeInterface runtime.composeInterface
func composeInterface(uintptr, unsafe.Pointer) interface{}

// interfaceAndImplementsHint returns the interface{} behind a reflect.Value, and may also return if it implements a
// reflect.Type (of an interface). Optimization: it will only return the interface if the type cannot be checked or it was ok.
func interfaceAndImplementsHint(value reflect.Value, kind reflect.Type) (interface{}, *bool) {
	// Unsafely convert to interface (internal api)
	hackedValueIface := valueInterfaceUnsafe(value)
	//log.Printf("hackedValueIface: %#+v", hackedValueIface) // This line breaks everything... (weak hack)
	return hackedValueIface, nil /* no way to know the hint */
}

func nameOfType(val reflect.Value, ptr uintptr) string {
	return val.Type().String() + "(" + strconv.Itoa(int(ptr)) + ")"
}
