// Copyright 2020 The Libc Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !(linux && (amd64 || arm64 || loong64))

package libc // import "modernc.org/sqlite/internal/libc"

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

const (
	modNone = iota
	modHH
	modH
	modL
	modLL
	modLD
	modQ
	modCapitalL
	modJ
	modZ
	modCapitalZ
	modT
	mod32
	mod64
)

func printf(format, args uintptr) []byte {
	buf := bytes.NewBuffer(nil)
	for {
		switch c := *(*byte)(unsafe.Pointer(format)); c {
		case '%':
			format = printfConversion(buf, format, &args)
		case 0:
			return buf.Bytes()
		default:
			format++
			buf.WriteByte(c)
		}
	}
}

func printfConversion(buf *bytes.Buffer, format uintptr, args *uintptr) uintptr {
	format++
	spec := "%"

flags:
	for {
		switch c := *(*byte)(unsafe.Pointer(format)); c {
		case '#':
			format++
			spec += "#"
		case '0':
			format++
			spec += "0"
		case '-':
			format++
			spec += "-"
		case ' ':
			format++
			spec += " "
		case '+':
			format++
			spec += "+"
		default:
			break flags
		}
	}
	format, width, hasWidth := parseFieldWidth(format, args)
	if hasWidth {
		spec += strconv.Itoa(width)
	}
	format, prec, hasPrecision := parsePrecision(format, args)
	format, mod := parseLengthModifier(format)

	var str string

more:
	switch c := *(*byte)(unsafe.Pointer(format)); c {
	case 'd', 'i':
		format++
		var arg int64
		if isWindows && mod == modL {
			mod = modNone
		}
		switch mod {
		case modL, modLL, mod64, modJ:
			arg = VaInt64(args)
		case modH:
			arg = int64(int16(VaInt32(args)))
		case modHH:
			arg = int64(int8(VaInt32(args)))
		case mod32, modNone:
			arg = int64(VaInt32(args))
		case modT:
			arg = int64(VaInt64(args))
		default:
			panic(todo("", mod))
		}

		if arg == 0 && hasPrecision && prec == 0 {
			break
		}

		if hasPrecision {
			panic(todo("", prec))
		}

		f := spec + "d"
		str = fmt.Sprintf(f, arg)
	case 'u':
		format++
		var arg uint64
		if isWindows && mod == modL {
			mod = modNone
		}
		switch mod {
		case modNone:
			arg = uint64(VaUint32(args))
		case modL, modLL, mod64:
			arg = VaUint64(args)
		case modH:
			arg = uint64(uint16(VaInt32(args)))
		case modHH:
			arg = uint64(uint8(VaInt32(args)))
		case mod32:
			arg = uint64(VaInt32(args))
		case modZ:
			arg = uint64(VaInt64(args))
		default:
			panic(todo("", mod))
		}

		if arg == 0 && hasPrecision && prec == 0 {
			break
		}

		if hasPrecision {
			panic(todo("", prec))
		}

		f := spec + "d"
		str = fmt.Sprintf(f, arg)
	case 'o':
		format++
		var arg uint64
		if isWindows && mod == modL {
			mod = modNone
		}
		switch mod {
		case modNone:
			arg = uint64(VaUint32(args))
		case modL, modLL, mod64:
			arg = VaUint64(args)
		case modH:
			arg = uint64(uint16(VaInt32(args)))
		case modHH:
			arg = uint64(uint8(VaInt32(args)))
		case mod32:
			arg = uint64(VaInt32(args))
		default:
			panic(todo("", mod))
		}

		if arg == 0 && hasPrecision && prec == 0 {
			break
		}

		if hasPrecision {
			panic(todo("", prec))
		}

		f := spec + "o"
		str = fmt.Sprintf(f, arg)
	case 'b':
		format++
		var arg uint64
		if isWindows && mod == modL {
			mod = modNone
		}
		switch mod {
		case modNone:
			arg = uint64(VaUint32(args))
		case modL, modLL, mod64:
			arg = VaUint64(args)
		case modH:
			arg = uint64(uint16(VaInt32(args)))
		case modHH:
			arg = uint64(uint8(VaInt32(args)))
		case mod32:
			arg = uint64(VaInt32(args))
		default:
			panic(todo("", mod))
		}

		if arg == 0 && hasPrecision && prec == 0 {
			break
		}

		if hasPrecision {
			panic(todo("", prec))
		}

		f := spec + "b"
		str = fmt.Sprintf(f, arg)
	case 'I':
		if !isWindows {
			panic(todo("%#U", c))
		}

		format++
		switch c = *(*byte)(unsafe.Pointer(format)); c {
		case 'x', 'X':
			if unsafe.Sizeof(int(0)) == 4 {
				mod = mod32
			}
		case '3':
			format++
			switch c = *(*byte)(unsafe.Pointer(format)); c {
			case '2':
				format++
				mod = mod32
				goto more
			default:
				panic(todo("%#U", c))
			}
		case '6':
			format++
			switch c = *(*byte)(unsafe.Pointer(format)); c {
			case '4':
				format++
				mod = mod64
				goto more
			default:
				panic(todo("%#U", c))
			}
		default:
			panic(todo("%#U", c))
		}
		fallthrough
	case 'X':
		fallthrough
	case 'x':
		format++
		var arg uint64
		if isWindows && mod == modL {
			mod = modNone
		}
		switch mod {
		case modNone:
			arg = uint64(VaUint32(args))
		case modL, modLL, mod64:
			arg = VaUint64(args)
		case modH:
			arg = uint64(uint16(VaInt32(args)))
		case modHH:
			arg = uint64(uint8(VaInt32(args)))
		case mod32:
			arg = uint64(VaInt32(args))
		case modZ:
			arg = uint64(VaInt64(args))
		default:
			panic(todo("", mod))
		}

		if arg == 0 && hasPrecision && prec == 0 {
			break
		}

		if strings.Contains(spec, "#") && arg == 0 {
			spec = strings.ReplaceAll(spec, "#", "")
		}
		var f string
		switch {
		case hasPrecision:
			f = fmt.Sprintf("%s.%d%c", spec, prec, c)
		default:
			f = spec + string(c)
		}
		str = fmt.Sprintf(f, arg)
	case 'e', 'E':
		format++
		arg := VaFloat64(args)
		if !hasPrecision {
			prec = 6
		}
		f := fmt.Sprintf("%s.%d%c", spec, prec, c)
		str = fmt.Sprintf(f, arg)
	case 'f', 'F':
		format++
		arg := VaFloat64(args)
		if !hasPrecision {
			prec = 6
		}
		f := fmt.Sprintf("%s.%d%c", spec, prec, c)
		str = fixNanInf(fmt.Sprintf(f, arg))
	case 'G':
		fallthrough
	case 'g':
		format++
		arg := VaFloat64(args)
		if !hasPrecision {
			prec = 6
		}
		if prec == 0 {
			prec = 1
		}

		f := fmt.Sprintf("%s.%d%c", spec, prec, c)
		str = fixNanInf(fmt.Sprintf(f, arg))
	case 's':
		format++
		arg := VaUintptr(args)
		switch mod {
		case modNone:
			var f string
			switch {
			case hasPrecision:
				f = fmt.Sprintf("%s.%ds", spec, prec)
				str = fmt.Sprintf(f, GoString(arg))
			default:
				f = spec + "s"
				str = fmt.Sprintf(f, GoString(arg))
			}
		default:
			panic(todo(""))
		}
	case 'p':
		format++
		switch runtime.GOOS {
		case "windows":
			switch runtime.GOARCH {
			case "386", "arm":
				fmt.Fprintf(buf, "%08X", VaUintptr(args))
			default:
				fmt.Fprintf(buf, "%016X", VaUintptr(args))
			}
		default:
			fmt.Fprintf(buf, "%#0x", VaUintptr(args))
		}
	case 'c':
		format++
		switch mod {
		case modNone:
			arg := VaInt32(args)
			buf.WriteByte(byte(arg))
		default:
			panic(todo(""))
		}
	case '%':
		format++
		buf.WriteByte('%')
	default:
		panic(todo("%#U", c))
	}

	buf.WriteString(str)
	return format
}

