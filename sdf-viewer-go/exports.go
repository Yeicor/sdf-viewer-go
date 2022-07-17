package sdf_viewer_go

import "C"
import (
	"unsafe"
)

// availableSDFs is the exported SDF hierarchy implementations.
var availableSDFs map[uint32]SDF

// SetRootSDF registers the root SDF, overriding any previous value.
func SetRootSDF(sdf SDF) {
	availableSDFs = map[uint32]SDF{}
	// Also register all children, recursively.
	registerSDFAndChildren(sdf)
}

func registerSDFAndChildren(sdf SDF) {
	//fmt.Printf("registerSDFAndChildren(%d)\n", sdf.ID())
	availableSDFs[sdf.ID()] = sdf
	for _, child := range sdf.Children() {
		registerSDFAndChildren(child)
	}
}

func getSDFOrPanic(sdfID uint32) SDF {
	if sdf, ok := availableSDFs[sdfID]; ok {
		return sdf
	}
	panic("SDF not found")
}

//export bounding_box
func BoundingBox(sdfID uint32) *[2][3]float32 {
	//fmt.Printf("-> BoundingBox(%d)\n", sdfID)
	minMax := getSDFOrPanic(sdfID).BoundingBox()
	//fmt.Printf("<- BoundingBox(%d) <- (%v, %v)\n", sdfID, minMax[0], minMax[1])
	return &minMax
}

//export sample
func Sample(sdfID uint32, point [3]float32, distanceOnly bool) *SDFSample {
	//fmt.Printf("BoundingBox(%d, %v, %v)\n", sdfID, point, distanceOnly)
	sample := getSDFOrPanic(sdfID).Sample(point, distanceOnly)
	//fmt.Printf("BoundingBox(%d, %v, %v) <- (%v)\n", sdfID, point, distanceOnly, sample)
	return &sample
}

//export children
func Children(sdfID uint32) *PointerLength {
	//fmt.Printf("-> Children(%d)\n", sdfID)
	children := getSDFOrPanic(sdfID).Children()
	if len(children) == 0 {
		res := PointerLength{Pointer: 0, Length: 0}
		return &res
	}
	childrenIDs := make([]uint32, len(children))
	for i, child := range children {
		childrenIDs[i] = child.ID()
	}
	res := PointerLength{Pointer: uintptr(unsafe.Pointer(&(childrenIDs[0]))), Length: uint32(uintptr(len(childrenIDs)) * unsafe.Sizeof(uint32(0)))}
	//fmt.Printf("<- Children(%d) <- (%v, %v)\n", sdfID, children.Pointer, children.Length)
	return &res
}

//export name
func Name(sdfID uint32) *PointerLength {
	//fmt.Printf("-> Children(%d)\n", sdfID)
	name := []byte(getSDFOrPanic(sdfID).Name())
	res := PointerLength{Pointer: uintptr(unsafe.Pointer(&(name[0]))), Length: uint32(uintptr(len(name)) * unsafe.Sizeof(uint8(0)))}
	//fmt.Printf("<- Children(%d) <- (%v, %v)\n", sdfID, children.Pointer, children.Length)
	return &res
}
