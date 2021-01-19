// +build !solution

package fsserver

import (
	"context"
	"dat320/lab8/grpcfs/proto"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrNilRequest occurs when a request is nil despite no error having occurred.
	ErrNilRequest = status.Error(codes.InvalidArgument, "request is nil")
	// ErrStreamChangedPath is produced when a Reader/Writer changes the path between a sequence
	// of messages in the same stream. The path must be constant.
	ErrStreamChangedPath = status.Error(codes.InvalidArgument, "path has changed from a previous request in the stream")
	errNotImplemented    = errors.New("this feature is not yet implemented")
)

// FileSystemServer performs operations on a local, in-memory filesystem in response to RPCs
type FileSystemServer struct {
	// in-memory filesystem
	fs *memfs.MemFS
	// this must be included in implementers of the proto.FileSystemServer interface
	proto.UnimplementedFileSystemServer
	// TODO(student): add necessary fields here (if any)
}

// NewFileSystemServer returns an initialized FileSystemServer
func NewFileSystemServer() *FileSystemServer {
	return &FileSystemServer{
		fs: memfs.Create(),
		// TODO(student): initialize additional fields here (if any)
	}
}

// Daniel

// Lookup produces information about the file/directory at the requested path.
// For regular files, the size in bytes is produced.
// For directories, a list of files/subdirectories within the looked up directory
// is produced, along with the number of elements inside the directory.
func (s *FileSystemServer) Lookup(ctx context.Context, r *proto.LookupRequest) (*proto.LookupResponse, error) {
	entryInfo, err := s.fs.Lstat(r.Path)
	if err != nil {
		return nil, createStatusError(err)
	}
	isDir := entryInfo.IsDir()
	var size int64
	var files []string
	if isDir {
		directory, err := s.fs.ReadDir(r.Path)
		if err != nil {
			return nil, createStatusError(err)
		}
		size = int64(len(directory))
		files = make([]string, size)
		for i, file := range directory {
			files[i] = file.Name()
		}
	} else {
		size = entryInfo.Size()
	}
	response := &proto.LookupResponse{
		IsDir: isDir,
		Size:  size,
		Files: files,
	}
	return response, nil
}

// Daniel

// Create creates a file at the requested path.
// An error is produced if there is already a file/directory at the path.
func (s *FileSystemServer) Create(ctx context.Context, r *proto.CreateRequest) (*proto.CreateResponse, error) {
	flag := os.O_CREATE + os.O_EXCL
	_, err := s.fs.OpenFile(r.Path, flag, 0666)
	if err != nil {
		return nil, createStatusError(err)
	}
	return new(proto.CreateResponse), nil
}

// Daniel

// Reader receives a stream of read requests for some file from the client, and
// responds to each request with the contents read at the file in the local FS.
// The requested file must be constant across the stream.
// The stream is not closed on EOF errors. Instead the server responds with the
// number of bytes read (in that call) prior to the EOF condition, and a
// notification to the client indicating an EOF condition occurred.
func (s *FileSystemServer) Reader(stream proto.FileSystem_ReaderServer) error {
	// TODO(student): implement this method
	// TODO Daniel: make sure this works
	var readContent []byte
	var file vfs.File
	var originalPath string
	for {
		in, err := stream.Recv()
		if err != nil {
			return err
		}
		if file == nil {
			// If first request: open file and seek to offset.
			file, err = s.fs.OpenFile(in.Path, os.O_RDONLY, 0666)
			originalPath = in.Path
			if err != nil {
				return createStatusError(err)
			}
			// Whence = 0: START OF FILE
			_, err := file.Seek(in.Offset, 0)
			if err != nil {
				return createStatusError(err)
			}
		} else {
			if originalPath != in.Path {
				return createStatusError(ErrStreamChangedPath)
			}
		}
		readContent = make([]byte, in.NumBytes)
		n, err := file.Read(readContent)
		if errors.Is(err, io.EOF) {
			readContent = readContent[0:n]
			err := stream.Send(&proto.ReadResponse{
				ReadContent: readContent,
				Eof:         true,
				BytesRead:   int64(n),
			})
			if err != nil {
				return createStatusError(err)
			}
		} else if err != nil {
			return createStatusError(err)
		}
		err = stream.Send(&proto.ReadResponse{
			ReadContent: readContent,
			Eof:         false,
			BytesRead:   int64(n),
		})
		if err != nil {
			return createStatusError(err)
		}

	}
}

