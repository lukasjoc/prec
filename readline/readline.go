package readline

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

// TODO:
// key loop
// pos movement functions
// event system
// in memory history

type KeyType int

const (
	KeyRune KeyType = iota
	KeyInvalid
	KeyArrowLeft
	KeyArrowRight
)

// TODO: i dont like it.. but we can refactor it later (to complicated like this)
var termcodes = map[KeyType]termcode{
	KeyArrowRight: {"\033[C", nil},
	KeyArrowLeft:  {"\033[D", nil},
}

// TODO: make more generic (add proper parser for terminal codes)
type termcode struct {
	code  string
	bytes []byte
}

func (ts *termcode) Len() int { return len(ts.code) }
func (ts *termcode) Bytes() []byte {
	if ts.bytes != nil {
		return ts.bytes
	}
	b := []byte(ts.code)
	ts.bytes = b
	return b
}

// TODO: do we even need buffered scanning?
// TODO: do we even need a bytes.Buffer?
type Readline struct {
	sc        *bufio.Scanner
	buf       bytes.Buffer
	pos       uint16
	linewidth uint16
}

func New(file *os.File) Readline {
	sc := bufio.NewScanner(file)
	sc.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// TODO: handle errors and atEOF correctly
		// check for terminal esc patterns
		for typ, code := range termcodes {
			l := code.Len()
			if bytes.Equal(data[0:l], code.Bytes()) {
				switch typ {
				case KeyArrowRight:
					return l, []byte{byte(KeyArrowRight)}, nil
				case KeyArrowLeft:
					return l, []byte{byte(KeyArrowLeft)}, nil
				}
			}
		}
		// else we want to read a single byte
		return 1, []byte{byte(KeyRune), data[0]}, nil
	})
	return Readline{
		sc:        sc,
		buf:       bytes.Buffer{},
		pos:       0,
		linewidth: 90,
	}
}

func (r *Readline) withinEditorBounds() bool {
	return r.pos > 1 && r.pos < r.linewidth
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

func (r *Readline) Put() {
	// TODO: do not erase whole line
	fmt.Print("\033[2K\r")
	fmt.Print(r.buf.String())
}

func (r *Readline) MoveToNextLine() {
	r.buf = bytes.Buffer{}
	r.pos = 0
	fmt.Print("\033[1E")
}

type Input interface {
	Key() KeyType
}

type RuneInput struct {
	typ   KeyType
	value rune
}

func (i RuneInput) Key() KeyType { return i.typ }
func (i *RuneInput) Value() rune { return i.value }

type KeyInput struct {
	typ KeyType
}

func (i KeyInput) Key() KeyType { return i.typ }

func (r *Readline) Poll() Input {
	r.sc.Scan()
	b := r.sc.Bytes()

	// TODO: store the events somewhere in the readline and return ptrs instead
	// TODO: actually have a key event queue and maybe even batching
	switch KeyType(b[0]) {
	case KeyArrowLeft:
		return KeyInput{KeyArrowLeft}
	case KeyArrowRight:
		return KeyInput{KeyArrowRight}
	case KeyRune:
		// we know that there is only a single character byte in here
		// otherwise something else went wrong. (^ Handle errors better)
		// NOTE: that we only know this in here once we know the key code is KeyRune
		rest := b[1:]
		return RuneInput{KeyRune, rune(rest[0])}
	}
	return KeyInput{KeyInvalid}
}
