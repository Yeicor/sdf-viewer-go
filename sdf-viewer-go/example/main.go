package main

import (
	"fmt"
	sdfviewergo "github.com/Yeicor/sdf-viewer-go/sdf-viewer-go"
	"math"
)

func init() {
	// This is the only function you need to call to initialize the SDF Viewer.
	sdfviewergo.SetRootSDF(sceneSDF())
}

func main() {
	fmt.Println("This is not an executable. Compile this with `tinygo build -o example.wasm -target wasi -opt 2 -x -no-debug .` and " +
		"use the SDF Viewer app (github.com/Yeicor/sdf-viewer) to visualize the SDF.")
}

// sceneSDF returns the root SDF of the scene.
func sceneSDF() sdfviewergo.SDF {
	return &SampleSDF{id: 0, name: "test-root-cube", cubeHalfSide: 0.99}
}

// ######################## START OF EXAMPLE MANUAL SDF IMPLEMENTATION ########################

type SampleSDF struct {
	id           uint32
	name         string
	cubeHalfSide float32 // Cube side length
}

func (s *SampleSDF) BoundingBox() [2][3]float32 {
	return [2][3]float32{{-1, -1, -1}, {1, 1, 1}}
}

func (s *SampleSDF) Sample(point [3]float32, distanceOnly bool) (sample sdfviewergo.SDFSample) {
	// Cube SDF
	sample.Distance = maxF32(maxF32(absF32(point[0]), absF32(point[1])), absF32(point[2])) - s.cubeHalfSide
	if !distanceOnly {
		sample.Color = [3]float32{sinF32(point[0] * 2.0), (point[1] + 1.0) / 2.0, (point[2] + 1.0) / 2.0}
		sample.Metallic = modF32(point[0], 1.0)
		sample.Roughness = modF32(point[1], 1.0)
		sample.Occlusion = modF32(point[2], 1.0)
	}
	return
}

func (s *SampleSDF) Children() []sdfviewergo.SDF {
	if s.id == 0 { // Fake, just for testing...
		return []sdfviewergo.SDF{&SampleSDF{id: 1, name: "test-fake-child", cubeHalfSide: 0.51}}
	}
	return []sdfviewergo.SDF{}
}

func (s *SampleSDF) ID() uint32 {
	return s.id
}

func (s *SampleSDF) Name() string {
	return s.name
}

func sinF32(f float32) float32 {
	return float32(math.Sin(float64(f)))
}

func modF32(f float32, f2 float32) float32 {
	res := float32(math.Mod(float64(f), float64(f2)))
	if res < 0 {
		return res + f2
	}
	return res
}

func maxF32(v1 float32, v2 float32) float32 {
	if v1 > v2 {
		return v1
	}
	return v2
}

func absF32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// ######################## END OF EXAMPLE MANUAL SDF IMPLEMENTATION ########################