// Withya

// Writer receives a stream of write requests to some file from the client,
// writes the contents of the request to the file in the local FS, and responds
// with the status of the request.
// The requested file must be constant across the stream.
// If the client tries to write at an offset past the end of the file, the
// status will contain an EOF notification, and the stream remains open.
func (s *FileSystemServer) Writer(stream proto.FileSystem_WriterServer) error {
	// TODO(student): implement this method
	var file vfs.File
	var init_path string

	for {

		request, err := stream.Recv()
		if err != nil {
			return err
		}

		if file == nil {
			init_path = request.Path
			file, err = s.fs.OpenFile(request.Path, os.O_WRONLY, 0666)
			if err != nil {
				return createStatusError(err)
			}
			//file.Seek(request.Offset, 0)

		} else {
			if init_path != request.Path {
				return createStatusError(ErrStreamChangedPath)
			}
		}

		file.Seek(request.Offset, 0)

		numBytes, errr := file.Write(request.Content)
		if errr != nil {
			if errors.Is(errr, io.EOF) {
				stream.Send(&proto.WriteStatus{
					NumBytes: int64(numBytes),
					Eof:      true,
				})
			} else {
				return errr
			}

		} else {
			stream.Send(&proto.WriteStatus{
				NumBytes: int64(numBytes),
				Eof:      false,
			})
		}

	}

	return nil
}

// Withya

// Remove removes the file or recursively removes the directory at the requested
// path in the local FS.
func (s *FileSystemServer) Remove(ctx context.Context, r *proto.RemoveRequest) (*proto.RemoveResponse, error) {

	err := s.fs.Remove(r.Path)
	if err != nil {
		return nil, createStatusError(err)
	}

	return new(proto.RemoveResponse), nil
}

// Withya

// Mkdir makes a directory at the requested path in the local FS.
func (s *FileSystemServer) Mkdir(ctx context.Context, r *proto.MakeDirRequest) (*proto.MakeDirResponse, error) {

	err := s.fs.Mkdir(r.Path, os.ModeDir) //os.FileMode
	if err != nil {
		return nil, createStatusError(err)
	}
	return new(proto.MakeDirResponse), nil
}

// createStatusError converts an error to a gRPC status error. This causes the
// error to include a status code when sent over gRPC, which can be checked by
// the receiver to determine which type of error occurred.
func createStatusError(err error) error {
	// unwrap os.PathError if it is of this type
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		err = pathErr.Err
	}

	// NOTE: The strings.Contains matching is a workaround. Sometimes the
	// errors are not matched as expected.
	if errors.Is(err, os.ErrNotExist) || (err != nil && strings.Contains(err.Error(), os.ErrNotExist.Error())) {
		return status.Error(codes.NotFound, err.Error())
	} else if errors.Is(err, os.ErrExist) ||
		// (blang/vfs uses fmt.Errorf for failed Mkdir attempts)
		(err != nil && strings.Contains(err.Error(), os.ErrExist.Error())) ||
		(err != nil && strings.Contains(err.Error(), "already exists")) {
		return status.Error(codes.AlreadyExists, err.Error())
	} else if errors.Is(err, os.ErrClosed) {
		return status.Error(codes.FailedPrecondition, err.Error())
	} else if errors.Is(err, os.ErrPermission) {
		return status.Error(codes.PermissionDenied, err.Error())
	} else if errors.Is(err, os.ErrInvalid) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	return err
}
