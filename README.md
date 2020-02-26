## ansi

Implements the ANSI VT100 control set.
Please refer to http://www.termsys.demon.co.uk/vtansi.htm

[![GoDoc](https://godoc.org/github.com/jpillora/ansi?status.svg)](https://pkg.go.dev/github.com/jpillora/ansi?tab=doc)

### Install

```
go get github.com/jpillora/ansi
```

### Usage

Get ANSI control code bytes:

``` go
ansi.Goto(2,4)
ansi.Set(ansi.Green, ansi.BlueBG)
```

Wrap an `io.ReadWriteCloser`:

``` go

a := ansi.Wrap(tcpConn)

//Read, Write, Close as normal
a.Read()
a.Write()
a.Close()

//Shorthand for a.Write(ansi.Set(..))
a.Set(ansi.Green, ansi.BlueBG)

//Send query
a.QueryCursorPosition()
//Await report
report := <- a.Reports
report.Type//=> ansi.Position
report.Pos.Row
report.Pos.Col
```

*Wrapped connections will intercept and remove ANSI report codes from `a.Read()`*

### API

https://pkg.go.dev/github.com/jpillora/ansi?tab=doc