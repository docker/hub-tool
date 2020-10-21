package tabwriter

import (
	"fmt"
	"io"
)

type tw struct {
	idx     int
	lines   [][]string
	widths  [][]int
	w       io.Writer
	padding string
}

// TabWriter interface :)
type TabWriter interface {
	Column(string, int)
	Line()
	Flush() error
}

// New creates a new tab writer that output to the io.Writer
func New(w io.Writer, padding string) TabWriter {
	return &tw{
		idx:     0,
		lines:   [][]string{},
		widths:  [][]int{},
		w:       w,
		padding: padding,
	}
}

func (t *tw) Column(s string, l int) {
	if len(t.lines) <= t.idx {
		t.lines = append(t.lines, []string{})
		t.widths = append(t.widths, []int{})
	}

	t.lines[t.idx] = append(t.lines[t.idx], s)
	t.widths[t.idx] = append(t.widths[t.idx], l)
}

func (t *tw) Line() {
	t.idx++
}

func (t *tw) Flush() error {
	maxs := []int{}
	for i := range t.lines[0] {
		maxs = append(maxs, t.max(i))
	}

	for k, l := range t.lines {
		for i, c := range l {
			pad := ""
			if i != 0 {
				pad = t.pad(maxs[i-1], t.widths[k][i-1])
			}
			if _, err := fmt.Fprint(t.w, pad+c); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(t.w, ""); err != nil {
			return err
		}
	}

	return nil
}

func (t *tw) pad(max int, l int) string {
	ret := t.padding
	for i := 0; i < max-l; i++ {
		ret += " "
	}
	return ret
}

func (t *tw) max(col int) int {
	w := 0
	for _, width := range t.widths {
		for i, ww := range width {
			if i == col {
				if ww > w {
					w = ww
				}
			}
		}
	}
	return w
}
