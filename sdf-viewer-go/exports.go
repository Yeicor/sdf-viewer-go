package sdf_viewer_go

import "C"
import (
	"math"
	"unsafe"
)

// availableSDFs is the exported SDF hierarchy implementations.
var availableSDFs map[uint32]SDF

// SetRootSDF registers the root SDF, overriding any previous Value.
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
func boundingBox(sdfID uint32) *[2][3]float32 {
	//fmt.Printf("-> BoundingBox(%d)\n", sdfID)
	minMax := getSDFOrPanic(sdfID).BoundingBox()
	//fmt.Printf("<- BoundingBox(%d) <- (%v, %v)\n", sdfID, minMax[0], minMax[1])
	return &minMax
}

//export sample
func sample(sdfID uint32, point [3]float32, distanceOnly bool) *SDFSample {
	//fmt.Printf("Sample(%d, %v, %v)\n", sdfID, point, distanceOnly)
	sample := getSDFOrPanic(sdfID).Sample(point, distanceOnly)
	//fmt.Printf("Sample(%d, %v, %v) <- (%v)\n", sdfID, point, distanceOnly, sample)
	return &sample
}

//export children
func children(sdfID uint32) *pointerLength {
	//fmt.Printf("-> Children(%d)\n", sdfID)
	children := getSDFOrPanic(sdfID).Children()
	if len(children) == 0 {
		res := pointerLength{Pointer: 0, Length: 0}
		return &res
	}
	childrenIDs := make([]uint32, len(children))
	for i, child := range children {
		childrenIDs[i] = child.ID()
	}
	res := pointerLength{Pointer: uintptr(unsafe.Pointer(&(childrenIDs[0]))), Length: uint32(uintptr(len(childrenIDs)) * unsafe.Sizeof(childrenIDs[0]))}
	//fmt.Printf("<- Children(%d) <- (%v, %v)\n", sdfID, children.Pointer, children.Length)
	return &res
}

//export name
func name(sdfID uint32) *pointerLength {
	//fmt.Printf("-> Children(%d)\n", sdfID)
	res := stringToPointerLength(getSDFOrPanic(sdfID).Name())
	//fmt.Printf("<- Children(%d) <- (%v, %v)\n", sdfID, children.Pointer, children.Length)
	return &res
}

func stringToPointerLength(str string) pointerLength {
	name := []byte(str)
	res := pointerLength{Pointer: uintptr(unsafe.Pointer(&(name[0]))), Length: uint32(uintptr(len(name)) * unsafe.Sizeof(name[0]))}
	return res
}

//export parameters
func parameters(sdfID uint32) *pointerLength {
	//fmt.Printf("-> Parameters(%d)\n", sdfID)
	params := getSDFOrPanic(sdfID).Parameters()
	if len(params) == 0 {
		res := pointerLength{Pointer: 0, Length: 0}
		return &res
	}
	paramsC := make([]sdfParamC, len(params))
	for i, param := range params {
		paramsC[i] = sdfParamC{
			ID:          param.ID,
			Name:        stringToPointerLength(param.Name),
			KindParams:  kindC(param.Kind),
			Value:       kindValue(param.Value),
			Description: stringToPointerLength(param.Description),
		}
	}
	res := pointerLength{Pointer: uintptr(unsafe.Pointer(&(paramsC[0]))), Length: uint32(uintptr(len(paramsC)) * unsafe.Sizeof(paramsC[0]))}
	//fmt.Printf("<- Parameters(%d) <- (%v, %v)\n", sdfID, res.Pointer, res.Length)
	return &res
}

//export set_parameter
func setParameter(sdfID, paramID, paramKindID, paramArg1, paramArg2 uint32) *setParameterRes {
	//fmt.Printf("-> SetParameter(%d, %d, %d, %d, %d)\n", sdfID, paramID, paramKindID, paramArg1, paramArg2)
	var paramVal SDFParamValue
	switch paramKindID {
	case 0: // bool
		paramVal = paramArg1 != 0
	case 1: // int
		paramVal = *(*int32)(unsafe.Pointer(&paramArg1))
	case 2: // float
		paramVal = math.Float32frombits(paramArg1)
	case 3: // string
		//goland:noinspection GoVetUnsafePointer
		memPtr := unsafe.Pointer(uintptr(paramArg1))
		memLen := paramArg2
		var bytes = unsafe.Slice((*byte)(memPtr), memLen)
		paramVal = string(bytes)
	default:
		panic("Invalid paramKindID")
	}
	err := getSDFOrPanic(sdfID).SetParameter(paramID, paramVal)
	res := setParameterRes{Error: 0, ErrorMsg: pointerLength{Pointer: 0, Length: 0}}
	//err = errors.New("testing error on set_parameter")
	if err != nil {
		name := []byte(err.Error())
		res.Error = 1
		res.ErrorMsg.Pointer = uintptr(unsafe.Pointer(&(name[0])))
		res.ErrorMsg.Length = uint32(uintptr(len(name)) * unsafe.Sizeof(name[0]))
	}
	//fmt.Printf("<- SetParameter(%d) <- (%v, %v)\n", sdfID, res.Pointer, res.Length)
	return &res
}

//export changed
func changed(sdfID uint32) *ChangedAABB {
	//fmt.Printf("-> Changed(%d)\n", sdfID)
	changed := getSDFOrPanic(sdfID).Changed()
	res := changedAABBC{
		Changed: 0,
		AABB:    changed.AABB,
	}
	if changed.Changed {
		res.Changed = 1
	}
	//fmt.Printf("<- Changed(%d) <- (%v)\n", sdfID, Changed)
	return &changed
}
