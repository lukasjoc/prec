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

type InputEventType int

const (
	InputEventGeneric InputEventType = iota
	InputEventArrrowLeft
	InputEventArrowRight
)

type Readline struct {
	sc        *bufio.Scanner
	buf       bytes.Buffer
	bufpos    uint16
	bufposmax uint16
}

// TODO: make more generic (add proper parser for terminal codes)
type termcode struct {
	code  string
	bytes []byte
}

func (ts *termcode) Bytes() []byte {
	if ts.bytes != nil {
		return ts.bytes
	}
	b := []byte(ts.code)
	ts.bytes = b
	return b
}

func (ts *termcode) Len() int { return len(ts.code) }

// TODO: i dont like it.. but we can refactor it later (to complicated like this)
var termcodes = map[InputEventType]termcode{
	InputEventArrowRight: {"\033[C", nil},
	InputEventArrrowLeft: {"\033[D", nil},
}

func New(file *os.File) Readline {
	sc := bufio.NewScanner(file)
	sc.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// check for terminal esc patterns
		for typ, code := range termcodes {
			l := code.Len()
			if bytes.Equal(data[0:l], code.Bytes()) {
				switch typ {
				case InputEventArrowRight:
					return l, []byte{byte(InputEventArrowRight)}, nil
				case InputEventArrrowLeft:
					return l, []byte{byte(InputEventArrrowLeft)}, nil
				}
			}
		}
		// else we want to read a single byte
		return 1, []byte{byte(InputEventGeneric), data[0]}, nil
	})
	return Readline{
		sc:        sc,
		buf:       bytes.Buffer{},
		bufpos:    0,
		bufposmax: 90,
	}
}

func (r *Readline) withinEditorBounds() bool {
	return r.bufpos > 1 && r.bufpos < r.bufposmax
}

func (r *Readline) MoveLeft() {
	if !r.withinEditorBounds() {
		return
	}
	r.bufpos -= 1
	fmt.Print("\033[1D")
}

func (r *Readline) MoveRight() {
	if !r.withinEditorBounds() {
		return
	}
	r.bufpos += 1
	fmt.Print("\033[1C")
}

func (r *Readline) Put() {
	// TODO: do not erase whole line
	fmt.Print("\033[2K\r")
	fmt.Print(r.buf.String())
}

func (r *Readline) MoveToNextLine() {
	r.buf = bytes.Buffer{}
	r.bufpos = 0
	fmt.Print("\033[1E")
}

type Input struct {
	Key   InputEventType
	Bytes *[]byte
}

func (r *Readline) Poll() Input {
	r.sc.Scan()
	b := r.sc.Bytes()

	// TODO: store the events somewhere in the readline and return ptrs instead
	switch InputEventType(b[0]) {
	case InputEventGeneric:
		a := b[0:]
		return Input{InputEventGeneric, &a}
	case InputEventArrrowLeft:
		return Input{InputEventArrrowLeft, nil}
	case InputEventArrowRight:
		return Input{InputEventArrowRight, nil}
	}
	return Input{InputEventGeneric, nil}

	// switch b {
	// case KeyEsc:
	// 	// TODO: actually have a key event queue and maybe even batching
	// 	b, _ := r.reader.ReadByte()
	// 	switch b {
	// 	case KeyOpenBracket:
	// 		b, _ := r.reader.ReadByte()
	// 		// TODO: ESC[#A	moves cursor up # lines
	// 		// TODO: ESC[#B	moves cursor down # lines
	// 		switch b {
	//
	// 		// ESC[#C	moves cursor right # columns
	// 		case KeyC:
	// 			r.MoveRight()
	// 			break
	//
	// 		// ESC[#D   moves cursor left # columns
	// 		case KeyD:
	// 			r.MoveLeft()
	// 			break
	// 		}
	// 	}
	// // case KeyEnter, KeyEnter1:
	// // 	r.MoveToNextLine()
	// // 	break
	// // case KeyBackspace:
	// // 	// TODO: remove content with backspace
	// // 	// need curr cursor pos + buffer pos etc.. first factor out
	// // 	break
	// default:
	// }
}
