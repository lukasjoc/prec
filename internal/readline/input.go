package readline

import "bytes"

type KeyType int

const (
	KeyRune KeyType = iota
	KeyInvalid
	KeyArrowLeft
	KeyArrowRight
	KeyArrowUp
	KeyArrowDown
)

var termcodes = map[KeyType]termcode{
	KeyArrowRight: {"\033[C", nil},
	KeyArrowLeft:  {"\033[D", nil},
	KeyArrowUp:    {"\033[A", nil},
	KeyArrowDown:  {"\033[B", nil},
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

type Input interface {
	Scannable
	Key() KeyType
}

var ScanInput SplitFunc[Input] = func(data []byte, atEOF bool) (advance int, token Input, err error) {
	for typ, code := range termcodes {
		l := code.Len()
		if bytes.Equal(data[:l], code.Bytes()) {
			return l, KeyInput{typ}, nil
		}
	}
	return 1, RuneInput{typ: KeyRune, value: rune(data[0])}, nil
}

type RuneInput struct {
	typ   KeyType
	value rune
}

func (i RuneInput) Invalid() bool { return i.typ == KeyInvalid }
func (i RuneInput) Bytes() []byte { return []byte{byte(i.typ), byte(i.value)} }
func (i RuneInput) Key() KeyType  { return i.typ }
func (i *RuneInput) Value() rune  { return i.value }

type KeyInput struct {
	typ KeyType
}

func (i KeyInput) Invalid() bool { return i.typ == KeyInvalid }
func (i KeyInput) Bytes() []byte { return []byte{byte(i.typ)} }
func (i KeyInput) Key() KeyType  { return i.typ }
