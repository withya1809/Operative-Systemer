package fs

import (
	"context"
	"dat320/lab8/grpcfs/fsserver"
	"dat320/lab8/grpcfs/proto"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
)

func TestMkdir(t *testing.T) {
	t.Parallel()
	for _, test := range mkdirTests {
		fsServer := fsserver.NewFileSystemServer()
		s, addr, err := setupServer(t, fsServer)
		if err != nil {
			t.Errorf("failed setting up server: %v\n(%s)", err, test.desc)
			return
		}

		t.Run(test.name, func(t *testing.T) {
			defer s.Stop()

			fs, err := NewFileSystem(addr, grpc.WithInsecure())
			if err != nil {
				t.Errorf("failed to setup filesystem: %v\n(%s)", err, test.desc)
				return
			}

			for i, ithErr := range test.errs {
				err = fs.Mkdir(test.paths[i], fileMode)

				if !cmp.Equal(err, ithErr, cmpOptErrorComparer) {
					t.Errorf("%s = '%v', want '%v'\n(%s)", fnParams{fn: cmdMkdir, name: test.paths[i]}, err, ithErr, test.desc)
				}
			}
		})
	}
}

func TestOpen(t *testing.T) {
	t.Parallel()
	for _, test := range openTests {
		mock := &mockFSServer{filesCreated: make(map[string]bool)}
		s, addr, err := setupServer(t, mock)
		if err != nil {
			t.Errorf("failed setting up server: %v\n(%s)", err, test.desc)
			return
		}

		t.Run(test.name, func(t *testing.T) {
			defer s.Stop()

			fs, err := NewFileSystem(addr, grpc.WithInsecure())
			if err != nil {
				t.Errorf("failed to setup filesystem: %v\n(%s)", err, test.desc)
				return
			}

			path := "test"

			if test.createBeforeOpen {
				// create file or directory
				if test.isDir {
					_, err = mock.Mkdir(context.Background(), &proto.MakeDirRequest{Path: path, FileMode: uint32(fileMode)})
					if err != nil {
						t.Errorf("failed to make directory before calling Open: %v\n(%s)", err, test.desc)
					}
				} else {
					_, err = mock.Create(context.Background(), &proto.CreateRequest{Path: path})
					if err != nil {
						t.Errorf("failed to create file before Open: %v.\nPlease notify the lab staff if you see this.\n(%s)", err, test.desc)
					}
				}
			}

			f, err := fs.Open(path, test.flag)
			fnCall := fnParams{fn: cmdOpen, name: path, flag: test.flag}
			if !cmp.Equal(err, test.err, cmpOptErrorComparer) {
				t.Errorf("%s = ..., '%v', want ..., '%v'\n(%s)", fnCall, err, test.err, test.desc)
			}

			// since internals of the File struct depend on the
			// implementation, we only check whether it's nil or not
			if !cmpObjStates(f, test.f) {
				t.Errorf("%s: got non-nil File when expecting nil File (or vice versa): got %v\n(%s)", fnCall, f, test.desc)
			}

			if test.checkFileNotCreated {
				// Files should only be created for successful requests involving Writers (or Mkdir).
				// Check if the file exists (Open should not create files opened only for reading).
				_, err = mock.Lookup(context.Background(), &proto.LookupRequest{Path: path})
				if err == nil {
					t.Errorf("%s: the file %s was created despite expected failure/read request to non-existent file\n(%s)", fnCall, path, test.desc)
				}
			}
		})
	}
}

func TestLookup(t *testing.T) {
	t.Parallel()
	for _, test := range lookupTests {
		fsServer := fsserver.NewFileSystemServer()
		s, addr, err := setupServer(t, fsServer)
		if err != nil {
			t.Errorf("failed setting up server: %v\n(%s)", err, test.desc)
			return
		}

		t.Run(test.name, func(t *testing.T) {
			defer s.Stop()

			fs, err := NewFileSystem(addr, grpc.WithInsecure())
			if err != nil {
				t.Errorf("failed to setup filesystem: %v\n(%s)", err, test.desc)
				return
			}

			var f *File
			for _, step := range test.steps {
				switch step.cmd {
				case cmdOpen:
					innerF, err := fs.Open(step.name, OpenReadWrite)
					if !cmp.Equal(err, step.err, cmpOptErrorComparer) {
						t.Errorf("%s = ..., '%v', want '%v'\n(%s)", fnParams{fn: cmdOpen, name: step.name, flag: OpenReadWrite}, err, step.err, test.desc)
					}
					if err == nil {
						f = innerF
						defer innerF.Close()
					}

				case cmdMkdir:
					err := fs.Mkdir(step.name, fileMode)
					if !cmp.Equal(err, step.err, cmpOptErrorComparer) {
						t.Errorf("%s = '%v', want '%v'\n(%s)", fnParams{fn: cmdMkdir, name: step.name}, err, step.err, test.desc)
					}

				case cmdWrite:
					fnCall := fnParams{fn: cmdWrite, p: step.p}

					if f == nil {
						t.Errorf("%s: file is nil, cannot write\n(%s)", fnCall, test.desc)
						break
					}

					n, err := f.Write(step.p)
					if !cmp.Equal(err, step.err, cmpOptErrorComparer) {
						t.Errorf("%s = %d, %v, want %d, '%v'\n(%s)", fnCall, n, err, step.n, step.err, test.desc)
					}

					if !cmp.Equal(step.n, n) {
						t.Errorf("%s = %d, %v, want %d, '%v'\n(%s)", fnCall, n, err, step.n, step.err, test.desc)
					}

				case cmdLookup:
					fnCall := fnParams{fn: cmdLookup, name: step.name}
					isDir, size, files, err := fs.Lookup(step.name)
					errMsg := fmt.Sprintf("%s = %v, %d, %v, '%v', want %v, %d, %v, '%v'\n(%s)",
						fnCall,
						isDir, size, files, err,
						step.isDir, step.size, step.files, step.err,
						test.desc)

					if !cmp.Equal(err, step.err, cmpOptErrorComparer) {
						t.Error(errMsg)
					}

					if step.isDir != isDir {
						t.Error(errMsg)
					}

					if step.size != size {
						t.Error(errMsg)
					}

					if diff := cmp.Diff(step.files, files); diff != "" {
						t.Errorf("%s: wrong files; (-want +got):\n%s\n(%s)", fnCall, diff, test.desc)
					}
				}
			}
		})
	}
}
