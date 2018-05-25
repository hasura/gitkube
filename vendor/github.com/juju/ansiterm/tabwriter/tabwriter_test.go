package tabwriter

import (
	"bytes"
	"testing"

	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	gc.TestingT(t)
}

type tabwriterSuite struct{}

var _ = gc.Suite(&tabwriterSuite{})

func (s *tabwriterSuite) TestRightAlignOverflow(c *gc.C) {
	var buf bytes.Buffer
	tw := NewWriter(&buf, 0, 1, 2, ' ', 0)
	tw.SetColumnAlignRight(2)
	tw.Write([]byte("not\tenough\ttabs"))
	tw.Flush()
	c.Assert(buf.String(), gc.Equals, "not  enough  tabs")
}
