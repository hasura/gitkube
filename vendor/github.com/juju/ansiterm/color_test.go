// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ansiterm

import gc "gopkg.in/check.v1"

type colorSuite struct{}

var _ = gc.Suite(&colorSuite{})

func (*colorSuite) TestString(c *gc.C) {
	c.Check(Default.String(), gc.Equals, "default")
	c.Check(Yellow.String(), gc.Equals, "yellow")
	c.Check(BrightMagenta.String(), gc.Equals, "brightmagenta")
	var blank Color
	c.Check(blank.String(), gc.Equals, "")
	var huge Color = 1234
	c.Check(huge.String(), gc.Equals, "")
	var negative Color = -1
	c.Check(negative.String(), gc.Equals, "")
}
