package sdf_viewer_go_auto

import (
	"errors"
	sdfviewergo "github.com/Yeicor/sdf-viewer-go/sdf-viewer-go"
	"github.com/Yeicor/sdf-viewer-go/sdf-viewer-go-auto/reflectwalktinygo"
	"github.com/ojrac/opensimplex-go"
	"hash/crc64"
	"math"
	"math/rand"
	"reflect"
	"unsafe"
)

var _ sdfviewergo.SDF = &SDF{}

type SDFCore interface {
	SDFCoreEval([3]float32) float32
	SDFCoreAABB() [2][3]float32
	SDFCoreChildrenRoot() interface{}
}

// SDF wraps an implementation-specific SDFCore object to introduce the advanced features of the SDF Viewer App by implementing sdf_viewer_go.SDF.
type SDF struct {
	// PARAMETERS TO SET BY IMPLEMENTATION WRAPPER
	// SDF is the SDFCore object from the underlying library.
	SDF SDFCore
	// Concrete type that can be converted to the SDFCore interface
	sdfCoreType reflect.Type
	// Casting/conversion code to the SDFCore interface
	castCoreType func(interface{}) (SDFCore, bool)
	// PARAMETERS TO CONFIGURE BY USER
	// BoundingBoxCache is the cached bounding box of the SDF.
	BoundingBoxCache *[2][3]float32
	// MaterialFunc allows you to modify the material of the SDF for each sampled point.
	// The `sample` parameter only has the distance to the surface set, and you may add any texture, lighting or material
	// property to it.
	// If left as nil, it will use the default material function.
	MaterialFunc func(point [3]float32, sample *sdfviewergo.SDFSample)
	// NameCache is the name of this SDF object. If left as empty, it will use the default name.
	NameCache string
	// ChildrenCache is the list of children of this SDF object.
	// If left as empty (or manually set to nil), it will automatically the default children by exploring the SDF
	// hierarchy using reflect. This is automatically set to nil after Changed is true.
	ChildrenCache []sdfviewergo.SDF
	// ParametersList is the set of parameters to dynamically configure this SDF.
	// Should be set manually as the default is no parameters.
	ParametersList []sdfviewergo.SDFParam
	// SetParameters is the function that modify the parameters at `ParametersList` to dynamically configure this SDF.
	SetParameters func(paramId uint32, value sdfviewergo.SDFParamValue) error
	// ChangedAABB is the modified bounding box of this SDF.
	// It is usually modified by a `SetParameters` call, but it may be modified at any point in time.
	// This is returned once and reset to no changes reported.
	// It should be implemented manually to be as precise as possible for each change.
	ChangedAABB sdfviewergo.ChangedAABB

	// BaseSample to which randomness will be added if using the default material
	BaseSample *sdfviewergo.SDFSample
	// Noise is the noise generator used to generate noise for this SDF.
	Noise opensimplex.Noise32
}

// NewSDF see SDF
func NewSDF(sdfCore SDFCore, sdfCoreType reflect.Type, castCoreType func(interface{}) (SDFCore, bool)) *SDF {
	return &SDF{
		SDF:          sdfCore,
		sdfCoreType:  sdfCoreType,
		castCoreType: castCoreType,
	}
}

func (s *SDF) AABB() (aabb [2][3]float32) {
	if s.BoundingBoxCache != nil {
		return *s.BoundingBoxCache
	}
	box := s.SDF.SDFCoreAABB()
	s.BoundingBoxCache = &box
	return *s.BoundingBoxCache
}

func (s *SDF) Sample(point [3]float32, distanceOnly bool) (sample sdfviewergo.SDFSample) {
	dist := s.SDF.SDFCoreEval(point)
	sample.Distance = dist
	if !distanceOnly {
		if s.MaterialFunc != nil {
			s.MaterialFunc(point, &sample) // Modifies the sample pointer
		} else { // Default material function
			children := s.Children()
			if len(children) == 0 { // Leaf nodes: pseudo-random color based on object name
				sample = s.getBaseSample() // Cached copy
				sample.Distance = dist     // Recover distance
			} else { // Non-leaf nodes (union, intersection, difference, etc...): copy closest child material
				closest := math.MaxFloat64
				var closestChild *SDF
				for _, child := range children {
					childSample := child.Sample(point, true)
					if math.Abs(float64(childSample.Distance)) <= closest { // <= seems to work better on ties, but it's a hack
						closest = float64(childSample.Distance)
						closestChild = child.(*SDF)
					}
				}
				savedParentDistance := sample.Distance
				sample = closestChild.Sample(point, false)
				sample.Distance = savedParentDistance
			}
		}
	}
	return
}

