package converter

import (
	"strings"
    "github.com/iancoleman/orderedmap"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// ProtoPackage describes a package of Protobuf, which is an container of message types.
type ProtoPackage struct {
	name     string
	parent   *ProtoPackage
	children *orderedmap.OrderedMap
	types    *orderedmap.OrderedMap
}

func (c *Converter) lookupType(pkg *ProtoPackage, name string) (*descriptor.DescriptorProto, bool) {
	if strings.HasPrefix(name, ".") {
		return c.relativelyLookupType(globalPkg, name[1:len(name)])
	}

	for ; pkg != nil; pkg = pkg.parent {
		if desc, ok := c.relativelyLookupType(pkg, name); ok {
			return desc, ok
		}
	}
	return nil, false
}

func (c *Converter) relativelyLookupType(pkg *ProtoPackage, name string) (*descriptor.DescriptorProto, bool) {
	components := strings.SplitN(name, ".", 2)
	switch len(components) {
	case 0:
		c.logger.Debug("empty message name")
		return nil, false
	case 1:
		found_tmp, ok := pkg.types.Get(components[0])
		found := found_tmp.(*descriptor.DescriptorProto)
		return found, ok
	case 2:
		c.logger.Tracef("Looking for %s in %s at %s (%v)", components[1], components[0], pkg.name, pkg)
		child_tmp, ok := pkg.children.Get(components[0])
		child := child_tmp.(*ProtoPackage)
		if ok {
			found, ok := c.relativelyLookupType(child, components[1])
			return found, ok
		}
		msg_tmp, ok := pkg.types.Get(components[0])
		msg := msg_tmp.(*descriptor.DescriptorProto)
		if ok {
			found, ok := c.relativelyLookupNestedType(msg, components[1])
			return found, ok
		}
		c.logger.WithField("component", components[0]).WithField("package_name", pkg.name).Info("No such package nor message in package")
		return nil, false
	default:
		c.logger.Error("Failed to lookup type")
		return nil, false
	}
}

func (c *Converter) relativelyLookupPackage(pkg *ProtoPackage, name string) (*ProtoPackage, bool) {
	components := strings.Split(name, ".")
	for _, c := range components {
		pkg_tmp, ok := pkg.children.Get(c)
		pkg = pkg_tmp.(*ProtoPackage)
		if !ok {
			return nil, false
		}
	}
	return pkg, true
}
