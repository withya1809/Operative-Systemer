package fsserver

import (
	"dat320/lab8/grpcfs/proto"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// gRPC server used in tests will by default serve at this address:port
	// (:0 indicates we want any available port selected by the system)
	defaultServerAddr = ":0"
	// files/directories are opened with these permissions in the tests
	fileMode = uint32(os.ModePerm)
)

// used in switch-cases in tests to perform sequences of operations, and used
// for formatting functions as strings
const (
	cmdCreate = iota
	cmdMkdir
	cmdRemove
	cmdLookup
	cmdWriteFile
)

var grpcServerAddr = flag.String("addr", defaultServerAddr, "address:port to open gRPC servers during tests")

type fnParams struct {
	// function type
	fn int
	// prepended to function/method name
	prefix string
	// path/filename
	path string
}

func (p fnParams) String() string {
	switch p.fn {
	case cmdLookup:
		return fmt.Sprintf("%sLookup(..., &proto.LookupRequest{Path: \"%s\"})", p.prefix, p.path)
	case cmdCreate:
		return fmt.Sprintf("%sCreate(..., &proto.CreateRequest{Path: \"%s\"})", p.prefix, p.path)
	case cmdMkdir:
		return fmt.Sprintf("%sMkdir(..., &proto.MakeDirRequest{Path: \"%s\", FileMode: %d})", p.prefix, p.path, fileMode)
	case cmdRemove:
		return fmt.Sprintf("%sRemove(..., &proto.RemoveRequest{Path: \"%s\"})", p.prefix, p.path)
	default:
		return "Unknown command"
	}
}

// setupServer sets up a gRPC server which serves in the background if no error occurs
func setupServer(t *testing.T, fsServer proto.FileSystemServer) (*grpc.Server, string, error) {
	t.Helper()
	l, err := net.Listen("tcp", *grpcServerAddr)
	if err != nil {
		return nil, "", err
	}
	s := grpc.NewServer()
	proto.RegisterFileSystemServer(s, fsServer)

	// serve in background if no error
	go func() {
		if err := s.Serve(l); err != nil {
			t.Logf("failed to serve: %v", err)
		}
	}()

	return s, l.Addr().String(), err
}

// setupClient sets up a gRPC client connected to the gRPC server
func setupClient(t *testing.T, addr string) (proto.FileSystemClient, error) {
	t.Helper()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return proto.NewFileSystemClient(conn), nil
}

// used with cmp.Equal or cmp.Diff to compare errors (cmpopts.EquateError did
// not appear to be using errors.Is at the time of writing)
var cmpOptErrorComparer = cmp.Comparer(func(a, b error) bool {
	if (a == nil && b != nil) || (a != nil && b == nil) {
		//  they are not equal if one is nil and the other is not
		return false
	}

	codeA := status.Code(a)
	codeB := status.Code(b)
	if codeA != codes.Unknown && codeB != codes.Unknown && codeA == codeB {
		// we consider two errors to be the same if they both have a
		// gRPC status code which matches
		return true
	}

	return errors.Is(a, b)
})

var createTests = []struct {
	name  string
	paths []string
	errs  []error
	desc  string
}{
	{
		name:  "single_file",
		paths: []string{"test"},
		errs:  []error{nil},
		desc:  "create single file",
	},
	{
		name:  "no_parent_dir",
		paths: []string{"foo/bar"},
		errs:  []error{status.Error(codes.NotFound, os.ErrNotExist.Error())},
		desc:  "try to create a file within a non-existent directory",
	},
	{
		name:  "duplicate_create",
		paths: []string{"test", "test"},
		errs:  []error{nil, status.Error(codes.AlreadyExists, os.ErrExist.Error())},
		desc:  "try to create the same file twice",
	},
	{
		name:  "file_as_parent_dir",
		paths: []string{"foo", "foo/bar"},
		errs:  []error{nil, status.Error(codes.NotFound, os.ErrNotExist.Error())},
		desc:  "try to use file 'foo' as a parent directory for file 'foo/bar'",
	},
	{
		name:  "3_files",
		paths: []string{"a", "b", "c"},
		errs:  []error{nil, nil, nil},
		desc:  "create 3 files",
	},
}

var mkdirTests = []struct {
	name  string
	paths []string
	errs  []error
	desc  string
}{
	{
		name:  "single_dir",
		paths: []string{"test"},
		errs:  []error{nil},
		desc:  "create single directory",
	},
	{
		name:  "subdir_no_parent",
		paths: []string{"foo/bar"},
		errs:  []error{status.Error(codes.NotFound, os.ErrNotExist.Error())},
		desc:  "try to create subdirectory within a non-existent directory",
	},
	{
		name:  "duplicate_dir",
		paths: []string{"test", "test"},
		errs:  []error{nil, status.Error(codes.AlreadyExists, os.ErrExist.Error())},
		desc:  "try to create the same directory twice",
	},
	{
		name:  "subdir",
		paths: []string{"foo", "foo/bar"},
		errs:  []error{nil, nil},
		desc:  "make a directory, then make a subdirectory",
	},
	{
		name:  "3_dirs",
		paths: []string{"a", "b", "c"},
		errs:  []error{nil, nil, nil},
		desc:  "create 3 directories",
	},
}

