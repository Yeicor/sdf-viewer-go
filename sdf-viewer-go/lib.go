package sdf_viewer_go

import (
	"math"
	"unsafe"
)

// SetRootSDF registers the root SDF, overriding any previous value.
func SetRootSDF(sdf SDF) {
	// Reset, in case this is called multiple times
	nextSDFID = 0
	availableSDFs = map[uint32]SDF{}
	// Also register all children, recursively.
	registerSDFAndChildren(sdf)
}

// SDF provides access to the Signed Distance Function data. Keep in sync with the SDF Viewer app.
// Comments may be outdated, so check the original SDF Viewer app for more details.
type SDF interface {
	// BoundingBox is the bounding box of the SDF. Returns the minimum and maximum coordinates of the SDF.
	// All operations MUST be inside this bounding box.
	BoundingBox() (aabb [2][3]float32)

	// Sample samples the surface at the given point. It should include the effect of all of its children
	// and none of its parents. See `SDFSample` for more information.
	// `distanceOnly` is a hint to the implementation that the caller only needs the distance.
	Sample(point [3]float32, distanceOnly bool) (sample SDFSample)

	// Children returns the list of sub-SDFs that are directly children of this node.
	// Note that modifications to the parameters of the returned children MUST affect this node.
	Children() (children []SDF)

	// Name returns a nice display Name for the SDF, which does not need to be unique in the hierarchy.
	Name() string

	// Parameters returns the list of parameters (including Values and metadata) that can be modified on this SDF.
	Parameters() []SDFParam

	// SetParameter modifies the given parameter. The Value must be valid for the reported type (same Kind and within allowed Values)
	// Implementations will probably need interior mutability to perform this.
	// Use [`Changed`](#method.Changed) to notify what part of the SDF needs to be updated.
	SetParameter(paramId uint32, value SDFParamValue) error

	// Changed Returns the bounding box that was modified since [`Changed`](#method.Changed) was last called.
	// It should also report if the children of this SDF need to be updated.
	// This may happen due to a parameter change ([`set_parameter`](#method.set_parameter)) or any
	// other event that may have Changed the SDF. It should delimit as much as possible the part of the
	// SDF that should be updated to improve performance.
	//
	// Multiple changes should be merged into a single bounding box or queued and returned in several
	// [`Changed`](#method.Changed) calls for a similar effect.
	// After returning Some(...) the implementation should assume that it was updated and no longer
	// notify of that change (to avoid infinite loops).
	// This function is called very frequently, so it should be very fast to avoid delaying frames.
	Changed() ChangedAABB
}

type SDFSample struct {
	Distance                       float32
	Color                          [3]float32
	Metallic, Roughness, Occlusion float32
}

type SDFParam struct {
	// The ID of the parameter. Must be unique within this SDF (not necessarily within the SDF hierarchy).
	ID uint32
	// The Name of the parameter.
	Name string
	// The type definition for the parameter.
	Kind SDFParamKind
	// The current Value of the parameter. MUST be of the same Kind as the type definition.
	Value SDFParamValue
	// The user-facing Description for the parameter.
	Description string
}

// SDFParamKind is one of the types below
type SDFParamKind interface{}

type SDFParamKindBool struct {
}

type SDFParamKindInt struct {
	// The minimum Value of the parameter.
	Min int32
	// The maximum Value of the parameter.
	Max int32
	// The Step size of the parameter.
	Step int32
}

type SDFParamKindFloat struct {
	// The minimum Value of the parameter.
	Min float32
	// The maximum Value of the parameter.
	Max float32
	// The Step size of the parameter.
	Step float32
}

type SDFParamKindString struct {
	// The list of possible Values of the parameter.
	Values []string
}

// SDFParamValue is one of bool, int, float or string
type SDFParamValue interface{}

type ChangedAABB struct {
	// Has it Changed?
	Changed bool
	// The bounding box that was Changed.
	AABB [2][3]float32
}

// === Private API ===

type pointerLength struct {
	Pointer uintptr // Always 32 bits on wasm32
	Length  uint32
}

type sdfParamC struct {
	ID          uint32
	Name        pointerLength
	KindParams  sdfParamKindC
	Value       sdfParamValueC
	Description pointerLength
}

func kindC(k SDFParamKind) sdfParamKindC {
	switch v := k.(type) {
	case SDFParamKindBool:
		return sdfParamKindC{
			KindID: 0,
			Params: [3]uint32{0, 0, 0},
		}
	case SDFParamKindInt:
		valMin := *(*uint32)(unsafe.Pointer(&v.Min)) // Same as math.Float32bits for int32 -> uint32
		valMax := *(*uint32)(unsafe.Pointer(&v.Max))
		valStep := *(*uint32)(unsafe.Pointer(&v.Step))
		return sdfParamKindC{
			KindID: 1,
			Params: [3]uint32{valMin, valMax, valStep},
		}
	case SDFParamKindFloat:
		return sdfParamKindC{
			KindID: 2,
			Params: [3]uint32{math.Float32bits(v.Min), math.Float32bits(v.Max), math.Float32bits(v.Step)},
		}
	case SDFParamKindString:
		values := pointerLength{Pointer: 0, Length: 0}
		if len(v.Values) > 0 {
			values.Pointer = uintptr(unsafe.Pointer(&v.Values[0]))
			values.Length = uint32(uintptr(len(v.Values)) * unsafe.Sizeof(v.Values[0]))
		}
		return sdfParamKindC{
			KindID: 3,
			Params: [3]uint32{uint32(values.Pointer), values.Length, 0},
		}
	default:
		panic("unknown kind")
	}
}

type sdfParamKindC struct {
	KindID uint32
	Params [3]uint32 // Interpretation depends on KindID (Go does not support enums with data)
}

func kindValue(v SDFParamValue) sdfParamValueC {
	switch v := v.(type) {
	case bool:
		res := uint32(0)
		if v {
			res = 1
		}
		return sdfParamValueC{
			KindID: 0,
			Params: [2]uint32{res, 0},
		}
	case int32:
		return sdfParamValueC{
			KindID: 1,
			Params: [2]uint32{uint32(v), 0},
		}
	case float32:
		return sdfParamValueC{
			KindID: 2,
			Params: [2]uint32{math.Float32bits(v), 0},
		}
	case string:
		res := []byte(v)
		return sdfParamValueC{
			KindID: 3,
			Params: [2]uint32{uint32(uintptr(unsafe.Pointer(&res[0]))), uint32(len(res))},
		}
	default:
		panic("unknown kind")
	}
}

type sdfParamValueC struct {
	KindID uint32
	Params [2]uint32 // Interpretation depends on KindID (Go does not support enums with data)
}

type setParameterRes struct {
	Error    uint32 // 0 or 1 for success or failure
	ErrorMsg pointerLength
}

type changedAABBC struct {
	// Has it Changed?
	Changed uint32 // 0 or 1
	// The bounding box that was Changed.
	AABB [2][3]float32
}
