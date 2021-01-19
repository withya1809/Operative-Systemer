package fs

import (
	"context"
	"dat320/lab8/grpcfs/proto"
	"os"
	"sync"
)

type mockFSServer struct {
	// true if read/write streams are currently open
	readOpen, writeOpen bool
	// key: name, value: true for directories
	filesCreated map[string]bool
	lock         sync.Mutex
	proto.UnimplementedFileSystemServer
}

// Create creates an entry in the filesCreated map for the requested path
func (s *mockFSServer) Create(ctx context.Context, r *proto.CreateRequest) (*proto.CreateResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.filesCreated[r.Path] = false

	return new(proto.CreateResponse), nil
}

// Lookup checks for an entry with the requested path as the key in filesCreated.
// It does not provide info such as size or a list files within a directory.
func (s *mockFSServer) Lookup(ctx context.Context, r *proto.LookupRequest) (*proto.LookupResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if isDir, ok := s.filesCreated[r.Path]; ok {
		return &proto.LookupResponse{IsDir: isDir}, nil
	}

	return nil, os.ErrNotExist
}

// Remove removes the entry with the requested path as the key in filesCreated
func (s *mockFSServer) Remove(ctx context.Context, r *proto.RemoveRequest) (*proto.RemoveResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.filesCreated[r.Path]; ok {
		delete(s.filesCreated, r.Path)
		return new(proto.RemoveResponse), nil
	}
	return nil, os.ErrNotExist
}

// Reader continuously receives read request and sets readOpen to true as long
// as the stream is open.
func (s *mockFSServer) Reader(stream proto.FileSystem_ReaderServer) error {
	for {
		_, err := stream.Recv()
		if err != nil {
			s.lock.Lock()
			defer s.lock.Unlock()
			s.readOpen = false
			return err
		}

		s.lock.Lock()
		s.readOpen = true
		s.lock.Unlock()
	}
}

// Writer continuously receives write request and sets writeOpen to true as long
// as the stream is open.
func (s *mockFSServer) Writer(stream proto.FileSystem_WriterServer) error {
	for {
		_, err := stream.Recv()
		if err != nil {
			s.lock.Lock()
			defer s.lock.Unlock()
			s.writeOpen = false
			return err
		}

		s.lock.Lock()
		s.writeOpen = true
		s.lock.Unlock()
	}
}

// Mkdir makes a directory at the requested path.
// Should not overwrite existing files/directories.
func (s *mockFSServer) Mkdir(ctx context.Context, r *proto.MakeDirRequest) (*proto.MakeDirResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.filesCreated[r.Path] = true

	return new(proto.MakeDirResponse), nil
}

func (s *mockFSServer) getOpenStatus() (r, w bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.readOpen, s.writeOpen
}
