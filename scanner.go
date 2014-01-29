package toml

import (
	"unicode/utf8"
)

const (
	EOF       = 0xF8
	RuneError = 0xFFFD
)

type Scanner interface {
	Get() string
	Next() rune
	Eof() bool
}

type scanner struct {
	buf    []byte
	offset int // offset for DecodeRune
	size   int // bytes of last rune
	cache  bool
	eof    bool
	err    string
	r      rune
}

func NewScanner(source []byte) Scanner {
	p := new(scanner)
	p.buf = source
	r := p.Next()
	if r == BOM {
		p.Get() // Truncate BOM
	}
	p.offset = 0
	p.size = 0
	p.cache = false
	p.eof = false
	p.r = 0

	return p
}
func (p *scanner) Eof() bool {
	return p.eof
}

// Get returns string(buffer[:Scan.offset - sizeOfLastChar]).
func (p *scanner) Get() (str string) {
	if p.buf == nil {
		panic("buffer null")
	}
	if p.eof {
		str = string(p.buf)
		p.buf = nil
		return
	}
	pos := p.offset - p.size
	if pos < 0 || p.offset == 0 {
		panic("expected Next() first")
	}
	if pos == 0 {
		pos = p.offset
		p.cache = false
		p.offset, p.size = 0, 0
	} else {
		p.cache = true
		p.offset = p.size
	}
	str = string(p.buf[:pos])
	p.buf = p.buf[pos:]
	return
}

// Next returns read char frome the buffer.
// b is byte(char) when size of char equal 1, otherwise it is const MultiBytes.
// r is rune value of char,If the encoding is invalid, it is RuneError.
// if end of buffer or encoding is invalid, char is error string, b equal const EOF, r equal const RuneError.
func (p *scanner) Next() rune {
	if p.eof {
		panic("EOF")
	}
	if p.cache {
		p.cache = false
		return p.r
	}
	if p.offset >= len(p.buf) {
		p.size = 0
		p.eof = true
		return EOF
	}
	p.r, p.size = utf8.DecodeRune(p.buf[p.offset:])
	if p.r == RuneError {
		p.size = 0
		return p.r
	}
	p.offset += p.size
	return p.r
}
