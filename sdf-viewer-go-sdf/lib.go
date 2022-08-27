package sdf_viewer_go_auto

import (
	sdfviewergoauto "github.com/Yeicor/sdf-viewer-go/sdf-viewer-go-auto"
	"github.com/soypat/sdf"
	"gonum.org/v1/gonum/spatial/r3"
	"reflect"
)

var _ sdfviewergoauto.SDFCore = &SDFCore{}

func NewSDF(s sdf.SDF3) *SDFWrapper {
	return &SDFWrapper{sdfviewergoauto.NewSDF(&SDFCore{s}, reflect.TypeOf((*sdf.SDF3)(nil)).Elem(), func(s interface{}) (sdfviewergoauto.SDFCore, bool) {
		s2, ok := s.(sdf.SDF3)
		if ok {
			if _, ok2 := s2.(*SDFWrapper); ok2 {
				return nil, false // Ignore our wrapper, which also implementes SDF3, to avoid losing custom SDF data
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
	return float32(s.SDF3.Evaluate(r3.Vec{X: float64(p[0]), Y: float64(p[1]), Z: float64(p[2])}))
}

func (s *SDFCore) SDFCoreAABB() [2][3]float32 {
	box := s.SDF3.Bounds()
	enlargeBy := r3.Add(r3.Vec{X: 1, Y: 1, Z: 1}, r3.Scale(0.01, r3.Sub(box.Max, box.Min)))
	box = box.Scale(enlargeBy) // Add a bit of padding to avoid rendering artifacts
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

func (s *SDFWrapper) Evaluate(p r3.Vec) float64 {
	return s.SDF.SDF.(*SDFCore).SDF3.Evaluate(p)
}

func (s *SDFWrapper) Bounds() r3.Box {
	return s.SDF.SDF.(*SDFCore).SDF3.Bounds()
}
