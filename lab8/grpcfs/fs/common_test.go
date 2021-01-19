package fs

import (
	"dat320/lab8/grpcfs/proto"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
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
	fileMode = os.ModePerm
)

const (
	cmdOpen = iota
	cmdLookup
	cmdMkdir
	cmdRemove
	cmdRead
	cmdWrite
	cmdSeek
	cmdClose
)

var grpcServerAddr = flag.String("addr", defaultServerAddr, "address:port to open gRPC servers during tests")

var openFlagToStr = map[int]string{
	OpenRead:      "OpenRead",
	OpenWrite:     "OpenWrite",
	OpenReadWrite: "OpenReadWrite",
}

var (
	errCreateInvalidPath  = errors.New("tried to create file/directory at invalid path")
	errOpenInvalidFlag    = errors.New("files must be opened with one of the OpenRead, OpenWrite or OpenReadWrite flags")
	errOpenDir            = errors.New("cannot open directories")
	errSeekNegativeOffset = errors.New("cannot Seek to a negative offset")
)

var utf8str = `ăѣ𝔠ծềſģȟᎥ𝒋ǩľḿꞑȯ𝘱𝑞𝗋𝘴ȶ𝞄𝜈ψ𝒙𝘆𝚣1234567890!@#$%^&*()-_=+[{]};:'",<.>/?~𝘈Ḇ𝖢𝕯٤ḞԍНǏ𝙅ƘԸⲘ𝙉০Ρ𝗤Ɍ𝓢ȚЦ𝒱Ѡ𝓧ƳȤѧᖯć𝗱ễ𝑓𝙜Ⴙ𝞲𝑗𝒌ļṃŉо𝞎𝒒ᵲꜱ𝙩ừ𝗏ŵ𝒙𝒚ź1234567890!@#$%^&*()-_=+[{]};:'",<.>/?~АḂⲤ𝗗𝖤𝗙ꞠꓧȊ𝐉𝜥ꓡ𝑀𝑵Ǭ𝙿𝑄Ŗ𝑆𝒯𝖴𝘝𝘞ꓫŸ𝜡ả𝘢ƀ𝖼ḋếᵮℊ𝙝Ꭵ𝕛кιṃդⱺ𝓅𝘲𝕣𝖘ŧ𝑢ṽẉ𝘅ყž1234567890!@#$%^&*()-_=+[{]};:'",<.>/?~Ѧ𝙱ƇᗞΣℱԍҤ١𝔍К𝓛𝓜ƝȎ𝚸𝑄Ṛ𝓢ṮṺƲᏔꓫ𝚈𝚭𝜶Ꮟçძ𝑒𝖿𝗀ḧ𝗂𝐣ҝɭḿ𝕟𝐨𝝔𝕢ṛ𝓼тú𝔳ẃ⤬𝝲𝗓1234567890!@#$%^&*()-_=+[{]};:'",<.>/?~𝖠Β𝒞𝘋𝙴𝓕ĢȞỈ𝕵ꓗʟ𝙼ℕ০𝚸𝗤ՀꓢṰǓⅤ𝔚Ⲭ𝑌𝙕𝘢𝕤`

var fileToWrite = func() []byte {
	b, err := ioutil.ReadFile("file_test.go")
	if err != nil {
		return []byte("file could not be opened...")
	}
	return b
}()

type fnParams struct {
	// function type
	fn int
	// prepended to function/method name
	prefix string
	// pathname/filename
	name string
	// input to Open
	flag int
	// input to Read/Write
	p []byte
	// inputs to Seek
	offset int64
	whence int
}

