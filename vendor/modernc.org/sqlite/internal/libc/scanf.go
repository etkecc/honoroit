// Copyright 2020 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !(linux && (amd64 || arm64 || loong64))

package libc // import "modernc.org/sqlite/internal/libc"

import (
	"io"
	"strconv"
	"strings"
	"unsafe"
)

func scanf(r io.ByteScanner, format, args uintptr) (nvalues int32) {
	var ok bool
out:
	for {
		c := *(*byte)(unsafe.Pointer(format))

		switch c {
		case '%':
			var n int
			var match bool
			format, n, match = scanfConversion(r, format, &args)
			if !match {
				break out
			}

			nvalues += int32(n)
			ok = true
		case 0:
			break out
		case ' ', '\t', '\n', '\r', '\v', '\f':
			format = skipWhiteSpace(format)
			ok = true
		next:
			for {
				c, err := r.ReadByte()
				if err != nil {
					break out
				}

				switch c {
				case ' ', '\t', '\n', '\r', '\v', '\f':
				default:
					r.UnreadByte()
					break next
				}
			}
		default:
			c2, err := r.ReadByte()
			if err != nil {
				break out
			}

			if c2 != c {
				r.UnreadByte()
				break out
			}

			format++
			ok = true
		}
	}
	if ok {
		return nvalues
	}

	return -1
}