type rmTestCase struct {
	cmd  int
	path string
	err  error
}

var rmTests = []struct {
	name  string
	steps []rmTestCase
	desc  string
}{
	{
		name: "rm_file_not_exist",
		steps: []rmTestCase{
			{cmd: cmdRemove, path: "test", err: status.Error(codes.NotFound, os.ErrNotExist.Error())},
		},
		desc: "try to remove non-existent file",
	},
	{
		name: "rm_file",
		steps: []rmTestCase{
			{cmd: cmdCreate, path: "test"},
			{cmd: cmdRemove, path: "test"},
		},
		desc: "create and remove a file",
	},
	{
		name: "rm_dir",
		steps: []rmTestCase{
			{cmd: cmdMkdir, path: "test"},
			{cmd: cmdRemove, path: "test"},
		},
		desc: "create and remove a directory",
	},
	{
		name: "rm_subdir",
		steps: []rmTestCase{
			{cmd: cmdMkdir, path: "foo"},
			{cmd: cmdMkdir, path: "foo/bar"},
			{cmd: cmdRemove, path: "foo/bar"},
		},
		desc: "create a directory and subdirectory, remove the subdirectory",
	},
	{
		name: "rm_parent_dir",
		steps: []rmTestCase{
			{cmd: cmdMkdir, path: "foo"},
			{cmd: cmdMkdir, path: "foo/bar"},
			{cmd: cmdRemove, path: "foo"},
		},
		desc: "create a directory and subdirectory, recursively remove the directories (remove the parent directory)",
	},
	{
		name: "rm_file_within_dir",
		steps: []rmTestCase{
			{cmd: cmdMkdir, path: "foo"},
			{cmd: cmdCreate, path: "foo/bar"},
			{cmd: cmdRemove, path: "foo/bar"},
		},
		desc: "create a directory and file within it, remove the file",
	},
	{
		name: "rm_parent_dir_multiple_files",
		steps: []rmTestCase{
			{cmd: cmdMkdir, path: "foo"},
			{cmd: cmdCreate, path: "foo/bar"},
			{cmd: cmdCreate, path: "foo/baz"},
			{cmd: cmdRemove, path: "foo/bar"},
		},
		desc: "create a directory and 2 files within it, recursively remove the parent directory and its contents",
	},
}

type lookupTestCase struct {
	cmd           int
	path, toWrite string
	err           error
	want          *proto.LookupResponse
}

var lookupTests = []struct {
	name  string
	steps []lookupTestCase
	desc  string
}{
	{
		name: "file_not_found",
		steps: []lookupTestCase{
			{cmd: cmdLookup, path: "test", err: status.Error(codes.NotFound, os.ErrNotExist.Error())},
		},
		desc: "lookup a non-existent file",
	},
	{
		name: "simple_dir",
		steps: []lookupTestCase{
			{cmd: cmdMkdir, path: "test"},
			{cmd: cmdLookup, path: "test", want: &proto.LookupResponse{IsDir: true}},
		},
		desc: "create and lookup a directory",
	},
	{
		name: "simple_file",
		steps: []lookupTestCase{
			{cmd: cmdCreate, path: "test"},
			{cmd: cmdLookup, path: "test", want: &proto.LookupResponse{}},
		},
		desc: "create and lookup an empty file",
	},
	{
		name: "subdir",
		steps: []lookupTestCase{
			{cmd: cmdMkdir, path: "foo"},
			{cmd: cmdMkdir, path: "foo/bar"},
			{cmd: cmdLookup, path: "foo/bar", want: &proto.LookupResponse{IsDir: true}},
		},
		desc: "create and lookup the directory foo/bar",
	},
	{
		name: "file_in_dir",
		steps: []lookupTestCase{
			{cmd: cmdMkdir, path: "foo"},
			{cmd: cmdCreate, path: "foo/bar"},
			{cmd: cmdLookup, path: "foo/bar", want: &proto.LookupResponse{}},
		},
		desc: "create and lookup the file foo/bar",
	},
	{
		name: "file_with_content",
		steps: []lookupTestCase{
			{cmd: cmdWriteFile, path: "test", toWrite: "this is a test"},
			{cmd: cmdLookup, path: "test", want: &proto.LookupResponse{Size: int64(len("this is a test"))}},
		},
		desc: "create and write to a file, then look it up",
	},
	{
		name: "dir_with_content",
		steps: []lookupTestCase{
			{cmd: cmdMkdir, path: "test"},
			{cmd: cmdMkdir, path: "test/dir"},
			{cmd: cmdCreate, path: "test/a"},
			{cmd: cmdCreate, path: "test/b"},
			{cmd: cmdCreate, path: "test/c"},
			{cmd: cmdLookup, path: "test", want: &proto.LookupResponse{IsDir: true, Size: 4, Files: []string{"a", "b", "c", "dir"}}},
		},
		desc: "create directory containing a subdirectory and 3 files, then lookup the directory",
	},
}
