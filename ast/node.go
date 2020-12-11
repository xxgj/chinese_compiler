package ast

import (
	"fmt"
	"io"
)

type Node interface {
	Line() int
	Column() int
	dump(io.Writer, int)
}

type nodeBase struct {
	line   int
	column int
}

type PositionHolder interface {
	Line() int
	Column() int
}

type TypeSpec interface{}

func (n *nodeBase) init(pos PositionHolder) {
	n.line = pos.Line()
	n.column = pos.Column()
}

func (n *nodeBase) Line() int   { return n.line }
func (n *nodeBase) Column() int { return n.column }

func Dump(tree Node, output io.Writer) {
	dumpNode(tree, output, 0)
}

func indent(w io.Writer, nest int) {
	if nest < 1 {
		return
	}
	for i := 0; i < nest; i++ {
		_, _ = fmt.Fprint(w, "    ")
	}
}

func header(n Node, w io.Writer, nest int, requireNewline bool) {
	indent(w, nest)
	_, _ = fmt.Fprintf(w, "「第%d行」:「第%d列」%T: ", n.Line(), n.Column(), n)
	if requireNewline {
		_, _ = fmt.Fprint(w, "\n")
	}
}

func tag(name string, w io.Writer, nest int) {
	indent(w, nest)
	_, _ = fmt.Fprintf(w, "#%s:\n", name)
}

func dumpNode(n Node, w io.Writer, nest int) {
	if n == nil {
		indent(w, nest)
		_, _ = fmt.Fprintln(w, "(nil)")
		return
	}
	n.dump(w, nest)
}

func dumpNodes(nodes []Node, w io.Writer, nest int) {
	if nodes == nil {
		indent(w, nest)
		_, _ = fmt.Fprintln(w, "(nil)")
		return
	}
	for _, n := range nodes {
		dumpNode(n, w, nest)
	}
}
