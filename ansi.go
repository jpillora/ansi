//Implements the ANSI VT100 control set.
//Please refer to http://www.termsys.demon.co.uk/vtansi.htm
package ansi

import (
	"io"
	"regexp"
	"strconv"
	"strings"
)

//Ansi represents a wrapped io.ReadWriteCloser.
//It will read the stream, parse and remove ANSI report codes
//and place them on the Reports queue.
type Ansi struct {
	rwc     io.ReadWriteCloser
	rerr    error
	rbuff   chan []byte
	Reports chan *Report
}

//Wrap an io.ReadWriteCloser (like a net.Conn) to
//easily read and write control codes
func Wrap(rwc io.ReadWriteCloser) *Ansi {
	a := &Ansi{}
	a.rwc = rwc
	a.rbuff = make(chan []byte)
	a.Reports = make(chan *Report)
	go a.read()
	return a
}

var reportCode = regexp.MustCompile(`\[([^a-zA-Z]*)(0c|0n|3n|R)`)

//reads the underlying ReadWriteCloser for real,
//extracts the ansi codes, places the rest
//in the read buffer
func (a *Ansi) read() {
	buff := make([]byte, 0xffff)
	for {
		n, err := a.rwc.Read(buff)
		if err != nil {
			a.rerr = err
			close(a.rbuff)
			break
		}

		var src = buff[:n]
		var dst []byte

		//contain ansi codes?
		m := reportCode.FindAllStringSubmatchIndex(string(src), -1)

		if len(m) == 0 {
			dst = make([]byte, n)
			copy(dst, src)
		} else {
			for _, i := range m {
				//slice off ansi code body and trailing char
				a.parse(string(src[i[2]:i[3]]), string(src[i[4]:i[5]]))
				//add surrounding bits to dst buffer
				dst = append(dst, src[:i[0]]...)
				dst = append(dst, src[i[1]:]...)
			}
			if len(dst) == 0 {
				continue
			}
		}

		a.rbuff <- dst
	}
}

// Report Device Code	<ESC>[{code}0c
// Report Device OK	<ESC>[0n
// Report Device Failure	<ESC>[3n
// Report Cursor Position	<ESC>[{ROW};{COLUMN}R
func (a *Ansi) parse(body, char string) {
	r := &Report{}
	switch char {
	case "0c":
		r.Type = Code
		r.Code, _ = strconv.Atoi(body)
	case "0n":
		r.Type = OK
	case "3n":
		r.Type = Failure
	case "R":
		r.Type = Position
		pair := strings.Split(body, ";")
		r.Pos.Col, _ = strconv.Atoi(pair[1])
		r.Pos.Row, _ = strconv.Atoi(pair[0])
	default:
		return
	}
	// fmt.Printf("parsed report: %+v", r)
	a.Reports <- r
}

//Reads the underlying ReadWriteCloser
func (a *Ansi) Read(dest []byte) (n int, err error) {
	//It doesn't really read the underlying ReadWriteCloser :)
	if a.rerr != nil {
		return 0, a.rerr
	}
	src, open := <-a.rbuff
	if !open {
		return 0, a.rerr
	}
	return copy(dest, src), nil
}

//Writes the underlying ReadWriteCloser
func (a *Ansi) Write(p []byte) (n int, err error) {
	return a.rwc.Write(p)
}

//Closes the underlying ReadWriteCloser
func (a *Ansi) Close() error {
	return a.rwc.Close()
}

//==============================

type ReportType int

const (
	Code ReportType = iota
	OK
	Failure
	Position
)

type Report struct {
	Type ReportType
	Code int
	Pos  struct {
		Row, Col int
	}
}

//==============================

const Esc = byte(27)

var QueryCode = []byte{Esc, '[', 'c'}

// Query Device Status	<ESC>[5n
// Query Cursor Position	<ESC>[6n

var QueryCursorPosition = []byte{Esc, '[', '6', 'n'}

func (a *Ansi) QueryCursorPosition() {
	a.Write(QueryCursorPosition)
}

// Reset Device		<ESC>c

var EnableLineWrap = []byte{Esc, '[', '7', 'h'}

