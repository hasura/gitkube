// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ansiterm

import gc "gopkg.in/check.v1"

type styleSuite struct{}

var _ = gc.Suite(&styleSuite{})

func (*styleSuite) TestString(c *gc.C) {
	c.Check(Bold.String(), gc.Equals, "bold")
	c.Check(Strikethrough.String(), gc.Equals, "strikethrough")
	var blank Style
	c.Check(blank.String(), gc.Equals, "")
	var huge Style = 1234
	c.Check(huge.String(), gc.Equals, "")
	var negative Style = -1
	c.Check(negative.String(), gc.Equals, "")
}
