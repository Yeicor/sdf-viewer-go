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

	// Children returns the list of sub-SDFs that are directly children of this node.
	// Note that modifications to the parameters of the returned children MUST affect this node.
	Children() (children []SDF)

	// ID returns an unique ID within this SDF hierarchy. Root must return 0.
	ID() uint32

	// Name returns anice display name for the SDF, which does not need to be unique in the hierarchy.
	Name() string
}

type SDFSample struct {
	Distance                       float32
	Color                          [3]float32
	Metallic, Roughness, Occlusion float32
}

type PointerLength struct {
	Pointer uintptr // Always 32 bits on wasm32
	Length  uint32
}
