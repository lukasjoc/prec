package readline

import (
	"bytes"
	"fmt"
	"os"
)

// TODO: do we even need buffered scanning?
// TODO: do we even need a bytes.Buffer?
// TODO: better general readline behavour
// TODO: keep bounds of line with cursor movements
// TODO: pos movement functions
// TODO: backspace support
// TODO: in memory history
type Readline struct {
	sc        *Scanner[Input]
	buf       bytes.Buffer
	pos       uint16
	linewidth uint16
}

func New(file *os.File) Readline {
	sc := NewScanner[Input](file)
	sc.Split(ScanInput)
	return Readline{
		sc:        sc,
		buf:       bytes.Buffer{},
		pos:       0,
		linewidth: 90,
	}
}

func (r *Readline) withinEditorBounds() bool {
	return r.pos > 0 && r.pos <= r.linewidth
}

func (r *Readline) MoveLeft() {
	if !r.withinEditorBounds() {
		return
	}
	r.pos -= 1
	fmt.Print("\033[1D")
}

func (r *Readline) MoveRight() {
	if !r.withinEditorBounds() {
		return
	}
	r.pos += 1
	fmt.Print("\033[1C")
}

func (r *Readline) MoveToNextLine() {
	r.pos = 0
	fmt.Print("\033[1E")
}

func (r *Readline) ClearLine() {
	r.pos = 0
	r.buf = bytes.Buffer{}
	fmt.Print("\033[2K\r")
}

func (r *Readline) Text() string {
	return r.buf.String()
}

func (r *Readline) Put(ch rune) {
	r.pos += 1
	r.buf.WriteRune(ch)
	// TODO: do not erase whole line
	fmt.Print("\033[2K\r")
	fmt.Print(r.Text())
}

func (r *Readline) Poll() Input {
	r.sc.Scan()
	return r.sc.Next()
}