func scanfConversion(r io.ByteScanner, format uintptr, args *uintptr) (_ uintptr, nvalues int, match bool) {
	format++

	mod := 0
	width := -1
	discard := false
flags:
	for {
		switch c := *(*byte)(unsafe.Pointer(format)); c {
		case '*':
			format++
			discard = true
		case '\'':
			format++
			panic(todo(""))
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			width = 0
		num:
			for {
				var digit int
				switch c := *(*byte)(unsafe.Pointer(format)); {
				default:
					break num
				case c >= '0' && c <= '9':
					format++
					digit = int(c) - '0'
				}
				width0 := width
				width = 10*width + digit
				if width < width0 {
					panic(todo(""))
				}
			}
		case 'h', 'j', 'l', 'L', 'q', 't', 'z':
			format, mod = parseLengthModifier(format)
		default:
			break flags
		}
	}

	switch c := *(*byte)(unsafe.Pointer(format)); c {
	case '%':
		format++
		skipReaderWhiteSpace(r)
		c, err := r.ReadByte()
		if err != nil {
			return format, -1, false
		}

		if c == '%' {
			return format, 1, true
		}

		r.UnreadByte()
		return format, 0, false
	case 'd':
		format++
		skipReaderWhiteSpace(r)
		var digit, n uint64
		allowSign := true
		neg := false
	dec:
		for ; width != 0; width-- {
			c, err := r.ReadByte()
			if err != nil {
				if match {
					break dec
				}

				return 0, 0, false
			}

			if allowSign {
				switch c {
				case '-':
					allowSign = false
					neg = true
					continue
				case '+':
					allowSign = false
					continue
				}
			}

			switch {
			case c >= '0' && c <= '9':
				digit = uint64(c) - '0'
			default:
				r.UnreadByte()
				break dec
			}
			match = true
			n0 := n
			n = n*10 + digit
			if n < n0 {
				panic(todo(""))
			}
		}
		if !match {
			break
		}

		if !discard {
			arg := VaUintptr(args)
			v := int64(n)
			if neg {
				v = -v
			}
			switch mod {
			case modNone:
				*(*int32)(unsafe.Pointer(arg)) = int32(v)
			case modH:
				*(*int16)(unsafe.Pointer(arg)) = int16(v)
			case modHH:
				*(*int8)(unsafe.Pointer(arg)) = int8(v)
			case modL:
				*(*long)(unsafe.Pointer(arg)) = long(v)
			case modLL:
				*(*int64)(unsafe.Pointer(arg)) = int64(v)
			default:
				panic(todo("", mod))
			}
		}
		nvalues = 1
	case 'D':
		format++
		panic(todo(""))
	case 'i':
		format++
		panic(todo(""))
	case 'o':
		format++
		panic(todo(""))
	case 'u':
		format++
		panic(todo(""))
	case 'x', 'X':
		format++
		skipReaderWhiteSpace(r)
		var digit, n uint64
		allowPrefix := true
		var b []byte
	hex:
		for ; width != 0; width-- {
			c, err := r.ReadByte()
			if err != nil {
				if match || err == io.EOF {
					break hex
				}

				panic(todo("", err))
			}

			if allowPrefix {
				if len(b) == 1 && b[0] == '0' && (c == 'x' || c == 'X') {
					allowPrefix = false
					match = false
					b = nil
					continue
				}

				b = append(b, c)
			}

			switch {
			case c >= '0' && c <= '9':
				digit = uint64(c) - '0'
			case c >= 'a' && c <= 'f':
				digit = uint64(c) - 'a' + 10
			case c >= 'A' && c <= 'F':
				digit = uint64(c) - 'A' + 10
			default:
				r.UnreadByte()
				break hex
			}
			match = true
			n0 := n
			n = n<<4 + digit
			if n < n0 {
				panic(todo(""))
			}
		}
		if !match {
			break
		}

		if !discard {
			arg := VaUintptr(args)
			switch mod {
			case modNone:
				*(*uint32)(unsafe.Pointer(arg)) = uint32(n)
			case modH:
				*(*uint16)(unsafe.Pointer(arg)) = uint16(n)
			case modHH:
				*(*byte)(unsafe.Pointer(arg)) = byte(n)
			case modL:
				*(*ulong)(unsafe.Pointer(arg)) = ulong(n)
			default:
				panic(todo(""))
			}
		}
		nvalues = 1
	case 'f', 'e', 'g', 'E', 'a':
		format++
		skipReaderWhiteSpace(r)
		seq := fpLiteral(r)
		if len(seq) == 0 {
			return 0, 0, false
		}

		var neg bool
		switch seq[0] {
		case '+':
			seq = seq[1:]
		case '-':
			neg = true
			seq = seq[1:]
		}
		n, err := strconv.ParseFloat(string(seq), 64)
		if err != nil {
			panic(todo("", err))
		}

		if !discard {
			arg := VaUintptr(args)
			if neg {
				n = -n
			}
			switch mod {
			case modNone:
				*(*float32)(unsafe.Pointer(arg)) = float32(n)
			case modL:
				*(*float64)(unsafe.Pointer(arg)) = n
			default:
				panic(todo("", mod, neg, n))
			}
		}
		return format, 1, true
	case 's':
		var c byte
		var err error
		var arg uintptr
		if !discard {
			arg = VaUintptr(args)
		}
	scans:
		for ; width != 0; width-- {
			if c, err = r.ReadByte(); err != nil {
				if err != io.EOF {
					nvalues = -1
				}
				break scans
			}

			switch c {
			case ' ', '\t', '\n', '\r', '\v', '\f':
				break scans
			}

			nvalues = 1
			match = true
			if !discard {
				*(*byte)(unsafe.Pointer(arg)) = c
				arg++
			}
		}
		if match {
			switch {
			case width == 0:
				r.UnreadByte()
				fallthrough
			default:
				if !discard {
					*(*byte)(unsafe.Pointer(arg)) = 0
				}
			}
		}
	case 'c':
		format++
		panic(todo(""))
	case '[':
		format++
		var re0 []byte
	bracket:
		for i := 0; ; i++ {
			c := *(*byte)(unsafe.Pointer(format))
			format++
			if c == ']' && i != 0 {
				break bracket
			}

			re0 = append(re0, c)
		}
		set := map[byte]struct{}{}
		re := string(re0)
		neg := strings.HasPrefix(re, "^")
		if neg {
			re = re[1:]
		}
		for len(re) != 0 {
			switch {
			case len(re) >= 3 && re[1] == '-':
				for c := re[0]; c <= re[2]; c++ {
					set[c] = struct{}{}
				}
				re = re[3:]
			default:
				set[c] = struct{}{}
				re = re[1:]
			}
		}
		var arg uintptr
		if !discard {
			arg = VaUintptr(args)
		}
		for ; width != 0; width-- {
			c, err := r.ReadByte()
			if err != nil {
				if err == io.EOF {
					return format, nvalues, match
				}

				return format, -1, match
			}

			if _, ok := set[c]; ok == !neg {
				match = true
				nvalues = 1
				if !discard {
					*(*byte)(unsafe.Pointer(arg)) = c
					arg++
				}
			}
		}
		if match {
			switch {
			case width == 0:
				r.UnreadByte()
				fallthrough
			default:
				if !discard {
					*(*byte)(unsafe.Pointer(arg)) = 0
				}
			}
		}
	case 'p':
		format++
		skipReaderWhiteSpace(r)
		c, err := r.ReadByte()
		if err != nil {
			panic(todo("", err))
		}

		if c == '0' {
			if c, err = r.ReadByte(); err != nil {
				panic(todo("", err))
			}

			if c != 'x' && c != 'X' {
				r.UnreadByte()
			}
		}

		var digit, n uint64
	ptr:
		for ; width != 0; width-- {
			c, err := r.ReadByte()
			if err != nil {
				if match {
					break ptr
				}

				panic(todo(""))
			}

			switch {
			case c >= '0' && c <= '9':
				digit = uint64(c) - '0'
			case c >= 'a' && c <= 'f':
				digit = uint64(c) - 'a' + 10
			case c >= 'A' && c <= 'F':
				digit = uint64(c) - 'A' + 10
			default:
				r.UnreadByte()
				break ptr
			}
			match = true
			n0 := n
			n = n<<4 + digit
			if n < n0 {
				panic(todo(""))
			}
		}
		if !match {
			break
		}

		if !discard {
			arg := VaUintptr(args)
			*(*uintptr)(unsafe.Pointer(arg)) = uintptr(n)
		}
		nvalues = 1
	case 'n':
		format++
		panic(todo(""))
	default:
		panic(todo("%#U", c))
	}

	return format, nvalues, match
}

