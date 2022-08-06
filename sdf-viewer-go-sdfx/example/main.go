package main

import (
	"errors"
	"fmt"
	sdfviewergo "github.com/Yeicor/sdf-viewer-go/sdf-viewer-go"
	sdfviewergosdfx "github.com/Yeicor/sdf-viewer-go/sdf-viewer-go-sdfx"
	. "github.com/deadsy/sdfx/sdf"
	"math"
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

	// Simple scene:
	box, _ := Box3D(V3{X: 1., Y: 1., Z: 1.}, 0.25)
	cyl, _ := Cylinder3D(1.5, 0.25, 0.25)
	sdfxSDF := Union3D(box, cyl)
	cyl2, _ := Cylinder3D(1.5, 0.25, 0.25)
	cyl2rot := Transform3D(cyl2, RotateY(DtoR(90)))
	sdfxSDF = Difference3D(sdfxSDF, cyl2rot)

	// Complex scene:
	// Use some advanced features of the SDF Viewer that enhance the core SDF interface (that would also work by itself).
	mainBody := outerShell()
	bodyAdvancedSDF := sdfviewergosdfx.NewSDF(mainBody)
	customMaterialFunc := func(point [3]float32, sample *sdfviewergo.SDFSample) {
		// Make any custom pattern on sample.Color
		sample.Color = [3]float32{ // FIXME: Broken colors?
			float32(0.1 + 0.1*math.Sin(float64(point[0]/10))),
			float32(0.1 + 0.1*math.Cos(float64(point[1]/10))),
			float32(0.1 + 0.1*math.Sin(float64(point[2]/10))),
		}
	}
	makeMaterialParam := func(materialValue string) sdfviewergo.SDFParam {
		return sdfviewergo.SDFParam{
			ID:          0,
			Name:        "Material",
			Kind:        sdfviewergo.SDFParamKindString{Values: []string{"Default", "Custom"}},
			Value:       materialValue,
			Description: "The material to use for this SDF object.",
		}
	}
	bodyAdvancedSDF.ParametersList = []sdfviewergo.SDFParam{makeMaterialParam("Default")}
	bodyAdvancedSDF.MaterialFunc = nil
	bodyAdvancedSDF.SetParameters = func(paramId uint32, value sdfviewergo.SDFParamValue) error {
		if paramId == 0 {
			materialValue := value.(string)
			bodyAdvancedSDF.ParametersList = []sdfviewergo.SDFParam{makeMaterialParam(materialValue)}
			if materialValue == "Default" {
				bodyAdvancedSDF.MaterialFunc = nil
			} else {
				bodyAdvancedSDF.MaterialFunc = customMaterialFunc
			}
			bodyAdvancedSDF.ChangedAABB.Changed = true
			bodyAdvancedSDF.ChangedAABB.AABB = bodyAdvancedSDF.AABB()
			return nil
		}
		return errors.New("unsupported parameter")
	}
	sdfxSDF = Difference3D(bodyAdvancedSDF, subtractive())

	return sdfviewergosdfx.NewSDF(sdfxSDF)
}

// The rest of this file is a copied example SDF from https://github.com/deadsy/sdfx

//-----------------------------------------------------------------------------

// phone body
var phoneW = 78.0  // width
var phoneH = 146.5 // height
var phoneT = 18.0  // thickness
var phoneR = 12.0  // corner radius

// camera hole
var cameraW = 23.5 // width
var cameraH = 33.0 // height
var cameraR = 3.0  // corner radius
var cameraXofs = 0.0
var cameraYofs = 48.0

// speaker hole
var speakerW = 12.5 // width
var speakerH = 15.0 // height
var speakerR = 3.0  // corner radius
var speakerXofs = 23.0
var speakerYofs = -46.0

// wall thickness
var wallT = 3.0

//-----------------------------------------------------------------------------

func phoneBody() SDF3 {
	s2d := Box2D(V2{X: phoneW, Y: phoneH}, phoneR)
	s3d := Extrude3D(s2d, phoneT)
	m := Translate3d(V3{Z: wallT / 2.0})
	return Transform3D(s3d, m)
}

