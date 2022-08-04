package sdf_viewer_go_sdfx

import (
	"errors"
	sdfviewergo "github.com/Yeicor/sdf-viewer-go/sdf-viewer-go"
	"github.com/Yeicor/sdf-viewer-go/sdf-viewer-go-sdfx/reflectwalktinygo"
	"github.com/deadsy/sdfx/sdf"
	"github.com/ojrac/opensimplex-go"
	"hash/crc64"
	"math"
	"math/rand"
	"reflect"
	"unsafe"
)

var _ sdfviewergo.SDF = &SDF{}

// SDF wraps a sdf.SDF3 object to introduce the advanced features of the SDF Viewer App by implementing sdf_viewer_go.SDF.
type SDF struct {
	// SDF is the underlying SDF object from the SDFX library.
	sdf.SDF3
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
func NewSDF(sdfCore sdf.SDF3) *SDF {
	return &SDF{
		SDF3: sdfCore,
	}
}

func (s *SDF) BoundingBox() (aabb [2][3]float32) {
	if s.BoundingBoxCache != nil {
		return *s.BoundingBoxCache
	}
	box := s.SDF3.BoundingBox()
	enlargeBy := box.Max.Sub(box.Min).MulScalar(0.01)
	enlargeBy = enlargeBy.Clamp(sdf.V3{X: 2.5, Y: 2.5, Z: 2.5}, enlargeBy)
	box = box.Enlarge(enlargeBy) // Add a bit of padding to avoid rendering artifacts
	s.BoundingBoxCache = &[2][3]float32{
		{float32(box.Min.X), float32(box.Min.Y), float32(box.Min.Z)},
		{float32(box.Max.X), float32(box.Max.Y), float32(box.Max.Z)},
	}
	return *s.BoundingBoxCache
}

func (s *SDF) Sample(point [3]float32, distanceOnly bool) (sample sdfviewergo.SDFSample) {
	dist := s.SDF3.Evaluate(sdf.V3{X: float64(point[0]), Y: float64(point[1]), Z: float64(point[2])})
	sample.Distance = float32(dist)
	if !distanceOnly {
		if s.MaterialFunc != nil {
			s.MaterialFunc(point, &sample) // Modifies the sample pointer
		} else { // Default material function
			children := s.Children()
			if len(children) == 0 { // Leaf nodes: pseudo-random color based on object name
				sample = s.getBaseSample()      // Cached copy
				sample.Distance = float32(dist) // Recover distance
				// Add some noise for imperfections of the material properties along the surface
				// WARNING: ~3 times slower...
				//noise := s.getNoise() // Cached
				//aabb := s.BoundingBox()
				//mappedPoint := [3]float32{
				//	15. * (point[0] - aabb[0][0]) / (aabb[1][0] - aabb[0][0]),
				//	15. * (point[1] - aabb[0][1]) / (aabb[1][1] - aabb[0][1]),
				//	15. * (point[2] - aabb[0][2]) / (aabb[1][2] - aabb[0][2]),
				//}
				//sample.Color[0] -= 0.01 * noise.Eval3(mappedPoint[0], mappedPoint[1], mappedPoint[2])
				//sample.Color[1] -= 0.01 * noise.Eval3(mappedPoint[0], mappedPoint[1], mappedPoint[2])
				//sample.Color[2] -= 0.01 * noise.Eval3(mappedPoint[0], mappedPoint[1], mappedPoint[2])
				//sample.Roughness += -0.025 + 0.05*noise.Eval3(mappedPoint[0]+50, mappedPoint[1]+50, mappedPoint[2]+25)
				//sample.Roughness = myClamp(sample.Roughness, 0., 1.)
				//sample.Metallic += -0.025 + 0.05*noise.Eval3(mappedPoint[0]-50, mappedPoint[1]+50, mappedPoint[2]+25)
				//sample.Metallic = myClamp(sample.Metallic, 0., 1.)
				//sample.Occlusion += -0.025 + 0.05*noise.Eval3(mappedPoint[0]+50, mappedPoint[1]-50, mappedPoint[2]-25)
				//sample.Occlusion = myClamp(sample.Occlusion, 0., 1.)
				//log.Printf("Leaf node(%#v): %#v", point, sample)
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

func myClamp(v float32, min float32, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func (s *SDF) getBaseSample() sdfviewergo.SDFSample {
	if s.BaseSample == nil {
		name := s.Name()
		seed := crc64.Checksum([]byte(name), crc64.MakeTable(crc64.ISO))
		rng := rand.New(rand.NewSource(int64(seed)))
		s.BaseSample = &sdfviewergo.SDFSample{}
		s.BaseSample.Color[0] = rng.Float32()/2.0 + 0.5
		s.BaseSample.Color[1] = rng.Float32()/2.0 + 0.5
		s.BaseSample.Color[2] = rng.Float32()/2.0 + 0.5
		s.BaseSample.Roughness = rng.Float32()
		s.BaseSample.Metallic = rng.Float32()
		s.BaseSample.Occlusion = rng.Float32()
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
	walker := &childrenCollectorWalker{children: make([]sdfviewergo.SDF, 0, 5)}
	err := reflectwalktinygo.Walk(s.SDF3, walker)
	if err != nil {
		panic(err) // Shouldn't happen?
	}

	s.ChildrenCache = walker.children
	return s.ChildrenCache
}

func (s *SDF) Name() string {
	if s.NameCache == "" {
		s.NameCache = nameOfType(reflect.Indirect(reflect.ValueOf(s.SDF3)), uintptr(unsafe.Pointer(&s.SDF3)))
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
	return res
}
