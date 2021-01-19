package fs

import (
	"context"
	"dat320/lab8/grpcfs/fsserver"
	"dat320/lab8/grpcfs/proto"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
)

func TestClose(t *testing.T) {
	t.Parallel()
	for _, test := range closeTests {
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

			fileName := "test"
			if test.createBeforeOpen {
				_, err = mock.Create(context.Background(), &proto.CreateRequest{Path: fileName})
				if err != nil {
					t.Errorf("failed to create file before Open: %v.\nPlease notify the lab staff if you see this.\n(%s)", err, test.desc)
				}
			}

			f, err := fs.Open(fileName, test.flag)
			fnOpen := fnParams{fn: cmdOpen, name: fileName, flag: test.flag}
			if !cmp.Equal(err, test.openErr, cmpOptErrorComparer) {
				// in some cases, the file should not be opened successfully, e.g. for read-only files
				t.Errorf("%s = '%v', want '%v'\n(%s)", fnOpen, err, test.openErr, test.desc)
			} else if test.openErr != nil {
				// if test.openErr is not nil and matches err,
				// then the test case should end here
				return
			}

			if err != nil {
				t.Errorf("%s: failed to open: %v\n(%s)", fnOpen, err, test.desc)
				return
			}

			err = f.Close()
			fnClose := fnParams{fn: cmdClose}
			if !cmp.Equal(err, test.want, cmpOptErrorComparer) {
				t.Errorf("%s = '%v', want '%v'\n(%s)", fnClose, err, test.want, test.desc)
			}

			// check if streams are still open after closing the File
			r, w := mock.getOpenStatus()
			if r || w {
				t.Errorf("%s: Reader or Writer stream still open after closing File\n(%s)", fnClose, test.desc)
			}

			// try to close the file again
			if err = f.Close(); err == nil {
				t.Errorf("%s: successfully closed the same File twice\n(%s)", fnClose, test.desc)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	t.Parallel()
	for _, test := range writeTests {
		fsServer := fsserver.NewFileSystemServer()
		s, addr, err := setupServer(t, fsServer)
		if err != nil {
			t.Errorf("failed setting up server: %v\n(%s)", err, test.desc)
			return
		}

		t.Run(test.
			name, func(t *testing.T) {
			defer s.Stop()

			fs, err := NewFileSystem(addr, grpc.WithInsecure())
			if err != nil {
				t.Errorf("failed to setup filesystem: %v\n(%s)", err, test.desc)
				return
			}

			fname := "test"
			if test.isDir {
				// make a directory which should fail to open for writing
				err = fs.Mkdir(fname, fileMode)
				if err != nil {
					t.Errorf("%s: failed to make directory: %v\n(%s)", fnParams{fn: cmdMkdir, name: fname}, err, test.desc)
				}
			}

			// open the file we want to write to
			f, err := fs.Open(fname, OpenWrite)
			fnOpen := fnParams{fn: cmdOpen, name: fname, flag: OpenWrite}
			if err != nil {
				if !test.isDir {
					// we expect to be able to open the file
					t.Errorf("%s: failed to open: %v\n(%s)", fnOpen, err, test.desc)
				}
				return
			} else if test.isDir {
				// managed to open directory for writing
				t.Errorf("%s: opened a directory for writing\n(%s)", fnOpen, test.desc)
			}
			defer f.Close()

			// write the contents and check the results
			n, err := f.Write(test.in)
			fnWrite := fnParams{fn: cmdWrite, p: test.in}
			if diff := cmp.Diff(test.n, n); diff != "" {
				t.Errorf("%s: unexpected number of bytes written; (-want +got):\n%s\n(%s)", fnWrite, diff, test.desc)
			}

			if !cmp.Equal(err, test.err, cmpOptErrorComparer) {
				t.Errorf("%s = %d, '%v', want %d, '%v'\n(%s)", fnWrite, n, err, test.n, test.err, test.desc)
			}
		})
	}
}

// Write to remote FS with one file object, then Read the content from the remote FS with another file object
func TestReadWrite(t *testing.T) {
	t.Parallel()
	for _, test := range readWriteTests {
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

			fname := "test"
			if test.toWrite != nil {
				// open the file we want to write to
				fwrite, err := fs.Open(fname, OpenWrite)
				if err != nil {
					// we expect to be able to open the file
					t.Errorf("%s: failed to open: %v\n(%s)", fnParams{fn: cmdOpen, name: fname, flag: OpenWrite}, err, test.desc)
					return
				}
				defer fwrite.Close()

				// write the content that we want to read
				for _, b := range test.toWrite {
					_, err = fwrite.Write(b)
					if err != nil {
						t.Errorf("%s: unexpected error: %v\n(%s)", fnParams{fn: cmdWrite, p: b}, err, test.desc)
					}
				}
			}

			// open the file we want to read from
			fread, err := fs.Open(fname, OpenRead)
			fnOpen := fnParams{fn: cmdOpen, name: fname, flag: OpenRead}
			if test.toWrite == nil && err == nil {
				t.Errorf("%s: successfully opened file for reading, despite not having written to it (file should not exist)\n(%s)", fnOpen, test.desc)
			} else if test.toWrite == nil && err != nil {
				// nothing to write + file failed to open for
				// reading is the correct behavior
				return
			} else if err != nil {
				// we expect to be able to open the file
				t.Errorf("%s: failed to open: %v\n(%s)", fnOpen, err, test.desc)
				return
			}
			defer fread.Close()

			fnRead := fnParams{fn: cmdRead, p: test.in}
			// read the content we just wrote
			n, err := fread.Read(test.in)
			if !cmp.Equal(err, test.err, cmpOptErrorComparer) {
				t.Errorf("%s = %d, '%v', want %d, '%v'\n(%s)", fnRead, n, err, test.n, test.err, test.desc)
			}

			// check n is correct
			if diff := cmp.Diff(test.n, n); diff != "" {
				t.Errorf("%s: unexpected number of bytes read; (-want +got):\n%s\n(%s)", fnRead, diff, test.desc)
			}

			// check that we read the same bytes that we wrote
			if diff := cmp.Diff(test.wantRead, test.in); diff != "" {
				t.Errorf("%s: unexpected content read; (-want +got):\n%s\n(%s)", fnRead, diff, test.desc)
			}
		})
	}
}