func (p fnParams) String() string {
	switch p.fn {
	case cmdOpen:
		return fmt.Sprintf("%sOpen(\"%s\", %s)", p.prefix, p.name, openFlagToStr[p.flag])
	case cmdMkdir:
		return fmt.Sprintf("%sMkdir(\"%s\", %d)", p.prefix, p.name, fileMode)
	case cmdRemove:
		return fmt.Sprintf("%sRemove(\"%s\")", p.prefix, p.name)
	case cmdLookup:
		return fmt.Sprintf("%sLookup(\"%s\")", p.prefix, p.name)
	case cmdRead:
		if len(p.p) > 20 {
			return fmt.Sprintf(fmt.Sprintf("%sRead(<%d bytes (truncated)>)", p.prefix, len(p.p)))
		}
		return fmt.Sprintf("%sRead(%v)", p.prefix, p.p)
	case cmdWrite:
		if len(p.p) > 20 {
			return fmt.Sprintf(fmt.Sprintf("%sWrite(<%d bytes (truncated)>)", p.prefix, len(p.p)))
		}
		return fmt.Sprintf("%sWrite(%v)", p.prefix, p.p)
	case cmdSeek:
		return fmt.Sprintf("%sSeek(%d, %d)", p.prefix, p.offset, p.whence)
	case cmdClose:
		return fmt.Sprintf("%sClose()", p.prefix)
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

func cmpObjStates(a, b interface{}) bool {
	return (a == nil && b == nil) || (a != nil && b != nil)
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
		errs:  []error{os.ErrNotExist},
		desc:  "try to create subdirectory within a non-existent directory",
	},
	{
		name:  "duplicate_dir",
		paths: []string{"test", "test"},
		errs:  []error{nil, os.ErrExist}, desc: "try to create the same directory twice",
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
	{
		name:  "nested_dirs",
		paths: []string{"a", "a/b", "a/b/c", "a/d", "a/b/e"},
		errs:  []error{nil, nil, nil, nil, nil},
		desc:  "create 5 directories, including nested directories",
	},
}

var openTests = []struct {
	name string
	flag int
	f    *File
	err  error
	// whether a file should be created before the test calls Open
	createBeforeOpen bool
	// confirm that a file with name has not been created
	checkFileNotCreated bool
	// create the file as a directory
	isDir bool
	desc  string
}{
	{
		name:                "read_not_exist",
		flag:                OpenRead,
		err:                 os.ErrNotExist,
		checkFileNotCreated: true,
		desc:                "try to open non-existent file for reading -> fail",
	},
	{
		name: "open_invalid_flag_not_exist",
		flag: int(math.MaxInt64),
		err:  ErrOpenInvalidFlag,
		desc: "pass an invalid flag to Open -> fail",
	},
	{
		name:             "open_invalid_flag_exist",
		flag:             int(math.MaxInt64),
		err:              ErrOpenInvalidFlag,
		createBeforeOpen: true,
		desc:             "create a file, then pass an invalid flag to Open -> fail",
	},
	{
		name:             "read_exist",
		flag:             OpenRead,
		f:                new(File),
		createBeforeOpen: true,
		desc:             "create a file, then open it for reading",
	},
	{
		name: "write_not_exist",
		flag: OpenWrite,
		desc: "open a non-existent file for writing, creating it in the process",
	},
	{
		name:             "write_exist",
		flag:             OpenWrite,
		createBeforeOpen: true,
		desc:             "create a file, then open it for writing",
	},
	{
		name: "readwrite_not_exist",
		flag: OpenReadWrite,
		desc: "open a non-existent file for reading/writing, creating it in the process",
	},
	{
		name:             "readwrite_exist",
		flag:             OpenReadWrite,
		createBeforeOpen: true,
		desc:             "create a file, then open it for reading/writing",
	},
}

type lookupOp struct {
	cmd int
	// path to open/mkdir/close/remove
	name string
	// input to Write
	p []byte
	// want
	// number of bytes (Write)
	n   int
	err error
	// fs.Lookup results
	isDir bool
	size  int
	files []string
}

var lookupTests = []struct {
	name  string
	steps []lookupOp
	desc  string
}{
	{
		name: "invalid_path",
		steps: []lookupOp{
			{cmd: cmdLookup, name: "test", err: os.ErrNotExist},
		},
		desc: "try to look up a nonexistent file/directory",
	},
	{
		name: "empty_dir",
		steps: []lookupOp{
			{cmd: cmdMkdir, name: "test"},
			{cmd: cmdLookup, name: "test", isDir: true},
		},
		desc: "make and lookup an empty directory",
	},
	{
		name: "empty_file",
		steps: []lookupOp{
			{cmd: cmdOpen, name: "test"},
			{cmd: cmdLookup, name: "test"},
		},
		desc: "create (by opening as read-write) an empty file, then look it up",
	},
	{
		name: "dir_with_file",
		steps: []lookupOp{
			{cmd: cmdMkdir, name: "test"},
			{cmd: cmdOpen, name: "test/file"},
			{cmd: cmdLookup, name: "test", isDir: true, size: 1, files: []string{"file"}},
		},
		desc: "make a directory containing a file, then look it up",
	},
	{
		name: "dir_with_dir_and_file",
		steps: []lookupOp{
			{cmd: cmdMkdir, name: "test"},
			{cmd: cmdMkdir, name: "test/dir"},
			{cmd: cmdOpen, name: "test/file"},
			{cmd: cmdLookup, name: "test", isDir: true, size: 2, files: []string{"dir", "file"}},
		},
		desc: "make a directory containing a subdirectory and a file, then look it up",
	},
	{
		name: "dir_with_many_subs",
		steps: []lookupOp{
			{cmd: cmdMkdir, name: "test"},
			{cmd: cmdMkdir, name: "test/dir_a"},
			{cmd: cmdMkdir, name: "test/dir_b"},
			{cmd: cmdMkdir, name: "test/dir_c"},
			{cmd: cmdOpen, name: "test/a_file"},
			{cmd: cmdOpen, name: "test/file_b"},
			{cmd: cmdLookup, name: "test", isDir: true, size: 5, files: []string{"a_file", "dir_a", "dir_b", "dir_c", "file_b"}},
		},
		desc: "make a directory (test) containing 3 subdirectories (test/dir_a, test/dir_b, test/dir_c) and 2 files (test/a_file, test/file_b), then look it up",
	},
}

var closeTests = []struct {
	name    string
	flag    int
	openErr error
	// whether a file should be created before opening it for the test case
	createBeforeOpen bool
	want             error
	desc             string
}{
	{
		name:    "readonly_not_exist",
		flag:    OpenRead,
		openErr: os.ErrNotExist,
		want:    os.ErrNotExist,
		desc:    "try to open, then close a read-only file which does not yet exist",
	},
	{
		name:             "readonly_existing",
		flag:             OpenRead,
		createBeforeOpen: true,
		desc:             "create a file before the test, then open it read-only, then close it",
	},
	{
		name: "writeonly",
		flag: OpenWrite,
		desc: "open a write-only file, then close it",
	},
	{
		name:             "writeonly_existing",
		flag:             OpenWrite,
		createBeforeOpen: true,
		desc:             "create a file before the test, then open it write-only, then close it",
	},
	{
		name: "readwrite",
		flag: OpenReadWrite,
		desc: "open a read-write file, then close it",
	},
	{
		name:             "readwrite_existing",
		flag:             OpenReadWrite,
		createBeforeOpen: true,
		desc:             "create a file before the test, then open it read-write, then close it",
	},
}

var writeTests = []struct {
	name  string
	desc  string
	in    []byte
	isDir bool
	// want
	n   int
	err error
}{
	{
		name: "noop",
		in:   []byte{},
		desc: "write an empty slice -> success",
	},
	{
		name: "abc",
		in:   []byte("abc"),
		n:    3,
		desc: "write \"abc\"",
	},
	{
		name: "testing",
		in:   []byte("testing"),
		n:    7,
		desc: "write \"testing\"",
	},
	{
		name:  "write_to_dir",
		in:    []byte("abc"),
		err:   ErrOpenDir,
		isDir: true,
		desc:  "trying to write to a directory -> fail",
	},
	{
		name: "textfile",
		in:   fileToWrite,
		n:    len(fileToWrite),
		desc: "write the contents of a textfile",
	},
	{
		name: "utf8",
		in:   []byte(utf8str),
		n:    len([]byte(utf8str)),
		desc: "writing a UTF-8 string",
	},
}

var readWriteTests = []struct {
	name    string
	desc    string
	toWrite [][]byte
	// input to Read
	in []byte
	// want
	n   int
	err error
	// bytes we want to read
	wantRead []byte
}{
	{
		name:     "single_byte",
		toWrite:  [][]byte{{1}},
		in:       make([]byte, 1),
		n:        1,
		wantRead: []byte{1},
		desc:     "write 1 byte, then read it",
	},
	{
		name:    "abc",
		toWrite: [][]byte{[]byte("abc")},
		in:      make([]byte, 3), n: 3,
		wantRead: []byte("abc"),
		desc:     "write \"abc\", then read it",
	},
	{
		name:    "testing",
		toWrite: [][]byte{[]byte("testing")},
		in:      make([]byte, 7), n: 7,
		wantRead: []byte("testing"),
		desc:     "write \"testing\", then read it",
	},
	{
		name:     "textfile",
		toWrite:  [][]byte{fileToWrite},
		in:       make([]byte, len(fileToWrite)),
		n:        len(fileToWrite),
		wantRead: fileToWrite,
		desc:     "write the contents of a textfile, then read it",
	},
	{
		name:     "utf8",
		toWrite:  [][]byte{[]byte(utf8str)},
		in:       make([]byte, len([]byte(utf8str))),
		n:        len([]byte(utf8str)),
		wantRead: []byte(utf8str),
		desc:     "write an UTF-8 string, then read it",
	},
	{
		name:     "multiple",
		toWrite:  [][]byte{[]byte("a"), []byte("b"), []byte("c")},
		in:       make([]byte, 3),
		n:        3,
		wantRead: []byte("abc"),
		desc:     "write \"a\", then \"b\", then \"c\", then read \"abc\"",
	},
	{
		name:     "multiple_v2",
		toWrite:  [][]byte{[]byte("a"), []byte("bc"), []byte("def")},
		in:       make([]byte, 6),
		n:        6,
		wantRead: []byte("abcdef"),
		desc:     "write \"a\", then \"bc\", then \"def\", then read \"abcdef\"",
	},
}

type fileOp struct {
	// file "ID" (so we can have more than 1 File object in the same test)
	id  int
	cmd int
	// path to open/close/remove
	name string
	// flag to give to Open
	flag int
	// input to Read/Write
	p []byte
	// inputs to Seek
	offset int64
	whence int
	// want
	n        int
	err      error
	wantRead []byte
	// result of Seek
	newOffset int64
}

var readWriteSeekTests = []struct {
	name  string
	steps []fileOp
	desc  string
}{
	{
		name: "seek_to_start_absolute",
		steps: []fileOp{
			{cmd: cmdOpen, name: "test", flag: OpenReadWrite},
			{cmd: cmdWrite, p: []byte("testing"), n: 7},
			{cmd: cmdSeek, offset: 0, whence: 0, newOffset: 0},
			{cmd: cmdRead, p: make([]byte, len("testing")), n: 7, wantRead: []byte("testing")},
		},
		desc: "write 'testing', Seek to the start of the file, read 'testing'",
	},
	{
		name: "seek_to_start_relative",
		steps: []fileOp{
			{cmd: cmdOpen, name: "test", flag: OpenReadWrite},
			{cmd: cmdWrite, p: []byte("testing"), n: 7},
			{cmd: cmdSeek, offset: -7, whence: 1, newOffset: 0},
			{cmd: cmdRead, p: make([]byte, len("testing")), n: 7, wantRead: []byte("testing")},
		},
		desc: "write 'testing', Seek to the start of the file relative to current offset, read 'testing'",
	},
	{
		name: "seek_to_start_from_end",
		steps: []fileOp{
			{cmd: cmdOpen, name: "test", flag: OpenReadWrite},
			{cmd: cmdWrite, p: []byte("testing"), n: 7},
			{cmd: cmdSeek, offset: -6, whence: 2, newOffset: 0},
			{cmd: cmdRead, p: make([]byte, len("testing")), n: 7, wantRead: []byte("testing")},
		},
		desc: "write 'testing', Seek to the start of the file from the end of the file, read 'testing'",
	},
	{
		name: "seek_middle_absolute",
		steps: []fileOp{
			{cmd: cmdOpen, name: "test", flag: OpenReadWrite},
			{cmd: cmdWrite, p: []byte("testing"), n: 7},
			{cmd: cmdSeek, offset: 4, whence: 0, newOffset: 4},
			{cmd: cmdRead, p: make([]byte, len("ing")), n: 3, wantRead: []byte("ing")},
		},
		desc: "write 'testing', Seek to offset 4 (absolute), read 'ing'",
	},
	{
		name: "seek_middle_relative",
		steps: []fileOp{
			{cmd: cmdOpen, name: "test", flag: OpenReadWrite},
			{cmd: cmdWrite, p: []byte("testing"), n: 7},
			{cmd: cmdSeek, offset: -3, whence: 1, newOffset: 4},
			{cmd: cmdRead, p: make([]byte, len("ing")), n: 3, wantRead: []byte("ing")},
		},
		desc: "write 'testing', Seek to offset 4 (by seeking -3 relative to cur), read 'ing'",
	},
	{
		name: "seek_middle_from_end",
		steps: []fileOp{
			{cmd: cmdOpen, name: "test", flag: OpenReadWrite},
			{cmd: cmdWrite, p: []byte("testing"), n: 7},
			{cmd: cmdSeek, offset: -2, whence: 2, newOffset: 4},
			{cmd: cmdRead, p: make([]byte, len("ing")), n: 3, wantRead: []byte("ing")},
		},
		desc: "write 'testing', Seek to offset 4 (by seeking -2 relative to the end of the file), read 'ing'",
	},
	{
		name: "seek_eof",
		steps: []fileOp{
			{cmd: cmdOpen, name: "test", flag: OpenReadWrite},
			{cmd: cmdWrite, p: []byte("testing"), n: 7},
			{cmd: cmdSeek, offset: 8, whence: 0, newOffset: 8},
			{cmd: cmdRead, p: make([]byte, 7), n: 0, wantRead: make([]byte, 7), err: io.EOF},
			{cmd: cmdWrite, p: []byte{1}, n: 0, err: io.EOF},
		},
		desc: "write 'testing', Seek to offset 8 (after end of file), try to read/write -> fail both",
	},
	{
		name: "seek_negative_offset",
		steps: []fileOp{
			{cmd: cmdOpen, name: "test", flag: OpenReadWrite},
			{cmd: cmdSeek, offset: -1, whence: 0, newOffset: 0, err: ErrSeekNegativeOffset},
		},
		desc: "seek to a negative offset (-1) -> fail",
	},
	{
		name: "seek_check_offset",
		steps: []fileOp{
			{cmd: cmdOpen, name: "test", flag: OpenReadWrite},
			{cmd: cmdWrite, p: []byte("testing"), n: 7},
			{cmd: cmdSeek, offset: 0, whence: 1, newOffset: 7},
		},
		desc: "write 'testing', Seek to a relative offset of 0 -> new offset is current offset",
	},
	{
		name: "seek_overwrite",
		steps: []fileOp{
			{cmd: cmdOpen, name: "test", flag: OpenReadWrite},
			{cmd: cmdWrite, p: []byte("testing"), n: 7},
			{cmd: cmdSeek, offset: 0, whence: 0, newOffset: 0},
			{cmd: cmdWrite, p: []byte("TEST"), n: 4},
			{cmd: cmdSeek, offset: 0, whence: 0, newOffset: 0},
			{cmd: cmdRead, p: make([]byte, 7), n: 7, wantRead: []byte("TESTing")},
		},
		desc: "write 'testing', Seek to start, write 'TEST' (overwriting 'test'), Seek back to start, read 'TESTing'",
	},
	{
		name: "seek_fix_typos",
		steps: []fileOp{
			{cmd: cmdOpen, name: "test", flag: OpenReadWrite},
			{cmd: cmdWrite, p: []byte("Uinverstiy of Stavnager"), n: len("Uinverstiy of Stavnager")},
			{cmd: cmdSeek, offset: 1, whence: 0, newOffset: 1},
			{cmd: cmdWrite, p: []byte("ni"), n: 2},
			{cmd: cmdSeek, offset: 4, whence: 1, newOffset: 7},
			{cmd: cmdWrite, p: []byte("it"), n: 2},
			{cmd: cmdSeek, offset: -4, whence: 2, newOffset: 18},
			{cmd: cmdWrite, p: []byte("an"), n: 2},
			{cmd: cmdSeek, offset: 0, whence: 0, newOffset: 0},
			{cmd: cmdRead, p: make([]byte, len("University of Stavanger")), n: len("University of Stavanger"), wantRead: []byte("University of Stavanger")},
		},
		desc: "write 'Uinverstiy of Stavnager', fix typos with by seeking to start/current/end positions and writing the fixes, Seek to start, read 'University of Stavanger'",
	},
}
