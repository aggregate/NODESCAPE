package config

import (
	"io"
	"bytes"
)

/*
Terminals: (approximately)
	LB	[
	RB	]
	NL	\n, \r\n, \r
	ID	([a-z]|[A-Z]|_)+([a-z]|[A-Z]|_|[0-9])*
	ARG	^[,\n\r]+ 
	CM	,		

Grammar: (approximately)
	C  -> P C
		| e
	P  -> I MK
	I  -> '[' id ']' N 
	MK -> K N MK
		| K N 
	K ->  id A N
	A ->  a ',' A
		| a N 
	N ->  nl+

*/

// states and terminal codes
const (
	start = iota
	lb = iota
	rb = iota
	nl = iota
	id = iota
	arg = iota
	cm = iota
	eof = iota
)

type token struct {
	value string
	t_type int
	line int
	column int
}

type scanner_t struct {
	line, column, nst int
	next []byte
	scan_err error
	in io.Reader
}

func init_scanner(in io.Reader) (scanner_t, error) {
	scanner := scanner_t{1, 1, start, make([]byte, 2), nil, in}
	_, err := scanner.in.Read(scanner.next)
	scanner.next_state()
	return scanner, err
}

func (s* scanner_t) scan() token {

	var sofar bytes.Buffer

	start_col := s.column
	start_line := s.line
	for s.scan_err == nil {
		switch s.nst {
			case start:
				s.get_next()
				s.next_state()
			case lb, rb:
				sofar.Write(s.get_next())
				st := s.nst
				s.next_state()
				return token{sofar.String(), st, start_line, start_col}
			case nl:
				s.line++
				s.column = 1
				if s.next[0] == '\r' && s.next[1] == '\n' {
					_, s.scan_err = s.in.Read(s.next)
					s.next_state()
					return token{"\r\n", nl, s.line, s.column}
				} else {
					sofar.Write(s.next[0:1])
					s.next[0] = s.next[1]
					 _, s.scan_err = s.in.Read(s.next[1:])
					s.next_state()
					return token{sofar.String(), nl, start_line, start_col}
				}
			case id:
				sofar.Write(s.get_next())
				for valid_id(s.next[0]) {
					sofar.Write(s.get_next())
				}
				s.next_state()
				return token{sofar.String(), id, start_line, start_col}
			case arg:
				sofar.Write(s.get_next())
				for !delimiter(s.next[0]) {
					sofar.Write(s.get_next())
				}
				s.next_state()
				return token{sofar.String(), arg, start_line, start_col}
			case cm:
				s.get_next()
				s.next_state()
				return token{",", cm, start_line, start_col}
		} // switch
	} // for

	return token{"", eof, s.line, s.column}
}

func (s *scanner_t) next_state() {
	switch  {
		case s.next[0] == '[': s.nst = lb
		case s.next[0] == ']': s.nst = rb
		case s.next[0] == ',': s.nst = cm
		case valid_id_st(s.next[0]): s.nst = id
		case s.next[0] == ' ' || s.next[0] == '\t': s.nst = start
		case s.next[0] == '\n' || s.next[0] == '\r': s.nst = nl
		default: s.nst = arg
	}
}

func (s *scanner_t) get_next() []byte {
	old := make([]byte, 1)
	copy(old, s.next[0:1])
	s.next[0] = s.next[1]
	_, s.scan_err = s.in.Read(s.next[1:])
	s.column++
	return old
}

// Use isalpha or something.
func valid_id_st(next byte) bool {
	return 'a' <= next && next <= 'z' ||
			'A' <= next && next <= 'Z' ||
			next == '_'
}

func valid_id(next byte) bool {
	return valid_id_st(next) || '0' <= next  && next <= '9'
}

func delimiter(next byte) bool {
	return next == ',' || next == '\n' || next == '\r' || next == '[' ||
			next == ']' || next == ' ' || next == '\t'
}
