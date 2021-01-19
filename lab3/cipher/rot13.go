package cipher

import (
	"io"
)

/*
Task 3: Rot 13

This task is taken from http://tour.golang.org.

A common pattern is an io.Reader that wraps another io.Reader, modifying the
stream in some way.

For example, the gzip.NewReader function takes an io.Reader (a stream of
compressed data) and returns a *gzip.Reader that also implements io.Reader (a
stream of the decompressed data).

Implement a rot13Reader that implements io.Reader and reads from an io.Reader,
modifying the stream by applying the rot13 substitution cipher to all
alphabetical characters.

The rot13Reader type is provided for you. Make it an io.Reader by implementing
its Read method.
*/

type rot13Reader struct {
	r io.Reader
}

func rot13(b byte) byte {
	var beg byte

	if b >= 'A' && b <= 'Z' {
		beg = 'A'
	} else if b >= 'a' && b <= 'z' {
		beg = 'a'
	} else {
		return b
	}

	return (((b - beg) + 13) % 26) + beg
}

func (r rot13Reader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	for i:= range(p[:n]){
		p[i]=rot13(p[i])
	}

	return 
}