func cameraHole() SDF3 {
	s2d := Box2D(V2{X: cameraW, Y: cameraH}, cameraR)
	s3d := Extrude3D(s2d, wallT+phoneT)
	m := Translate3d(V3{X: cameraXofs, Y: cameraYofs})
	return Transform3D(s3d, m)
}

func speakerHole() SDF3 {
	s2d := Box2D(V2{X: speakerW, Y: speakerH}, speakerR)
	s3d := Extrude3D(s2d, wallT+phoneT)
	m := Translate3d(V3{X: speakerXofs, Y: speakerYofs})
	return Transform3D(s3d, m)
}

//-----------------------------------------------------------------------------
// holes for buttons, jacks, etc.

var holeR = 2.0 // corner radius

func holeLeft(length, yofs, zofs float64) SDF3 {
	w := phoneT * 2.0
	xofs := -(phoneW + wallT) / 2.0
	yofs = (phoneH-length)/2.0 - yofs
	zofs = phoneT + ((phoneT + wallT) / 2.0) - zofs
	s2d := Box2D(V2{X: w, Y: length}, holeR)
	s3d := Extrude3D(s2d, wallT)
	m := Translate3d(V3{X: xofs, Y: yofs, Z: zofs}).Mul(RotateY(DtoR(90)))
	return Transform3D(s3d, m)
}

func holeRight(length, yofs, zofs float64) SDF3 {
	w := phoneT * 2.0
	xofs := (phoneW + wallT) / 2.0
	yofs = (phoneH-length)/2.0 - yofs
	zofs = phoneT + ((phoneT + wallT) / 2.0) - zofs
	s2d := Box2D(V2{X: w, Y: length}, holeR)
	s3d := Extrude3D(s2d, wallT)
	m := Translate3d(V3{X: xofs, Y: yofs, Z: zofs}).Mul(RotateY(DtoR(90)))
	return Transform3D(s3d, m)
}

func holeTop(length, xofs, zofs float64) SDF3 {
	w := phoneT * 2.0
	xofs = -(phoneW-length)/2.0 + xofs
	yofs := (phoneH + wallT) / 2.0
	zofs = phoneT + ((phoneT + wallT) / 2.0) - zofs
	s2d := Box2D(V2{X: length, Y: w}, holeR)
	s3d := Extrude3D(s2d, wallT)
	m := Translate3d(V3{X: xofs, Y: yofs, Z: zofs}).Mul(RotateX(DtoR(90)))
	return Transform3D(s3d, m)
}

func holeBottom(length, xofs, zofs float64) SDF3 {
	w := phoneT * 2.0
	xofs = -(phoneW-length)/2.0 + xofs
	yofs := -(phoneH + wallT) / 2.0
	zofs = phoneT + ((phoneT + wallT) / 2.0) - zofs
	s2d := Box2D(V2{X: length, Y: w}, holeR)
	s3d := Extrude3D(s2d, wallT)
	m := Translate3d(V3{X: xofs, Y: yofs, Z: zofs}).Mul(RotateX(DtoR(90)))
	return Transform3D(s3d, m)
}

//-----------------------------------------------------------------------------

func outerShell() SDF3 {
	w := phoneW + (2.0 * wallT)
	h := phoneH + (2.0 * wallT)
	r := phoneR + wallT
	t := phoneT + wallT
	s2d := Box2D(V2{X: w, Y: h}, r)
	return Extrude3D(s2d, t)
}

//-----------------------------------------------------------------------------

func subtractive() SDF3 {
	return Union3D(
		phoneBody(),
		cameraHole(),
		speakerHole(),
		holeLeft(31.0, 19.5, 8.0),
		holeRight(20.0, 34.0, 8.0),
		holeTop(13.0, 16.0, 8.0),
		holeTop(13.0, 49.5, 9.0),
		holeBottom(35.0, 20.5, 9.0),
	)
}

//-----------------------------------------------------------------------------
