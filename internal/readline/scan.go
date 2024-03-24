package readline

import (
	"bufio"
	"os"
)

// NOTE:
// custom generic wrapper around `bufio.Scanner`. Its tailored to the use case
// here. It probably wasnt meant to be used like this and will not work for
// other contexts. Additionally the scanner relies heavily on the token being
// able to be nil. This cant be easily done with generics and structs without
// ptrs i think.. So types need to implement the custom invalidator instead, to
// continue scanning if the token is incomplete.  Nice try but I might just have
// two functions instead next time that do the conversions from and to bytes.
// (Like i had before)
type Scannable interface {
	Invalid() bool
	Bytes() []byte
}

type SplitFunc[T Scannable] func(data []byte, atEOF bool) (advance int, token T, err error)
type Scanner[T Scannable] struct {
	sc    *bufio.Scanner
	token T
}

func NewScanner[T Scannable](file *os.File) *Scanner[T] {
	return &Scanner[T]{sc: bufio.NewScanner(file)}
}
func (s *Scanner[T]) Next() T    { return s.token }
func (s *Scanner[T]) Scan() bool { return s.sc.Scan() }
func (s *Scanner[T]) Split(split SplitFunc[T]) {
	s.sc.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		a, t, err := split(data, atEOF)
		if err != nil {
			return 0, nil, err
		}
		if t.Invalid() {
			return 0, nil, err
		}
		s.token = t
		return a, t.Bytes(), err
	})
}
