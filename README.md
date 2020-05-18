# fpc #

[![GoDoc](http://godoc.org/github.com/spenczar/fpc?status.svg)](http://godoc.org/github.com/spenczar/fpc)

fpc is a Go implementation of Burtscher and Ratanaworabhan's ['FPC' algorithm](http://cs.txstate.edu/~burtscher/research/FPC/) for compressing a stream of floating point data.

## Why? ##

The FPC algorithm can losslessly encode and decode huge amounts of
floating-point data very quickly. It scales well to
gigabyte-per-second streams. Compression ratios are better than just
about any generic compressor like gzip or bzip, and compression and
decompression throughput are much better (like, 8x to 300x faster)
than other algorithms. For more on this,
[the paper introducing FPC](http://cs.txstate.edu/~burtscher/papers/dcc07a.pdf)
is really readable - I highly recommend it!

## Usage ##

fpc provides a `Writer` and a `Reader`, following the pattern set by
the Go standard library's compression packages. The Writer wraps an
io.Writer that you want to write compressed data into, and the Reader
wraps an io.Reader that you want to read compressed data out of.

Since FPC encodes streams of float64s, they impose some additional
expectations on callers: when calling `Reader.Read(p []byte)` or
`Writer.Write(p []byte)`, the length of `p` must be a multiple of 8,
to match the expectation that the bytes represent a stream of 8-byte
float64s.

In addition, utility methods are provided: `Reader` has a
`ReadFloats(fs []float64) (int, error)` method which will read bytes
from its underlying source, parse them as float64s, put them in `fs`,
and return the number of float64s it placed in fs. When it reaches the
end of the compressed stream, it will return `0, io.EOF`.

Similarly, `Writer` has a `WriteFloat(f float64) error` method which
writes a single float64 to the compressed stream.

## Performance ##

In benchmarks on a fairly vanilla laptop, reading or writing from an
in-memory stream, `fpc` is able to encode at about 1.2 gigabytes per
second, and it can decode at about 0.9 gigabytes per
second. Benchmarks can be run on your own hardware with `go test
-bench "Read|Write" .`.
