module github.com/Yeicor/sdf-viewer-go/sdf-viewer-go-auto

go 1.18

// Use the local version of sdf-viewer-go!
replace github.com/Yeicor/sdf-viewer-go/sdf-viewer-go v1.1.0 => ../sdf-viewer-go

require (
	github.com/Yeicor/sdf-viewer-go/sdf-viewer-go v1.1.0
	github.com/ojrac/opensimplex-go v1.0.2
)
