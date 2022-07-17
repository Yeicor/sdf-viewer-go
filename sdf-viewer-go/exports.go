package sdf_viewer_go

import "C"

// availableSDFs is the exported SDF hierarchy implementations.
var availableSDFs map[uint32]SDF

// SetRootSDF registers the root SDF, overriding any previous value.
func SetRootSDF(sdf SDF) {
	availableSDFs = map[uint32]SDF{}
	availableSDFs[0] = sdf
	// TODO: Children!
}

//export bounding_box
func BoundingBox(sdfID uint32) *[2][3]float32 {
	//fmt.Printf("-> BoundingBox(%d)\n", sdfID)
	minMax := availableSDFs[sdfID].BoundingBox()
	//fmt.Printf("<- BoundingBox(%d) <- (%v, %v)\n", sdfID, minMax[0], minMax[1])
	return &minMax
}

//export sample
func Sample(sdfID uint32, point [3]float32, distanceOnly bool) *SDFSample {
	//fmt.Printf("BoundingBox(%d, %v, %v)\n", sdfID, point, distanceOnly)
	sample := availableSDFs[sdfID].Sample(point, distanceOnly)
	//fmt.Printf("BoundingBox(%d, %v, %v) <- (%v)\n", sdfID, point, distanceOnly, sample)
	return &sample
}