func TestReadWriteSeek(t *testing.T) {
	t.Parallel()
	for _, test := range readWriteSeekTests {
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

			// file ID -> open file
			files := make(map[int]*File)

			for _, step := range test.steps {
				switch step.cmd {
				// open a file and store it in files if no error occurs
				case cmdOpen:
					f, err := fs.Open(step.name, step.flag)
					if !cmp.Equal(err, step.err, cmpOptErrorComparer) {
						t.Errorf("%s = ..., '%v', want ..., '%v'\n(%s)", fnParams{fn: cmdOpen, name: step.name, flag: step.flag}, err, step.err, test.desc)
					}
					if err == nil {
						files[step.id] = f
					}

				case cmdRead:
					if f, fileExists := files[step.id]; fileExists {
						// copy p so we can display the original values in test errors below
						pCopy := make([]byte, len(step.p))
						copy(pCopy, step.p)

						n, err := f.Read(step.p)
						fnCall := fnParams{fn: cmdRead, p: pCopy}

						if !cmp.Equal(err, step.err, cmpOptErrorComparer) {
							t.Errorf("%s = %d, '%v', want %d, '%v'\n(%s)", fnCall, n, err, step.n, step.err, test.desc)
						}

						if diff := cmp.Diff(step.n, n); diff != "" {
							t.Errorf("%s: unexpected number of bytes read; (-want +got):\n%s\n(%s)", fnCall, diff, test.desc)
						}

						if diff := cmp.Diff(step.wantRead, step.p); diff != "" {
							t.Errorf("%s: unexpected content read; (-want +got):\n%s\n(%s)", fnCall, diff, test.desc)
						}
					} else {
						t.Errorf("Could not perform Read since the file does not exist.\n(%s)", test.desc)
					}

				case cmdWrite:
					if f, fileExists := files[step.id]; fileExists {
						n, err := f.Write(step.p)
						fnCall := fnParams{fn: cmdWrite, p: step.p}

						if !cmp.Equal(err, step.err, cmpOptErrorComparer) {
							t.Errorf("%s = %d, '%v', want %d, '%v'\n(%s)", fnCall, n, err, step.n, step.err, test.desc)
						}

						if diff := cmp.Diff(step.n, n); diff != "" {
							t.Errorf("%s: unexpected number of bytes written; (-want +got):\n%s\n(%s)", fnCall, diff, test.desc)
						}
					} else {
						t.Errorf("Could not perform Write since the file does not exist.\n(%s)", test.desc)
					}

				case cmdSeek:
					if f, fileExists := files[step.id]; fileExists {
						newOffset, err := f.Seek(step.offset, step.whence)
						fnCall := fnParams{fn: cmdSeek, offset: step.offset, whence: step.whence}

						if !cmp.Equal(err, step.err, cmpOptErrorComparer) {
							t.Errorf("%s = %d, '%v', want %d, '%v'\n(%s)", fnCall, newOffset, err, step.newOffset, step.err, test.desc)
						}

						if diff := cmp.Diff(step.newOffset, newOffset); diff != "" {
							t.Errorf("%s: unexpected new offset; (-want +got):\n%s\n(%s)", fnCall, diff, test.desc)
						}
					} else {
						t.Errorf("Could not perform Seek since the file does not exist.\n(%s)", test.desc)
					}

				case cmdClose:
					if f, fileExists := files[step.id]; fileExists {
						err := f.Close()
						fnCall := fnParams{fn: cmdClose}

						if !cmp.Equal(err, step.err, cmpOptErrorComparer) {
							t.Errorf("%s = '%v', want '%v'\n(%s)", fnCall, err, step.err, test.desc)
						}
					} else {
						t.Errorf("Could not perform Close since the file does not exist.\n(%s)", test.desc)
					}
				}
			}
		})
	}
}
