package ast

import (
	"fmt"
	"io"
)

type Continue struct {
	nodeBase
	Label string
}

func NewContinue(pos PositionHolder, label string) *Continue {
	n := &Continue{Label: label}
	n.init(pos)
	return n
}

func (n *Continue) dump(w io.Writer, nest int) {
	header(n, w, nest, false)
	_, _ = fmt.Fprintf(w, "%v\n", n.Label)
}
