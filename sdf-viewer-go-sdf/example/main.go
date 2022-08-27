package main

import (
	"fmt"
	sdfviewergo "github.com/Yeicor/sdf-viewer-go/sdf-viewer-go"
	sdfviewergosdf "github.com/Yeicor/sdf-viewer-go/sdf-viewer-go-sdf"
	"github.com/soypat/sdf"
	"github.com/soypat/sdf/form3"
	"github.com/soypat/sdf/form3/obj3/thread"
	"gonum.org/v1/gonum/spatial/r3"
)

//export init
func init() {
	// This is the only function you need to call to initialize the SDF Viewer.
	sdfviewergo.SetRootSDF(sceneSDF())
}

func main() {
	fmt.Println("This is not an executable. Compile this with `" +
		"tinygo build -o example.wasm -target wasi -opt 2 -x -no-debug ." +
		"` and use the SDF Viewer app (github.com/Yeicor/sdf-viewer) to visualize the SDF.")
}

// sceneSDF returns the root SDF of the scene.
func sceneSDF() sdfviewergo.SDF {
	return sdfviewergosdf.NewSDF(getMainModel())
}

// The rest of this file is a copied example SDF from https://github.com/soypat/sdf

const (
	// thread length
	tlen             = 18 / 25.4
	internalDiameter = 1.5 / 2.
	flangeH          = 7 / 25.4
	flangeD          = 60. / 25.4
	// internal diameter scaling.
	plaScale = 1.03
)

func getMainModel() sdf.SDF3 {
	var (
		npt    thread.NPT
		flange sdf.SDF3
	)
	err := npt.SetFromNominal(1.0 / 2.0)
	if err != nil {
		panic(err)
	}
	pipe, err := thread.Nut(thread.NutParms{
		Thread: npt,
		Style:  thread.NutCircular,
	})
	if err != nil {
		panic(err)
	}
	// PLA scaling to thread
	pipe = sdf.Transform3D(pipe, sdf.Scale3D(r3.Vec{X: plaScale, Y: plaScale, Z: 1}))
	flange, err = form3.Cylinder(flangeH, flangeD/2, flangeH/8)
	if err != nil {
		panic(err)
	}
	flange = sdf.Transform3D(flange, sdf.Translate3D(r3.Vec{Z: -tlen / 2}))
	union := sdf.Union3D(pipe, flange)
	// set flange fillet
	union.SetMin(sdf.MinPoly(2, 0.2))
	// Make through-hole in flange bottom
	hole, err := form3.Cylinder(4*flangeH, internalDiameter/2, 0)
	if err != nil {
		panic(err)
	}
	pipe = sdf.Difference3D(union, hole)
	//pipe = sdf.ScaleUniform3D(pipe, 25.4) //convert to millimeters

	return pipe
}
