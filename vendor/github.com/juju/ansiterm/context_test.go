// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ansiterm

import (
	"bytes"

	gc "gopkg.in/check.v1"
)

type contextSuite struct{}

var _ = gc.Suite(&contextSuite{})

func (*contextSuite) newWriter() (*bytes.Buffer, *Writer) {
	buff := &bytes.Buffer{}
	writer := NewWriter(buff)
	writer.noColor = false
	return buff, writer
}

func (*contextSuite) TestBlank(c *gc.C) {
	var context Context
	c.Assert(context.sgr(), gc.Equals, "")
}

func (*contextSuite) TestAllUnknown(c *gc.C) {
	context := Context{
		Foreground: 123,
		Background: 432,
		Styles:     []Style{456, 99},
	}
	c.Assert(context.sgr(), gc.Equals, "")
}

func (*contextSuite) TestForeground(c *gc.C) {
	context := Foreground(Yellow)
	c.Assert(context.sgr(), gc.Equals, "\x1b[33m")
}

func (*contextSuite) TestBackground(c *gc.C) {
	context := Background(Blue)
	c.Assert(context.sgr(), gc.Equals, "\x1b[44m")
}

func (*contextSuite) TestStyles(c *gc.C) {
	context := Styles(Bold, Italic)
	c.Assert(context.sgr(), gc.Equals, "\x1b[1;3m")
}

func (*contextSuite) TestValid(c *gc.C) {
	context := Context{
		Foreground: Yellow,
		Background: Blue,
		Styles:     []Style{Bold, Italic},
	}
	c.Assert(context.sgr(), gc.Equals, "\x1b[1;3;33;44m")
}

func (*contextSuite) TestSetForeground(c *gc.C) {
	var context Context
	context.SetForeground(Yellow)
	c.Assert(context.sgr(), gc.Equals, "\x1b[33m")
}

func (*contextSuite) TestSetBackground(c *gc.C) {
	var context Context
	context.SetBackground(Blue)
	c.Assert(context.sgr(), gc.Equals, "\x1b[44m")
}

func (*contextSuite) TestSetStyles(c *gc.C) {
	var context Context
	context.SetStyle(Bold, Italic)
	c.Assert(context.sgr(), gc.Equals, "\x1b[1;3m")
}

func (s *contextSuite) TestFprintfNoColor(c *gc.C) {
	buff, writer := s.newWriter()
	writer.noColor = true

	context := Context{
		Foreground: Yellow,
		Background: Blue,
		Styles:     []Style{Bold, Italic},
	}

	context.Fprintf(writer, "hello %s, %d", "world", 42)
	c.Assert(buff.String(), gc.Equals, "hello world, 42")
}

func (s *contextSuite) TestFprintfColor(c *gc.C) {
	buff, writer := s.newWriter()

	context := Context{
		Foreground: Yellow,
		Background: Blue,
		Styles:     []Style{Bold, Italic},
	}

	context.Fprintf(writer, "hello %s, %d", "world", 42)
	c.Assert(buff.String(), gc.Equals, "\x1b[1;3;33;44mhello world, 42\x1b[0m")
}
