// +build !solution

package fs

import (
	"context"
	"dat320/lab8/grpcfs/proto"
	"errors"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// OpenRead is passed to Open to indicate a file should be opened as read-only
	OpenRead = iota
	// OpenWrite is passed to Open to indicate a file should be opened as write-only
	OpenWrite
	// OpenReadWrite is passed to Open to indicate a file should be opened as read-write
	OpenReadWrite
)

var (
	// ErrOpenDir occurs if the caller tries to open a directory (only files
	// can be opened since none of the File operations fit directories)
	ErrOpenDir = errors.New("cannot open a directory")
	// ErrOpenInvalidFlag occurs if the caller passes an unknown flag to Open
	ErrOpenInvalidFlag = errors.New("tried to open file with an unrecognized flag")
	errNotImplemented  = errors.New("file does not exist")
)

// FileSystem acts as a pseudo-filesystem over gRPC. Files can be opened by the
// FS, which then interact with the filesystem of the remote server over gRPC
// for their supported operations.
type FileSystem struct {
	// gRPC client - communicates with the gRPC file system server
	client proto.FileSystemClient
	// TODO(student): add necessary fields here (if any)
}

// NewFileSystem returns a FileSystem with an initialized gRPC client connection
func NewFileSystem(addr string, opts ...grpc.DialOption) (*FileSystem, error) {
	var conn grpc.ClientConnInterface
	var err error

	if len(opts) == 0 {
		// if no options are provided, default to insecure connection
		conn, err = grpc.Dial(addr, grpc.WithInsecure())
	} else {
		// use the options provided by the caller
		conn, err = grpc.Dial(addr, opts...)
	}

	if err != nil {
		return nil, err
	}

	fs := &FileSystem{
		client: proto.NewFileSystemClient(conn),
		// TODO(student): initialize additional fields here (if any)
	}

	return fs, nil
}

// Daniel

// Open opens the file `name` at the remote server.
// `flag` must be one of `OpenRead`, `OpenWrite` or `OpenReadWrite`.
// Files opened with `OpenRead` must already exist.
// Calling `Open` with `OpenWrite` or `OpenReadWrite` creates the file on the
// remote server if it does not already exist.
// Calling `Open` with `OpenRead` on a non-existent file returns an error.
// Directories cannot be opened.
func (fs *FileSystem) Open(name string, flag int) (f *File, err error) {
	if flag != OpenRead && flag != OpenWrite && flag != OpenReadWrite {
		return nil, ErrOpenInvalidFlag
	}
	isDir, size, _, err := fs.Lookup(name)
	if isDir {
		return nil, ErrOpenDir
	}
	if flag == OpenRead {
		if err != nil {
			return nil, err
		}
		readerClient, err := fs.client.Reader(context.Background())
		if err != nil {
			return nil, translateStatusError(err)
		}
		return &File{
			path:    name,
			rClient: readerClient,
			size:    int64(0),
		}, nil
	}
	// Flag is Write or ReadWrite
	// If file does not exist: Create file
	if err != nil {
		if errors.Is(translateStatusError(err), os.ErrNotExist) {

			_, err = fs.client.Create(context.Background(), &proto.CreateRequest{
				Path: name,
			})
			if err != nil {
				return nil, translateStatusError(err)
			}
		} else {
			return nil, translateStatusError(err)
		}
	}

	writerClient, err := fs.client.Writer(context.Background())
	if err != nil {
		return nil, translateStatusError(err)
	}
	if flag == OpenReadWrite {
		// If flag is ReadWrite also make reader
		readerClient, err := fs.client.Reader(context.Background())
		if err != nil {
			return nil, translateStatusError(err)
		}
		return &File{
			path:    name,
			rClient: readerClient,
			wClient: writerClient,
			size:    int64(size),
		}, nil
	}
	return &File{
		path:    name,
		wClient: writerClient,
		size:    int64(size),
	}, nil
}

// Remove removes the file/directory at the remote server
func (fs *FileSystem) Remove(name string) error {
	_, err := fs.client.Remove(context.Background(), &proto.RemoveRequest{Path: name})
	return translateStatusError(err)
}

// Withya
// PASS

// Mkdir makes a directory at the remote server
func (fs *FileSystem) Mkdir(path string, perm os.FileMode) error {
	_, err := fs.client.Mkdir(context.Background(), &proto.MakeDirRequest{Path: path})
	if err != nil {
		return translateStatusError(err)
	}
	return nil
}

// Withya

// Lookup looks up the path at the remote server and returns any information it receives
func (fs *FileSystem) Lookup(path string) (isDir bool, size int, files []string, err error) {

	response, errr := fs.client.Lookup(context.Background(), &proto.LookupRequest{Path: path})
	if errr != nil {
		return false, 0, nil, translateStatusError(errr)
	}
	return response.IsDir, int(response.Size), response.Files, nil
	//return false, 0, nil, errNotImplemented
	//FINN UT AV HVORFOR DU FÅR FEIL!
}

// translateStatusError translates status errors to errors from the `os` package
// by matching the codes to those defined by the server
func translateStatusError(err error) error {
	if _, isStatusError := status.FromError(err); isStatusError {
		code := status.Code(err)
		switch code {
		case codes.NotFound:
			err = os.ErrNotExist
		case codes.AlreadyExists:
			err = os.ErrExist
		case codes.FailedPrecondition:
			err = os.ErrClosed
		case codes.PermissionDenied:
			err = os.ErrPermission
		case codes.InvalidArgument:
			err = os.ErrInvalid
		default:
			// Workaround. For some reason these are sometimes not
			// matched at the server-side.
			if err != nil && strings.Contains(err.Error(), os.ErrNotExist.Error()) {
				err = os.ErrNotExist
			} else if (err != nil && strings.Contains(err.Error(), os.ErrExist.Error())) ||
				(err != nil && strings.Contains(err.Error(), "already exists")) {
				err = os.ErrExist
			}
		}
	}

	return err
}