func parseFieldWidth(format uintptr, args *uintptr) (_ uintptr, n int, ok bool) {
	first := true
	for {
		var digit int
		switch c := *(*byte)(unsafe.Pointer(format)); {
		case first && c == '0':
			return format, n, ok
		case first && c == '*':
			format++
			switch c := *(*byte)(unsafe.Pointer(format)); {
			case c >= '0' && c <= '9':
				panic(todo(""))
			default:
				return format, int(VaInt32(args)), true
			}
		case c >= '0' && c <= '9':
			format++
			ok = true
			first = false
			digit = int(c) - '0'
		default:
			return format, n, ok
		}

		n0 := n
		n = 10*n + digit
		if n < n0 {
			panic(todo(""))
		}
	}
}

func parsePrecision(format uintptr, args *uintptr) (_ uintptr, n int, ok bool) {
	for {
		switch c := *(*byte)(unsafe.Pointer(format)); c {
		case '.':
			format++
			first := true
			for {
				switch c := *(*byte)(unsafe.Pointer(format)); {
				case first && c == '*':
					format++
					n = int(VaInt32(args))
					return format, n, true
				case c >= '0' && c <= '9':
					format++
					first = false
					n0 := n
					n = 10*n + (int(c) - '0')
					if n < n0 {
						panic(todo(""))
					}
				default:
					return format, n, true
				}
			}
		default:
			return format, 0, false
		}
	}
}

func parseLengthModifier(format uintptr) (_ uintptr, n int) {
	switch c := *(*byte)(unsafe.Pointer(format)); c {
	case 'h':
		format++
		n = modH
		switch c := *(*byte)(unsafe.Pointer(format)); c {
		case 'h':
			format++
			n = modHH
		}
		return format, n
	case 'l':
		format++
		n = modL
		switch c := *(*byte)(unsafe.Pointer(format)); c {
		case 'l':
			format++
			n = modLL
		}
		return format, n
	case 'q':
		panic(todo(""))
	case 'L':
		format++
		n = modLD
		return format, n
	case 'j':
		format++
		n = modJ
		return format, n
	case 'z':
		format++
		return format, modZ
	case 'Z':
		format++
		return format, modCapitalZ
	case 't':
		format++
		return format, modT
	default:
		return format, 0
	}
}

func fixNanInf(s string) string {
	switch s {
	case "NaN":
		return "nan"
	case "+Inf", "-Inf":
		return "inf"
	default:
		return s
	}
}
