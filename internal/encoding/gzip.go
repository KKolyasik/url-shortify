package encoding

import (
	"compress/gzip"
	"io"
)

type Gzip struct {
	level int
}

func NewGzip(level int) *Gzip {
	return &Gzip{level: level}
}

func (g *Gzip) Name() string { return "gzip" }

func (g *Gzip) WrapWriter(w io.Writer) (io.WriteCloser, error) {
	level := g.level
	if level == 0 {
		level = gzip.DefaultCompression
	}
	return gzip.NewWriterLevel(w, level)
}

func (g *Gzip) WrapReader(r io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}