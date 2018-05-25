// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ansiterm

import (
	"bytes"

	gc "gopkg.in/check.v1"
)

type writerSuite struct{}

var _ = gc.Suite(&writerSuite{})

func (*writerSuite) TestNoColor(c *gc.C) {
	buff := &bytes.Buffer{}
	writer := NewWriter(buff)
	c.Check(writer.noColor, gc.Equals, true)

	writer.SetForeground(Yellow)
	writer.SetBackground(Blue)
	writer.SetStyle(Bold)
	writer.ClearStyle(Bold)
	writer.Reset()

	c.Check(buff.String(), gc.Equals, "")
}

func (*writerSuite) TestSetColorCapable(c *gc.C) {
	buff := &bytes.Buffer{}
	writer := NewWriter(buff)
	c.Check(writer.noColor, gc.Equals, true)

	writer.SetColorCapable(true)
	c.Check(writer.noColor, gc.Equals, false)

	writer.SetColorCapable(false)
	c.Check(writer.noColor, gc.Equals, true)
}

func (*writerSuite) newWriter() (*bytes.Buffer, *Writer) {
	buff := &bytes.Buffer{}
	writer := NewWriter(buff)
	writer.noColor = false
	return buff, writer
}

func (s *writerSuite) TestSetForeground(c *gc.C) {
	buff, writer := s.newWriter()
	writer.SetForeground(Yellow)
	c.Check(buff.String(), gc.Equals, "\x1b[33m")
}

func (s *writerSuite) TestSetBackground(c *gc.C) {
	buff, writer := s.newWriter()
	writer.SetBackground(Blue)
	c.Check(buff.String(), gc.Equals, "\x1b[44m")
}

func (s *writerSuite) TestSetStyle(c *gc.C) {
	buff, writer := s.newWriter()
	writer.SetStyle(Bold)
	c.Check(buff.String(), gc.Equals, "\x1b[1m")
}

func (s *writerSuite) TestClearStyle(c *gc.C) {
	buff, writer := s.newWriter()
	writer.ClearStyle(Bold)
	c.Check(buff.String(), gc.Equals, "\x1b[21m")
}

func (s *writerSuite) TestReset(c *gc.C) {
	buff, writer := s.newWriter()
	writer.Reset()
	c.Check(buff.String(), gc.Equals, "\x1b[0m")
}
