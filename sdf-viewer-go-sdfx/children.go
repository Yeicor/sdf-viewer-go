package sdf_viewer_go_sdfx

import (
	"github.com/Yeicor/sdf-viewer-go/sdf-viewer-go"
	"github.com/Yeicor/sdf-viewer-go/sdf-viewer-go-sdfx/reflectwalktinygo"
	"github.com/deadsy/sdfx/sdf"
	"reflect"
)

var _ reflectwalktinygo.PrimitiveWalker = &childrenCollectorWalker{}
var _ reflectwalktinygo.InterfaceWalker = &childrenCollectorWalker{}
var _ reflectwalktinygo.MapWalker = &childrenCollectorWalker{}
var _ reflectwalktinygo.SliceWalker = &childrenCollectorWalker{}
var _ reflectwalktinygo.ArrayWalker = &childrenCollectorWalker{}
var _ reflectwalktinygo.StructWalker = &childrenCollectorWalker{}
var _ reflectwalktinygo.EnterExitWalker = &childrenCollectorWalker{}
var _ reflectwalktinygo.PointerWalker = &childrenCollectorWalker{}

type childrenCollectorWalker struct {
	// children is the list of children of the SDF that will be returned.
	children                           []sdf_viewer_go.SDF
	curDepthLevel, skipEntryUntilLevel int
}

func (c *childrenCollectorWalker) Primitive(value reflect.Value) error {
	_ = c.checkValue(value)
	return nil
}

func (c *childrenCollectorWalker) Interface(value reflect.Value) error {
	_ = c.checkValue(value)
	return nil
}

func (c *childrenCollectorWalker) Map(m reflect.Value) error {
	_ = c.checkValue(m)
	return nil
}

func (c *childrenCollectorWalker) MapElem(m, k, v reflect.Value) error {
	_ = c.checkValue(k)
	_ = c.checkValue(v)
	return nil
}

func (c *childrenCollectorWalker) Slice(value reflect.Value) error {
	_ = c.checkValue(value)
	return nil
}

func (c *childrenCollectorWalker) SliceElem(i int, value reflect.Value) error {
	_ = c.checkValue(value)
	return nil
}

func (c *childrenCollectorWalker) Array(value reflect.Value) error {
	_ = c.checkValue(value)
	return nil
}

func (c *childrenCollectorWalker) ArrayElem(i int, value reflect.Value) error {
	_ = c.checkValue(value)
	return nil
}

func (c *childrenCollectorWalker) Struct(value reflect.Value) error {
	return c.checkValue(value)
}

func (c *childrenCollectorWalker) StructField(_ reflect.StructField, _ reflect.Value) error {
	return nil // Will be called again as whatever type it is, so we don't care now.
}

func (c *childrenCollectorWalker) Enter(_ reflectwalktinygo.Location) error {
	c.curDepthLevel++
	return nil
}

func (c *childrenCollectorWalker) Exit(_ reflectwalktinygo.Location) error {
	c.curDepthLevel--
	if c.curDepthLevel < c.skipEntryUntilLevel {
		c.skipEntryUntilLevel = 0
	}
	return nil
}

func (c *childrenCollectorWalker) PointerEnter(b bool, value reflect.Value) error {
	//_ = c.checkValue(value)
	return nil
}

func (c *childrenCollectorWalker) PointerExit(b bool) error {
	return nil
}

var sdfAdvancedType = reflect.TypeOf((*sdf_viewer_go.SDF)(nil)).Elem()
var sdfCoreType = reflect.TypeOf((*sdf.SDF3)(nil)).Elem()

func (c *childrenCollectorWalker) checkValue(value reflect.Value) error {
	// Stop recursion if a parent was already found as a child of the root node.
	if c.skipEntryUntilLevel > 0 && c.curDepthLevel > c.skipEntryUntilLevel {
		return nil // Ignore descendants of already found child node
	}

	// Look for the advanced SDF implementations and register them automatically as children.
	advImpl, advImplOk := interfaceAndImplementsHint(value, sdfAdvancedType)
	if s, ok := advImpl.(sdf_viewer_go.SDF); advImplOk != nil && *advImplOk || ok {
		c.foundChild(s)
		return reflectwalktinygo.SkipEntry // No more recursion
	}

	// Look for the core SDF implementations and register them automatically as children.
	coreImpl, coreImplOk := interfaceAndImplementsHint(value, sdfCoreType)
	if s, ok := coreImpl.(sdf.SDF3); coreImplOk != nil && *coreImplOk || ok {
		// Automatic (default) conversion of core type to advanced type
		c.foundChild(NewSDF(s))
		return reflectwalktinygo.SkipEntry // No more recursion
	}

	// Any other type is explored recursively
	return nil
}

func (c *childrenCollectorWalker) foundChild(s sdf_viewer_go.SDF) {
	c.children = append(c.children, s)
	c.skipEntryUntilLevel = c.curDepthLevel // Ignore all children of this node
	//fmt.Printf("Found child: %#+v\n", s)
}
