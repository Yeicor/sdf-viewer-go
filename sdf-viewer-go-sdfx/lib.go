package sdf_viewer_go_auto

import (
	sdfviewergoauto "github.com/Yeicor/sdf-viewer-go/sdf-viewer-go-auto"
	"github.com/deadsy/sdfx/sdf"
	"github.com/deadsy/sdfx/vec/v3"
	"reflect"
)

var _ sdfviewergoauto.SDFCore = &SDFCore{}

func NewSDF(s sdf.SDF3) *SDFWrapper {
	return &SDFWrapper{sdfviewergoauto.NewSDF(&SDFCore{s}, reflect.TypeOf((*sdf.SDF3)(nil)).Elem(), func(s interface{}) (sdfviewergoauto.SDFCore, bool) {
		s2, ok := s.(sdf.SDF3)
		if ok {
			if _, ok2 := s2.(*SDFWrapper); ok2 {
				return nil, false // Ignore our wrapper, which also implementes SDF3, to avoid infinite recursion
			}
			return &SDFCore{s2}, ok
		} else {
			return nil, false
		}
	})}
}

type SDFCore struct {
	SDF3 sdf.SDF3
}

func (s *SDFCore) SDFCoreEval(p [3]float32) float32 {
	return float32(s.SDF3.Evaluate(v3.Vec{X: float64(p[0]), Y: float64(p[1]), Z: float64(p[2])}))
}

func (s *SDFCore) SDFCoreAABB() [2][3]float32 {
	box := s.SDF3.BoundingBox()
	enlargeBy := box.Max.Sub(box.Min).MulScalar(0.01)
	enlargeBy = enlargeBy.Clamp(v3.Vec{X: 2.5, Y: 2.5, Z: 2.5}, enlargeBy)
	box = box.Enlarge(enlargeBy) // Add a bit of padding to avoid rendering artifacts
	return [2][3]float32{
		{float32(box.Min.X), float32(box.Min.Y), float32(box.Min.Z)},
		{float32(box.Max.X), float32(box.Max.Y), float32(box.Max.Z)},
	}
}

func (s *SDFCore) SDFCoreChildrenRoot() interface{} {
	return s.SDF3 // Avoid infinite recursion
}

var _ sdf.SDF3 = &SDFWrapper{}

type SDFWrapper struct {
	*sdfviewergoauto.SDF
}

func (s *SDFWrapper) Evaluate(p v3.Vec) float64 {
	return s.SDF.SDF.(*SDFCore).SDF3.Evaluate(p)
}

func (s *SDFWrapper) BoundingBox() sdf.Box3 {
	return s.SDF.SDF.(*SDFCore).SDF3.BoundingBox()
}
