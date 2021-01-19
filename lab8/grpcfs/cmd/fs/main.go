package main

import (
	"bufio"
	grpcfs "dat320/lab8/grpcfs/fs"
	"dat320/lab8/grpcfs/fsserver"
	pb "dat320/lab8/grpcfs/proto"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
)

var (
	addr      = flag.String("addr", "localhost:61111", "address:port where server listens for gRPC requests.")
	useServer = flag.Bool("server", false, "If true, the application runs the gRPC file server. If false, it runs the client instead.")
)

var (
	errInvalidParams = errors.New("invalid parameter values provided - see 'help' for usage")
	errNilFile       = errors.New("cannot perform file operations on nil File pointers")
	errNegativeParam = errors.New("integer parameters to Read/Seek must be positive")
)

// serve sets up a gRPC server which serves until it crashes
func serve() error {
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		return fmt.Errorf("failed setting up TCP listener: %w", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFileSystemServer(grpcServer, fsserver.NewFileSystemServer())

	fmt.Printf("Serving gRPC at '%s'.\n", *addr)

	return grpcServer.Serve(listener)
}

var clientCliDoc = map[string]string{
	"help": `With this CLI you can open/close files and perform the read, write and seek operations. Only a single file can be "in focus" at the same time. If you open another file before closing the previous, it will no longer be accessible.
The following shorthand function names are available: help (h), open (o), close (c), read (r), write (w), seek (s), lookup (ls), remove (rm).
The following operations are supported:
	help                      displays this prompt.`,
	"open":  "\topen <r|w|rw> <name>      opens the file 'name' as read-only (r), write-only (w) or read-write (rw). E.g.: 'open rw somefile.txt'",
	"close": "\tclose                     closes the most recently opened file.",
	"read":  "\tread <n>                  reads 'n' bytes from the most recently opened file.",
	"write": "\twrite <rest>              writes 'rest' to the most recently opened file.",
	"seek": `	seek <offset> <whence>    seek sets the offset within the most recently opened file to 'offset'.
	                          The optional 'whence' argument can be 0 for offset relative to the start of the file (default), 
	                          1 for offset relative to the current offset or 2 for offset relative to the end of the file.`,
	"mkdir":  "\tmkdir <name>              makes the directory 'name'.",
	"lookup": "\tlookup <name>             looks up  the file/directory 'name' and display its size (file) or content (directory).",
	"remove": "\tremove <name>             removes the file/directory 'name'.",
	"quit":   "\tquit                      stops processing commands. Synonyms: q, exit, stop",
}

func generateCLIDocumentation() string {
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n",
		clientCliDoc["help"],
		clientCliDoc["open"],
		clientCliDoc["close"],
		clientCliDoc["read"],
		clientCliDoc["write"],
		clientCliDoc["seek"],
		clientCliDoc["mkdir"],
		clientCliDoc["lookup"],
		clientCliDoc["remove"],
		clientCliDoc["quit"])
}

// fileProcessor takes user input to perform operations on the gRPC filesystem
func fileProcessor(fs *grpcfs.FileSystem) {
	helpStr := generateCLIDocumentation()
	fmt.Print(helpStr)

	// helper function for determining if enough parameters were provided
	missing := func(params []string, min int) bool {
		if len(params) < min {
			fmt.Println("Not enough parameters provided to perform the operation. See 'help' for usage.")
			return true
		}
		return false
	}

	scanner := bufio.NewScanner(os.Stdin)
	var f *grpcfs.File
	// print the prompt before the first scan
	fmt.Print("> ")
	// repeatedly perform operations on f until client quits
	for scanner.Scan() {
		// split the string into space-separated params
		params := strings.Split(strings.TrimSpace(scanner.Text()), " ")
		switch params[0] {
		case "h", "help":
			fmt.Print(helpStr)

		case "open", "o":
			if missing(params, 3) {
				break
			}

			if newf, err := handleOpen(f, fs, params[1], strings.Join(params[2:], " ")); err != nil {
				fmt.Println(err)
			} else {
				f = newf
			}

		case "close", "c":
			if f == nil {
				fmt.Println(errNilFile)
				break
			}

			err := f.Close()
			if err != nil {
				fmt.Printf("failed to close file: %v\n", err)
			} else {
				fmt.Println("closed file ")
			}

		case "read", "r":
			if missing(params, 2) {
				break
			}

			if err := handleRead(f, params[1]); err != nil {
				fmt.Println(err)
			}

		case "write", "w":
			if missing(params, 2) {
				break
			}

			if err := handleWrite(f, params[1:]...); err != nil {
				fmt.Println(err)
			}

		case "seek", "s":
			if missing(params, 2) {
				break
			}

			var whence string
			if len(params) > 2 {
				whence = params[2]
			}

			if err := handleSeek(f, params[1], whence); err != nil {
				fmt.Println(err)
			}

		case "mkdir":
			if missing(params, 2) {
				break
			}

			name := strings.Join(params[1:], " ")
			err := fs.Mkdir(name, os.ModePerm)
			if err != nil {
				fmt.Printf("failed to make directory: %v\n", err)
			} else {
				fmt.Printf("made directory '%s'\n", name)
			}

		case "lookup", "ls":
			var name string // path defaults to "", i.e. root directory
			if len(params) > 1 {
				name = strings.Join(params[1:], " ")
			}

			if err := handleLookup(fs, name); err != nil {
				fmt.Println(err)
			}

		case "remove", "rm":
			if missing(params, 2) {
				break
			}

			name := strings.Join(params[1:], " ")
			err := fs.Remove(name)
			if err != nil {
				fmt.Printf("failed to remove '%s': %v\n", name, err)
			} else {
				fmt.Printf("'%s' removed\n", name)
			}

		case "quit", "q", "exit", "stop":
			fmt.Println("Quitting client...")
			if f != nil {
				defer f.Close()
			}
			return

		default:
			fmt.Println("Invalid operation. See 'help' for usage.")
		}

		// print the prompt
		fmt.Print("> ")
	}
}