func (s *SDF) getBaseSample() sdfviewergo.SDFSample {
	if s.BaseSample == nil {
		name := s.Name()
		seed := crc64.Checksum([]byte(name), crc64.MakeTable(crc64.ISO))
		rng := rand.New(rand.NewSource(int64(seed)))
		s.BaseSample = &sdfviewergo.SDFSample{}
		s.BaseSample.Color[0] = rng.Float32()*0.5 + 0.5
		s.BaseSample.Color[1] = rng.Float32()*0.5 + 0.5
		s.BaseSample.Color[2] = rng.Float32()*0.5 + 0.5
		s.BaseSample.Roughness = rng.Float32()
		s.BaseSample.Metallic = rng.Float32() * 0.5 // Too dark when set to max
		s.BaseSample.Occlusion = rng.Float32()
		//log.Println("Base sample for", name, ":", s.BaseSample)
	}
	return *s.BaseSample
}

func (s *SDF) getNoise() opensimplex.Noise32 {
	if s.Noise == nil {
		name := s.Name()
		seed := crc64.Checksum([]byte(name), crc64.MakeTable(crc64.ISO))
		s.Noise = opensimplex.NewNormalized32(int64(seed))
	}
	return s.Noise
}

func (s *SDF) Children() []sdfviewergo.SDF {
	if s.ChildrenCache != nil { // Cached children to avoid slow automatic reflect operation
		return s.ChildrenCache
	}

	// Children are auto-generated by exploring the underlying SDF struct using the `reflect` package.
	// Any interface matching the basic SDF interface will be added to the list of children and stop recursion.
	// If this advanced SDF struct is found, the same behavior will be applied, but with access to more advanced features.

	// HACK: This will "fail" for SDF3s that store unused instances of other SDF3s, by showing them when they are not used.
	// WARNING: This is a slow operation (reflect is used), so it is cached. However, you may need to invalidate the cache manually.
	// FIXME: This will crash for SDF3s that store pointers to themselves (or cyclic dependencies in general).
	// You may implement this yourself to bypass the above hacks.

	// Start walking the underlying SDF struct, and collecting children.
	walker := &childrenCollectorWalker{
		sdfCoreType:         s.sdfCoreType,
		castCoreType:        s.castCoreType,
		children:            make([]sdfviewergo.SDF, 0, 5),
		curDepthLevel:       0,
		skipEntryUntilLevel: 0,
	}
	err := reflectwalktinygo.Walk(s.SDF.SDFCoreChildrenRoot(), walker)
	if err != nil {
		panic(err) // Shouldn't happen?
	}

	s.ChildrenCache = walker.children
	return s.ChildrenCache
}

func (s *SDF) Name() string {
	if s.NameCache == "" { // Auto compute name from type info of the actual implementation node
		root := s.SDF.SDFCoreChildrenRoot()
		s.NameCache = nameOfType(reflect.Indirect(reflect.ValueOf(root)), uintptr(unsafe.Pointer(&root)))
	}
	return s.NameCache
}

func (s *SDF) Parameters() []sdfviewergo.SDFParam {
	return s.ParametersList // empty list by default
}

func (s *SDF) SetParameter(paramId uint32, value sdfviewergo.SDFParamValue) error {
	if s.SetParameters != nil {
		return s.SetParameters(paramId, value)
	}
	return errors.New("SetParameters is not configured")
}

func (s *SDF) Changed() sdfviewergo.ChangedAABB {
	res := s.ChangedAABB
	if res.Changed {
		s.ChildrenCache = nil // Re-compute children automatically, just in case.
	}
	s.ChangedAABB.Changed = false // Reset always (after being returned)
	// Merge with changes from children!
	for _, child := range s.Children() {
		changed := child.Changed()
		if changed.Changed {
			res.Changed = true
			res.AABB = aabbMerge(res.AABB, changed.AABB)
		}
	}
	return res
}

func aabbMerge(aabb1, aabb2 [2][3]float32) [2][3]float32 {
	return [2][3]float32{
		{
			float32(math.Min(float64(aabb1[0][0]), float64(aabb2[0][0]))),
			float32(math.Min(float64(aabb1[0][1]), float64(aabb2[0][1]))),
			float32(math.Min(float64(aabb1[0][2]), float64(aabb2[0][2]))),
		},
		{
			float32(math.Max(float64(aabb1[1][0]), float64(aabb2[1][0]))),
			float32(math.Max(float64(aabb1[1][1]), float64(aabb2[1][1]))),
			float32(math.Max(float64(aabb1[1][2]), float64(aabb2[1][2]))),
		},
	}
}
