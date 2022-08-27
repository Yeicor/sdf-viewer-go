package sdf_viewer_go_auto

import (
	"github.com/Yeicor/sdf-viewer-go/sdf-viewer-go"
	"github.com/Yeicor/sdf-viewer-go/sdf-viewer-go-auto/reflectwalktinygo"
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
	// PARAMETERS
	sdfCoreType  reflect.Type // Type of the core SDF implementation
	castCoreType func(interface{}) (SDFCore, bool)
	// OUTPUT
	// children is the list of children of the SDF that will be returned.
	children []sdf_viewer_go.SDF
	// TEMPORARY
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

func (c *childrenCollectorWalker) MapElem(_, k, v reflect.Value) error {
	_ = c.checkValue(k)
	_ = c.checkValue(v)
	return nil
}

func (c *childrenCollectorWalker) Slice(value reflect.Value) error {
	_ = c.checkValue(value)
	return nil
}

func (c *childrenCollectorWalker) SliceElem(_ int, value reflect.Value) error {
	_ = c.checkValue(value)
	return nil
}

func (c *childrenCollectorWalker) Array(value reflect.Value) error {
	_ = c.checkValue(value)
	return nil
}

func (c *childrenCollectorWalker) ArrayElem(_ int, value reflect.Value) error {
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

func (c *childrenCollectorWalker) PointerEnter(_ bool, _ reflect.Value) error {
	//_ = c.checkValue(value)
	return nil
}

func (c *childrenCollectorWalker) PointerExit(_ bool) error {
	return nil
}

func (c *childrenCollectorWalker) checkValue(value reflect.Value) error {
	// Stop recursion if a parent was already found as a child of the root node.
	if c.skipEntryUntilLevel > 0 && c.curDepthLevel > c.skipEntryUntilLevel {
		return nil // Ignore descendants of already found child node
	}

	// Stop recursion on invalid values.
	if !value.IsValid() {

		return nil
	}

	// Look for the core SDF implementations and register them automatically as children.
	coreImpl, coreImplOk := interfaceAndImplementsHint(value, c.sdfCoreType)
	if s, ok := c.castCoreType(coreImpl); coreImplOk != nil && *coreImplOk || ok {
		if s2, ok := s.(sdf_viewer_go.SDF); ok {
			// Already and advanced SDF, keep it
			//log.Printf("Found ADVANCED SDF child2: %#+v\n", s2)
			c.foundChild(s2)
		} else {
			// Automatic (default) conversion of core type to advanced type
			//log.Printf("Found core SDF child: %#+v\n", s)
			c.foundChild(NewSDF(s, c.sdfCoreType, c.castCoreType))
		}
		return reflectwalktinygo.SkipEntry // No more recursion TODO: implement this for all type callbacks
	}

	// Any other type is explored recursively
	return nil
}

func (c *childrenCollectorWalker) foundChild(s sdf_viewer_go.SDF) {
	c.children = append(c.children, s)
	c.skipEntryUntilLevel = c.curDepthLevel // Ignore all children of this node
	//fmt.Printf("Found child: %#+v\n", s)
}
