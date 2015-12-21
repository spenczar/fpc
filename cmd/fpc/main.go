package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/spenczar/fpc"
)

const bufferSize = 1024

func main() {
	decompress := flag.Bool("d", false, "Decompress input data and write output to stdout.")
	level := flag.Int("l", fpc.DefaultCompression, "Compression level to use when compressing. Ignored when decompressing.")
	help := flag.Bool("h", false, "Print this help text")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *decompress {
		decompressStream(os.Stdin, os.Stdout)
	} else {
		compressStream(os.Stdin, os.Stdout, *level)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "fatal: %s\n", err.Error())
	os.Exit(1)
}

func compressStream(in io.Reader, out io.Writer, level int) {
	w, err := fpc.NewWriterLevel(out, level)
	if err != nil {
		fatal(err)
	}

	buf := make([]byte, bufferSize)
	for {
		n, err := in.Read(buf)
		//log.Printf("bytes off wire: %#v", buf[:n])
		if err == io.EOF {
			w.Write(buf[:n])
			err = w.Close()
			if err != nil {
				fatal(err)
			}
			return
		} else if err != nil {
			fatal(err)
		}
		w.Write(buf[:n])
	}
}

func decompressStream(in io.Reader, out io.Writer) {
	r := fpc.NewReader(os.Stdin)
	buf := make([]byte, bufferSize)
	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			out.Write(buf[:n])
			return
		} else if err != nil {
			fatal(err)
		}
		out.Write(buf[:n])
	}
}
