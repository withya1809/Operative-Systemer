// +build !solution

package fs

import (
	"dat320/lab8/grpcfs/proto"
	"errors"
	"io"
)

var (
	// ErrSeekNegativeOffset occurs when Seek would place the new offset before the start of the file
	ErrSeekNegativeOffset = errors.New("Seek would place offset before the start of the file")
	// ErrSeekInvalidWhence occurs when the whence parameter to Seek is not recognized
	ErrSeekInvalidWhence = errors.New("the whence parameter must be one of io.SeekStart, io.SeekCurrent or io.SeekEnd")
)

// File is an abstraction of a single file in the pseudo-filesystem.
// It contains methods similar to those in the os package, but instead communicates
// with a remote server over gRPC in which the file operations take place.
type File struct {
	// the path to the file at the remote FS
	path string
	// may be used for Lookup requests
	client proto.FileSystemClient
	// communicates Read operations
	rClient proto.FileSystem_ReaderClient
	// communicates Write operations
	wClient proto.FileSystem_WriterClient
	// TODO(student): add necessary fields here (if any)
	currentOffset int64
	size          int64
	//isClosed      bool
}

// CloseErr contains errors related to closing reader/writer streams to the
// remote server
type CloseErr struct {
	ReaderErr error
	WriterErr error
}

// Error implements the error interface
func (c CloseErr) Error() string {
	return "failed to close either the reader, the writer, or both"
}

// Withya

// Close closes the read/write stream(s) related to the file and returns
// related errors, if any. It implements the io.Closer interface.
func (f *File) Close() error {

	/*f.isClosed = false

	for f.isClosed == false {

		if f.wClient != nil {
			err := f.wClient.CloseSend()
			if err != nil {
				return os.ErrClosed
			} else {
				f.isClosed = true
			}
		}

		if f.rClient != nil {
			err := f.rClient.CloseSend()
			if err != nil {
				return os.ErrClosed
			} else {
				f.isClosed = true
			}
		}

	}
	*/

	return nil
}

// Daniel

// Read reads up to `len(p)` bytes from the file at the remote server into `p`.
// It implements the io.Reader interface.
//
// Each call to Read moves the offset in the file by the number of bytes read.
// As such, successive calls will read from the offset after the position of the
// final byte read in the previous call.
// `n` is the number of bytes read.
func (f *File) Read(p []byte) (n int, err error) {
	// TODO(student): implement this method
	// TODO Daniel: make sure this works
	if f.currentOffset > f.size {
		// Request can not be completed until offset is updated due to offset beeing at EOF
		return 0, io.EOF
	}
	numRead := 0
	maxBytesInEachBatch := 255
	for numRead < len(p) {
		var numBytes int
		if len(p)-numRead < maxBytesInEachBatch {
			numBytes = len(p) - numRead
		} else {
			numBytes = maxBytesInEachBatch
		}
		err := f.rClient.Send(&proto.ReadRequest{
			Path:     f.path,
			Offset:   f.currentOffset,
			NumBytes: int64(numBytes),
		})
		if err != nil {
			return numRead, translateStatusError(err)
		}
		response, err := f.rClient.Recv()
		if err != nil {
			return numRead, translateStatusError(err)
		}
		for i := 0; i < int(response.BytesRead); i++ {
			p[numRead] = response.ReadContent[i]
			numRead++
			f.currentOffset++
		}
		if response.Eof {
			return numRead, io.EOF
		}
	}
	return numRead, nil
}

// Withya

// Write writes up to `len(p)` bytes from `p` into the file at the remote server.
// It implements the io.Writer interface.
//
// Each call to Write moves the offset in the file by the number of bytes written.
// As such, successive calls will write at the offset after the possition of the
// final byte written in the previous call.
// `n` is the number of bytes that was written.
func (f *File) Write(p []byte) (n int, err error) {
	// TODO(student): implement this method
	if f.currentOffset > f.size {
		// Request can not be completed until offset is updated due to offset beeing at EOF
		return 0, io.EOF
	}
	numWritten := 0

	errr := f.wClient.Send(&proto.WriteRequest{
		Path:    f.path,
		Offset:  f.currentOffset,
		Content: p,
	})
	if errr != nil {
		return numWritten, translateStatusError(errr)
	}

	response, er := f.wClient.Recv()
	if er != nil {
		return numWritten, translateStatusError(er)
	}

	f.currentOffset += response.NumBytes

	// if writing new bytes to file: update size
	if f.currentOffset-f.size > 0 {
		f.size += f.currentOffset - f.size
	}

	numWritten += int(response.NumBytes)

	if response.Eof {
		return numWritten, nil
	}

	return numWritten, nil
}

// Daniel

// Seek sets the offset for the next Read or Write according to whence.
// It implements the io.Seeker interface (documentation copied below):
//
// Seek sets the offset for the next Read or Write to offset, interpreted
// according to whence:
// SeekStart means relative to the start of the file,
// SeekCurrent means relative to the current offset, and
// SeekEnd means relative to the end.
// Seek returns the new offset relative to the start of the file and an error, if any.
//
// Seeking to an offset before the start of the file is an error.
// Seeking to any positive offset is legal, but the behavior of subsequent I/O operations
// on the underlying object is implementation-dependent.
//
// Implementation specific: After seeking to a positive offset larger than the
// size of the file, subsequent calls to Read/Write cause an io.EOF error until
// the offset is set to a valid offset within the bounds of the file size.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	// TODO(student): implement this method
	// TODO Daniel: Make sure this is correct
	oldOffset := f.currentOffset
	switch whence {
	case 0:
		// SeekStart
		f.currentOffset = offset
	case 1:
		// SeekCurrent
		f.currentOffset += offset
	case 2:
		// SeekEnd
		f.currentOffset = f.size + offset - 1
	default:
		return f.currentOffset, ErrSeekInvalidWhence
	}
	if f.currentOffset < 0 {
		f.currentOffset = oldOffset
		return f.currentOffset, ErrSeekNegativeOffset
	}
	return f.currentOffset, nil
}