func (a *Ansi) EnableLineWrap() {
	a.Write(DisableLineWrap)
}

var DisableLineWrap = []byte{Esc, '[', '7', 'l'}

func (a *Ansi) DisableLineWrap() {
	a.Write(DisableLineWrap)
}

// Font Set G0		<ESC>(
// Font Set G1		<ESC>)

// Cursor Home 		<ESC>[{ROW};{COLUMN}H
// Cursor Up		<ESC>[{COUNT}A
// Cursor Down		<ESC>[{COUNT}B
// Cursor Forward		<ESC>[{COUNT}C
// Cursor Backward		<ESC>[{COUNT}D
// Force Cursor Position	<ESC>[{ROW};{COLUMN}f
func Goto(r, c uint16) []byte {
	rb := []byte(strconv.Itoa(int(r)))
	cb := []byte(strconv.Itoa(int(c)))
	b := append([]byte{Esc, '['}, rb...)
	b = append(b, ';')
	b = append(b, cb...)
	b = append(b, 'f')
	return b
}

func (a *Ansi) Goto(r, c uint16) {
	a.Write(Goto(r, c))
}

// Save Cursor		<ESC>[s
// Unsave Cursor		<ESC>[u
// Save Cursor & Attrs	<ESC>7
// Restore Cursor & Attrs	<ESC>8
// Scroll Screen		<ESC>[r
// Scroll Screen		<ESC>[{start};{end}r
// Scroll Down		<ESC>D
// Scroll Up		<ESC>M

var CursorHide = []byte{Esc, '[', '?', '2', '5', 'l'}

func (a *Ansi) CursorHide() {
	a.Write(CursorHide)
}

var CursorShow = []byte{Esc, '[', '?', '2', '5', 'h'}

func (a *Ansi) CursorShow() {
	a.Write(CursorShow)
}

// Tab Control
// Set Tab 		<ESC>H
// Clear Tab 		<ESC>[g
// Clear All Tabs 		<ESC>[3g

// Erase End of Line	<ESC>[K
// Erase Start of Line	<ESC>[1K
// Erase Line		<ESC>[2K
// Erase Down		<ESC>[J
// Erase Up		<ESC>[1J

var EraseScreen = []byte{Esc, '[', '2', 'J'}

func (a *Ansi) EraseScreen() {
	a.Write(EraseScreen)
}

// Printing
// Print Screen		<ESC>[i
// Print Line		<ESC>[1i
// Stop Print Log		<ESC>[4i
// Start Print Log		<ESC>[5i

// Set Key Definition	<ESC>[{key};"{ascii}"p

// Sets multiple display attribute settings. The following lists standard attributes:
type Attribute string

var (
	Reset      Attribute = "0"
	Bright     Attribute = "1"
	Dim        Attribute = "2"
	Italic     Attribute = "3"
	Underscore Attribute = "4"
	Blink      Attribute = "5"
	Reverse    Attribute = "7"
	Hidden     Attribute = "8"
)

const (
	Black   Attribute = "30"
	Red     Attribute = "31"
	Green   Attribute = "32"
	Yellow  Attribute = "33"
	Blue    Attribute = "34"
	Magenta Attribute = "35"
	Cyan    Attribute = "36"
	White   Attribute = "37"
)

const (
	BlackBG   Attribute = "40"
	RedBG     Attribute = "41"
	GreenBG   Attribute = "42"
	YellowBG  Attribute = "43"
	BlueBG    Attribute = "44"
	MagentaBG Attribute = "45"
	CyanBG    Attribute = "46"
	WhiteBG   Attribute = "47"
)

//Set attributes
func Set(attrs ...Attribute) []byte {
	s := make([]string, len(attrs))
	for i, a := range attrs {
		s[i] = string(a)
	}
	b := []byte(strings.Join(s, ";"))
	b = append(b, 'm')
	return append([]byte{Esc, '['}, b...)
}

// Set Attribute Mode	<ESC>[{attr1};...;{attrn}m
func (a *Ansi) Set(attrs ...Attribute) {
	a.Write(Set(attrs...))
}