func skipReaderWhiteSpace(r io.ByteScanner) error {
	for {
		c, err := r.ReadByte()
		if err != nil {
			return err
		}

		switch c {
		case ' ', '\t', '\n', '\r', '\v', '\f':
		default:
			r.UnreadByte()
			return nil
		}
	}
}

func skipWhiteSpace(s uintptr) uintptr {
	for {
		switch c := *(*byte)(unsafe.Pointer(s)); c {
		case ' ', '\t', '\n', '\r', '\v', '\f':
			s++
		default:
			return s
		}
	}
}

func fpLiteral(rd io.ByteScanner) (seq []byte) {
	const endOfText = 0x110000
	var pos, width, length int

	defer func() {
		if len(seq) > length {
			rd.UnreadByte()
			seq = seq[:len(seq)-1]
		}
	}()

	var r rune
	step := func(pos int) (rune, int) {
		b, err := rd.ReadByte()
		if err != nil {
			return endOfText, 0
		}

		seq = append(seq, b)
		return rune(b), 1
	}
	move := func() {
		pos += width
		if r != endOfText {
			r, width = step(pos + width)
		}
	}
	accept := func(x rune) bool {
		if r == x {
			move()
			return true
		}
		return false
	}
	accept2 := func(x rune) bool {
		if r <= x {
			move()
			return true
		}
		return false
	}
	r = endOfText
	width = 0
	r, width = step(pos)
	if accept('.') {
		goto l7
	}
	if accept('+') {
		goto l30
	}
	if accept('-') {
		goto l30
	}
	if r < '0' {
		goto l4out
	}
	if accept2('9') {
		goto l35
	}
l4out:
	return seq
l7:
	if r < '0' {
		goto l7out
	}
	if accept2('9') {
		goto l10
	}
l7out:
	return seq
l10:
	length = pos
	if accept('E') {
		goto l18
	}
	if accept('e') {
		goto l18
	}
	if r < '0' {
		goto l15out
	}
	if accept2('9') {
		goto l10
	}
l15out:
	return seq
l18:
	if accept('+') {
		goto l23
	}
	if accept('-') {
		goto l23
	}
	if r < '0' {
		goto l20out
	}
	if accept2('9') {
		goto l26
	}
l20out:
	return seq
l23:
	if r < '0' {
		goto l23out
	}
	if accept2('9') {
		goto l26
	}
l23out:
	return seq
l26:
	length = pos
	if r < '0' {
		goto l27out
	}
	if accept2('9') {
		goto l26
	}
l27out:
	return seq
l30:
	if accept('.') {
		goto l7
	}
	if r < '0' {
		goto l32out
	}
	if accept2('9') {
		goto l35
	}
l32out:
	return seq
l35:
	length = pos
	if accept('.') {
		goto l7
	}
	if accept('E') {
		goto l18
	}
	if accept('e') {
		goto l18
	}
	if r < '0' {
		goto l42out
	}
	if accept2('9') {
		goto l35
	}
l42out:
	return seq
}
