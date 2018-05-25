// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ansiterm

import gc "gopkg.in/check.v1"

type attributeSuite struct{}

var _ = gc.Suite(&attributeSuite{})

func (*attributeSuite) TestSGR(c *gc.C) {
	c.Check(unknownAttribute.sgr(), gc.Equals, "")
	c.Check(reset.sgr(), gc.Equals, "\x1b[0m")
	var yellow attribute = 33
	c.Check(yellow.sgr(), gc.Equals, "\x1b[33m")
}

func (*attributeSuite) TestAttributes(c *gc.C) {
	var a attributes
	c.Check(a.sgr(), gc.Equals, "")
	a = append(a, Yellow.foreground())
	c.Check(a.sgr(), gc.Equals, "\x1b[33m")
	a = append(a, Blue.background())
	c.Check(a.sgr(), gc.Equals, "\x1b[33;44m")

	// Add bold to the end to show sorting of the attributes.
	a = append(a, Bold.enable())
	c.Check(a.sgr(), gc.Equals, "\x1b[1;33;44m")
}
