package fsserver

import (
	"context"
	"dat320/lab8/grpcfs/proto"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestCreate(t *testing.T) {
	t.Parallel()
	for _, test := range createTests {
		fsServer := NewFileSystemServer()
		s, addr, err := setupServer(t, fsServer)
		if err != nil {
			t.Errorf("failed setting up server: %v\n(%s)", err, test.desc)
			continue
		}

		t.Run(test.name, func(t *testing.T) {
			defer s.Stop()

			c, err := setupClient(t, addr)
			if err != nil {
				t.Errorf("failed to set up client: %v.\nPlease notify the lab staff if you see this.\n(%s)", err, test.desc)
				return
			}

			for i, ithErr := range test.errs {
				_, err = c.Create(context.Background(), &proto.CreateRequest{Path: test.paths[i]})
				fnCall := fnParams{fn: cmdCreate, path: test.paths[i]}

				if !cmp.Equal(ithErr, err, cmpOptErrorComparer) {
					t.Errorf("%s: unexpected error: got '%v', want '%v'\n(%s)", fnCall, err, ithErr, test.desc)
				}

				if ithErr == nil {
					// check that a Lookup request is successful and contains correct values
					fs := fsServer.fs
					if fs == nil {
						t.Errorf("Test %s: gRPC server's filesystem is nil\n(%s)", test.name, test.desc)
						continue
					}

					// use Stat to get the FileInfo of the file created during the test
					finfo, err := fs.Stat(test.paths[i])
					if err != nil {
						t.Errorf("Test %s: failed to Stat created file in the server's filesystem: %v\n(%s)", test.name, err, test.desc)
						continue
					}

					if finfo.IsDir() {
						t.Errorf("%s: the created file is a directory\n(%s)", fnCall, test.desc)
					}

					if finfo.Size() > 0 {
						t.Errorf("%s: the created file is not empty\n(%s)", fnCall, test.desc)
					}
				}
			}
		})
	}
}

func TestMkdir(t *testing.T) {
	t.Parallel()
	for _, test := range mkdirTests {
		fsServer := NewFileSystemServer()
		s, addr, err := setupServer(t, fsServer)
		if err != nil {
			t.Errorf("failed setting up server: %v\n(%s)", err, test.desc)
			continue
		}

		t.Run(test.name, func(t *testing.T) {
			defer s.Stop()

			c, err := setupClient(t, addr)
			if err != nil {
				t.Errorf("failed to set up client: %v.\nPlease notify the lab staff if you see this.\n(%s)", err, test.desc)
				return
			}

			for i, ithErr := range test.errs {
				_, err = c.Mkdir(context.Background(), &proto.MakeDirRequest{Path: test.paths[i], FileMode: fileMode})
				fnCall := fnParams{fn: cmdMkdir, path: test.paths[i]}

				if !cmp.Equal(ithErr, err, cmpOptErrorComparer) {
					t.Errorf("%s: unexpected error: got '%v', want '%v'\n(%s)", fnCall, err, ithErr, test.desc)
				}

				if ithErr == nil {
					// check that a Lookup request is successful and contains correct values
					fs := fsServer.fs
					if fs == nil {
						t.Errorf("Test %s: gRPC server's filesystem is nil\n(%s)", test.name, test.desc)
						continue
					}

					// use Stat to get the FileInfo of the file created during the test
					finfo, err := fs.Stat(test.paths[i])
					if err != nil {
						t.Errorf("Test %s: failed to Stat created directory in the server's filesystem: %v\n(%s)", test.name, err, test.desc)
						continue
					}

					if !finfo.IsDir() {
						t.Errorf("%s: the created directory is not a directory\n(%s)", fnCall, test.desc)
					}
				}
			}
		})
	}
}

func TestRemove(t *testing.T) {
	t.Parallel()
	for _, test := range rmTests {
		fsServer := NewFileSystemServer()
		s, addr, err := setupServer(t, fsServer)
		if err != nil {
			t.Errorf("failed setting up server: %v\n(%s)", err, test.desc)
			continue
		}

		t.Run(test.name, func(t *testing.T) {
			defer s.Stop()

			c, err := setupClient(t, addr)
			if err != nil {
				t.Errorf("failed to set up client: %v.\nPlease notify the lab staff if you see this.\n(%s)", err, test.desc)
				return
			}

			for _, step := range test.steps {
				fs := fsServer.fs
				if fs == nil {
					t.Errorf("Test %s: gRPC server's filesystem is nil\n(%s)", test.name, test.desc)
					continue
				}

				// create a file, make a directory or remove a file/directory
				switch step.cmd {
				case cmdCreate:
					f, err := fs.OpenFile(step.path, os.O_CREATE, os.FileMode(fileMode))
					if err != nil {
						t.Errorf("Test %s: failed to create file %s: %v\n(%s)", test.name, step.path, err, test.desc)
						continue
					}
					f.Close()
				case cmdMkdir:
					err := fs.Mkdir(step.path, os.FileMode(fileMode))
					if err != nil {
						t.Errorf("Test %s: failed to create directory %s: %v\n(%s)", test.name, step.path, err, test.desc)
					}
				case cmdRemove:
					_, err := c.Remove(context.Background(), &proto.RemoveRequest{Path: step.path})
					if !cmp.Equal(err, step.err, cmpOptErrorComparer) {
						t.Errorf("%s: unexpected error: got '%v', want '%v'\n(%s)", fnParams{fn: cmdRemove, path: step.path}, err, step.err, test.desc)
					}

					// check if file/directory still exists after being deleted
					_, err = fs.Stat(step.path)
					if err == nil {
						t.Errorf("Test %s: file/directory at %s should be removed, but Stat does not return an error\n(%s)", test.name, step.path, test.desc)
					}
				default:
					continue
				}
			}

		})
	}
}

func TestLookup(t *testing.T) {
	t.Parallel()
	for _, test := range lookupTests {
		fsServer := NewFileSystemServer()
		s, addr, err := setupServer(t, fsServer)
		if err != nil {
			t.Errorf("failed setting up server: %v\n(%s)", err, test.desc)
			continue
		}

		t.Run(test.name, func(t *testing.T) {
			defer s.Stop()

			c, err := setupClient(t, addr)
			if err != nil {
				t.Errorf("failed to set up client: %v.\nPlease notify the lab staff if you see this.\n(%s)", err, test.desc)
				return
			}

			for _, step := range test.steps {
				fs := fsServer.fs
				if fs == nil {
					t.Errorf("Test %s: gRPC server's filesystem is nil\n(%s)", test.name, test.desc)
					continue
				}

				switch step.cmd {
				case cmdCreate:
					f, err := fs.OpenFile(step.path, os.O_CREATE, os.FileMode(fileMode))
					if err != nil {
						t.Errorf("Test %s: failed to create file %s: %v\n(%s)", test.name, step.path, err, test.desc)
						continue
					}
					f.Close()

				case cmdWriteFile:
					// open a file, write some content to it, close it
					f, err := fs.OpenFile(step.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(fileMode))
					if err != nil {
						t.Errorf("Test %s: failed to create file %s: %v\n(%s)", test.name, step.path, err, test.desc)
						continue
					}

					if _, err := f.Write([]byte(step.toWrite)); err != nil {
						t.Errorf("Test %s: failed to write to file %s: %v\n(%s)", test.name, step.path, err, test.desc)
					}
					f.Close()

				case cmdMkdir:
					err := fs.Mkdir(step.path, os.FileMode(fileMode))
					if err != nil {
						t.Errorf("Test %s: failed to create directory %s: %v\n(%s)", test.name, step.path, err, test.desc)
					}

				case cmdRemove:
					err := fs.Remove(step.path)
					if err != nil {
						t.Errorf("Test %s: failed to remove file/directory %s: %v\n(%s)", test.name, step.path, err, test.desc)
					}

				case cmdLookup:
					r, err := c.Lookup(context.Background(), &proto.LookupRequest{Path: step.path})
					fnCall := fnParams{fn: cmdLookup, path: step.path}

					if !cmp.Equal(err, step.err, cmpOptErrorComparer) {
						t.Errorf("%s: unexpected error: got '%v', want '%v'\n(%s)", fnCall, err, step.err, test.desc)
					}

					// check that the response has the expected content
					if diff := cmp.Diff(step.want, r, cmpopts.IgnoreUnexported(proto.LookupResponse{})); diff != "" {
						t.Errorf("%s: invalid response; (-want +got):\n%s\n(%s)", fnCall, diff, test.desc)
					}

				default:
					continue
				}
			}
		})
	}
}