func handleOpen(f *grpcfs.File, fs *grpcfs.FileSystem, flag, name string) (*grpcfs.File, error) {
	var openFlag int
	switch flag {
	case "r":
		openFlag = grpcfs.OpenRead
	case "w":
		openFlag = grpcfs.OpenWrite
	case "rw", "wr":
		openFlag = grpcfs.OpenReadWrite
	default:
		return nil, errInvalidParams
	}

	newf, err := fs.Open(name, openFlag)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", name, err)
	}

	if f != nil {
		// if there was already an open file, we close it first
		fmt.Println("closing previously opened file")
		if err = f.Close(); err != nil {
			fmt.Printf("failed to close previously opened file: %v\n", err)
		}
	}
	fmt.Printf("opened %s\n", name)

	// caller is responsible for using the newly opened file
	return newf, nil
}

func handleRead(f *grpcfs.File, bufLenStr string) error {
	if f == nil {
		return errNilFile
	}

	bufLen, err := parseParameter("n", bufLenStr)
	if err != nil {
		return err
	}

	b := make([]byte, bufLen)
	n, err := f.Read(b)

	if (err != nil && !errors.Is(err, io.EOF)) || (n == 0 && errors.Is(err, io.EOF)) {
		// return if non-EOF error or if EOF and nothing is read
		return fmt.Errorf("failed to read: %w", err)
	}

	fmt.Println(string(b[:n]))

	return nil
}

func handleWrite(f *grpcfs.File, text ...string) error {
	n, err := f.Write([]byte(strings.Join(text, " ")))
	if err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	fmt.Printf("wrote %d bytes\n", n)

	return nil
}

func handleSeek(f *grpcfs.File, offsetAsStr, whenceAsStr string) error {
	if f == nil {
		return errNilFile
	}

	offsetInt, err := parseParameter("offset", offsetAsStr)
	if err != nil {
		return err
	}
	offset := int64(offsetInt)

	// whence is optionally provided, defaulting to 0
	whence, err := parseParameter("whence", whenceAsStr)
	if err != nil {
		return err
	}

	offset, err = f.Seek(offset, whence)
	if err != nil {
		return fmt.Errorf("failed to seek: %v", err)
	}

	fmt.Printf("new offset: %d\n", offset)

	return nil
}

func handleLookup(fs *grpcfs.FileSystem, name string) error {
	isDir, n, files, err := fs.Lookup(name)
	if err != nil {
		return fmt.Errorf("failed lookup: %w", err)
	}

	if isDir {
		if len(files) == 0 {
			fmt.Println("empty directory")
		}

		for _, file := range files {
			fmt.Printf("%s\n", file)
		}
	} else {
		fmt.Printf("%s\t%d bytes\n", name, n)
	}

	return nil
}

// parseParamater parses `val` to int.
// Negative `val` and conversion failure causes errors.
func parseParameter(parName, val string) (value int, err error) {
	if val == "" {
		// no value defaults to 0
		return 0, nil
	}

	value, err = strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("failed to convert the '%v' parameter: %w", parName, err)
	}

	if value < 0 {
		return 0, errNegativeParam
	}

	return
}

func main() {
	flag.Parse()

	if *useServer {
		if err := serve(); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	} else {
		// open file system client
		fs, err := grpcfs.NewFileSystem("localhost:61111")
		if err != nil {
			log.Fatalf("failed to start gRPC client: %v", err)
		}

		// accept user input and translate it into RPCs executed at the server
		fileProcessor(fs)
	}
}
