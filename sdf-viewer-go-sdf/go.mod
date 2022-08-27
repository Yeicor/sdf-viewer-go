module github.com/Yeicor/sdf-viewer-go/sdf-viewer-go-sdf

go 1.18

// Use the local version of sdf-viewer-go!
replace github.com/Yeicor/sdf-viewer-go/sdf-viewer-go v1.1.0 => ../sdf-viewer-go

replace github.com/Yeicor/sdf-viewer-go/sdf-viewer-go-auto v1.1.0 => ../sdf-viewer-go-auto

require (
	github.com/Yeicor/sdf-viewer-go/sdf-viewer-go v1.1.0
	github.com/Yeicor/sdf-viewer-go/sdf-viewer-go-auto v1.1.0
	github.com/soypat/sdf v0.0.0-20220713042432-54592899bb0e
	gonum.org/v1/gonum v0.11.1-0.20220625074215-67f3e1dbfccc
)

require github.com/ojrac/opensimplex-go v1.0.2 // indirect
