package sdf_viewer_go

// SDF provides access to the Signed Distance Function data. Keep in sync with the SDF Viewer app.
type SDF interface {
	// BoundingBox is the bounding box of the SDF. Returns the minimum and maximum coordinates of the SDF.
	// All operations MUST be inside this bounding box.
	BoundingBox() (aabb [2][3]float32)

	// Sample samples the surface at the given point. It should include the effect of all of its children
	// and none of its parents. See `SDFSample` for more information.
	// `distanceOnly` is a hint to the implementation that the caller only needs the distance.
	Sample(point [3]float32, distanceOnly bool) (sample SDFSample)
}

type SDFSample struct {
	Distance                       float32
	Color                          [3]float32
	Metallic, Roughness, Occlusion float32
}
