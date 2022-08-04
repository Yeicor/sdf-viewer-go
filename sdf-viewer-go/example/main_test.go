package main

import (
	sdfviewergo "github.com/Yeicor/sdf-viewer-go/sdf-viewer-go"
	"testing"
)

func TestScene(t *testing.T) {
	sdfviewergo.TestImpl(t, sceneSDF())
}
