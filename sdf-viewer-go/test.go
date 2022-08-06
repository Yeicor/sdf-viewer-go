package sdf_viewer_go

import (
	"strconv"
	"testing"
)

func TestImpl(_ *testing.T, s SDF) {
	// Configure the root SDF
	SetRootSDF(s)

	// Test that operations on root and ALL descendant nodes don't panic
	for childID := range availableSDFs {
		testSubSDF(childID, 1)
	}

	// TODO: More and better tests
}

func BenchmarkImpl(t *testing.B, s SDF) {
	// Configure the root SDF
	SetRootSDF(s)

	// Test that operations on root and ALL descendant nodes don't panic
	for childID := range availableSDFs {
		t.Run("Child#"+strconv.Itoa(int(childID)), func(b *testing.B) {
			testSubSDF(childID, b.N)
		})
	}

	// TODO: More and better tests
}

func testSubSDF(childID uint32, times int) {
	for i := 0; i < times; i++ {
		bounding_box(childID)
		sample(childID, [3]float32{0, 0, 0}, false)
		children(childID)
		//for _, child := range getSDFOrPanic(childID).Children() {
		//	fmt.Printf("Children[%d]: %#+v\n", childID, child)
		//}
		parameters(childID)
		changed(childID)
	}
}
