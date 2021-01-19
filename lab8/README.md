# Lab 8: Distributed File System With gRPC

| Lab 8: | Distributed File System With gRPC |
| ---------------------    | --------------------- |
| Subject:                 | DAT320 Operating Systems and Systems Programming |
| Deadline:                | **November 13, 2020 23:59** |
| Expected effort:         | 15-30 hours |
| Grading:                 | Pass/fail |
| Submission:              | Group |

## Table of Contents

1. [Introduction](#introduction)
2. [Remote Procedure Calls](#remote-procedure-calls)
3. [Simplified File System over gRPC](#simplified-file-system-over-grpc)
4. [Installing the Protobuf Compiler and Plugins](#installing-the-protobuf-compiler-and-plugins)
5. [Compiling .proto Files](#compiling-.proto-files)
6. [The Simple Shell](#the-simple-shell)
7. [Compiling the Simple Shell Application](#compiling-the-simple-shell-application)
8. [Running the Simple Shell](#running-the-simple-shell)
9. [Note on gRPC Errors](#note-on-grpc-errors)
10. [Task 1: File System Server: Unary RPCs](#task-1-file-system-server-unary-rpcs)
11. [Task 2: File System gRPC Client](#task-2-file-system-grpc-client)
12. [Task 3: File Abstraction and `Reader`/`Writer` Streams](#task-3-file-abstraction-and-readerwriter-streams)

## Introduction

This lab introduces you to the gRPC remote procedure call (RPC) framework.
Using gRPC, you will implement a simplified file system API in which several clients can perform file operations remotely on the server.
Some of the principles applied in the client-server interaction are inspired by the Network File System (NFS) as described in [Chapter 49 in the textbook](http://pages.cs.wisc.edu/~remzi/OSTEP/dist-nfs.pdf), which we recommend you to read.

## Remote Procedure Calls

A popular way to design distributed applications is by means of remote procedure calls.
A remote procedure call allows a client application to invoke a server's procedure or method almost as if the server's method was local to the client application.
The main difference between local procedure calls and remote procedure calls is that clients (and servers) need to perform some setup before procedures can be called, and clients must be equipped to handle errors due to network delays or disconnection.

gRPC is an open source framework for working with remote procedure calls that can interact across different languages and operating systems.
Before getting started you should read through the following documents:

- [Introduction to gRPC](https://grpc.io/docs/what-is-grpc/introduction/):
  This gives a brief overview of how gRPC works and how to work with protocol buffers.
- [gRPC Quick Start (Go)](https://grpc.io/docs/languages/go/quickstart/):
  Guides you through setting up the gRPC working environment in Go and installing the `protoc` protocol buffer compiler.
- [gRPC Basics Tutorial (Go)](https://grpc.io/docs/languages/go/basics/):
  This document contains several important pieces of information:

  - Overview of service definitions and code generation.
  - Distinction between simple RPCs, server-side streaming RPCs, client-side streaming RPCs and bidirectional streaming RPCs.
  - Creating and starting the server.
    NOTE: We have already set up the skeleton code for the server implementation and started the server in the application.
  - Creating the client.
    NOTE: We have already set up the `FileSystem` with a gRPC client stub.

- [Protocol buffers](https://developers.google.com/protocol-buffers/) and [Protocol Buffer Basics: Go](https://developers.google.com/protocol-buffers/docs/gotutorial):
  You should get familiar with what protocol buffers are and how to implement them.
  These links introduce protocol buffers, why they are useful, how to define and compile them, and how to use the generated message types in Go.

Other useful resources:

- [API (`package grpc`)](https://pkg.go.dev/google.golang.org/grpc?tab=doc):
  This is the API of the `grpc` package in Go, which is used by the generated code as well as the portions of your application that interact with gRPC.
- [Language Guide (proto3)](https://developers.google.com/protocol-buffers/docs/proto3)
  Detailed specification of the `proto3` language syntax.

## Simplified File System over gRPC

Now that you are familiar with gRPC you will implement a simplified file system which operates over gRPC, as described in the introduction.
The application will consist of five parts:

1. Generated gRPC code.
   In gRPC you specify messages and RPCs in a *protocol buffer (protobuf)* file (uses the `.proto` extension).
   Then you can use the protobuf compiler `protoc` to generate compliant code in several languages, including Go.
   It will generate code such as structs, constants and interfaces, which you will use as the basis for your implementation.

2. The [`FileSystemServer`](./grpcfs/fsserver/server.go):
   The file system server accepts incoming connections from gRPC clients and processes their requests.
   It implements the gRPC server interface.
   The file system server will manipulate its local file system by executing operations requested by clients, such as creating directories or reading files.
   The server should not depend on the state of clients; Clients must provide all necessary information to complete a request, such that the server can complete operations without tracking previous operations performed by each client.
   This makes the server more resistant to crashes, for the same reasons as [NFS as described in textbook](http://pages.cs.wisc.edu/~remzi/OSTEP/dist-nfs.pdf).

3. The [`FileSystem`](./grpcfs/fs/fs.go):
   The file system is responsible for opening files (opening `Reader` and `Writer` gRPC streams), making directories, removing files/directories and looking up files/directories.
   The file system is a wrapper around the *gRPC client stub*.
   When compiling the protobuf file, a client stub which knows how to communicate with the server is automatically generated.
   The client stub implements methods for any RPCs defined in the protobuf file.
   Additionally, streaming RPCs implement the `Send` and `Recv` methods for communication over streams.

4. The [`File`](./grpcfs/fs/file.go):
   The file is an abstraction for performing file operations on the remote file system.
   The `Read` and `Write` methods use underlying `Reader` and/or `Writer` bidirectional streaming RPCs, which were opened by the `FileSystem` when the `File` was opened.
   Essentially, as long as the file is open, the stream(s) are open, and any `Read` or `Write` request on the `File` will send a message on the `Reader` or `Writer` gRPC stream to the server using the `Send` method, providing all necessary information to perform the request.
   Then the server responds with the result on the same stream, which the client can receive by using the `Recv` method.

5. A [simple shell to interact with the application](./grpcfs/cmd/fs/main.go) is already implemented.
   You can give commands such as `open`, `write` and `mkdir` to the shell, which will then use your implementation of the application to perform the operations and show the output.
   Instructions on how to run this application are given below.

The API is defined in [`fs.proto`](grpcfs/proto/fs.proto) (the file also contains comments with descriptions):

```proto3
syntax = "proto3";

package proto;
option go_package = "dat320/lab8/grpcfs/proto";

service FileSystem {
  rpc Lookup(LookupRequest) returns (LookupResponse) {}
  rpc Create(CreateRequest) returns (CreateResponse) {}
  rpc Reader(stream ReadRequest) returns (stream ReadResponse) {}
  rpc Writer(stream WriteRequest) returns (stream WriteStatus) {}
  rpc Remove(RemoveRequest) returns (RemoveResponse) {}
  rpc Mkdir(MakeDirRequest) returns (MakeDirResponse) {}
}

message LookupRequest {
  string path = 1;
}

message LookupResponse {
  bool is_dir = 1;
  int64 size = 2;
  repeated string files = 3;
}

message CreateRequest {
  string path = 1;
}

message CreateResponse {}

message ReadRequest {
  // TODO(student): add necessary fields here
}

message ReadResponse {
  // TODO(student): add necessary fields here
}

message WriteRequest {
  // TODO(student): add necessary fields here
}

message WriteStatus {
  // TODO(student): add necessary fields here
}

message RemoveRequest {
  string path = 1;
}

message RemoveResponse {}

message MakeDirRequest {
  string path = 1;
  uint32 file_mode = 2;
}

message MakeDirResponse {}
```

As you can see, the RPCs and messages are already defined.
The API will be explained as part of the following task descriptions.
In later tasks you will need to add fields to the `ReadRequest`, `ReadResponse`, `WriteRequest` and `WriteStatus` messages.

### Installing the Protobuf Compiler and Plugins

The protobuf compiler is called `protoc`, and you will need to install this compiler for this assignment.
Most Linux distributions provides a `protobuf` package.

```shell
$ apt install -y protobuf-compiler
$ protoc --version
libprotoc 3.13.0  # Ensure version is 3+
```

On macOS, if you have installed homebrew, you can simply run:

```shell
$ brew install protobuf
$ protoc --version
libprotoc 3.13.0  # Ensure version is 3+
```

If you do not use a package manager with your OS, you should download the appropriate package from the [official release page of the Protobuf compiler](https://github.com/protocolbuffers/protobuf/releases).
Make sure to test that the installation is working by running:

```shell
$ protoc --version
libprotoc 3.13.0  # Ensure version is 3+
```

Next, you need to install the plugins that are needed to generate Go protobuf code and gRPC code.
This can be done with the following command:

```shell
$ go get google.golang.org/protobuf/cmd/protoc-gen-go \
         google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

This will install the `protoc-gen-go` and `protoc-gen-go-grpc` commands in your `$GOPATH/bin` folder.
To test whether or not you can use these plugins, run:

```shell
$ protoc-gen-go --version
protoc-gen-go v1.25.0
$ protoc-gen-go-grpc --version
protoc-gen-go-grpc 1.0.1
```

If the plugins are not found, then you need to add the following line to your shell's configuration file:

```shell
export PATH="$PATH:$(go env GOPATH)/bin"
```

- If you are using `zsh`, add the line above to your `$HOME/.zshrc` file.
- If you are using `bash`, add the line above to your `$HOME/.bashrc` file.

### Compiling .proto Files

In the rest of this assignment, whenever you make changes to the [`fs.proto`](./grpcfs/proto/fs.proto) file, you need to recompile it for the changes to take effect (become visible to the Go client/server implementation part):

```console
protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. fs.proto
```

### The Simple Shell

We have implemented a simple shell for interacting with the distributed file system.
The implementation is in [grpcfs/cmd/fs/main.go](./grpcfs/cmd/fs/main.go).
This application can run in two modes, either as the server or as a client.
As you implement your code, you can use this application to test functionality of your file system, in addition to the tests we provide.

### Compiling the Simple Shell Application

To compile the application, you can simply install it.

```console
go install dat320/lab8/grpcfs/cmd/fs
```

This will create a binary file called `fs` in your `$GOBIN` or `$GOPATH/bin` folder.
To make this command available ensure that your shell finds it in its `$PATH`.

### Running the Simple Shell

Run the gRPC server:

```console
fs -server
```

Then, open another terminal and run the application as a client:

```console
fs
```

You can run multiple clients in different terminal windows to access the same files.

### Note on gRPC Errors

Since gRPC works across multiple programming languages, error handling may depend on the language.
In Go, errors are values, and we like to be able to compare errors or check whether they wrap some error we expect, e.g. using the `errors.Is` function.
gRPC instead uses `Status`es (see [`package status`](https://godoc.org/google.golang.org/grpc/status)) which describe the error and use `Code`s (see [`package codes`](https://godoc.org/google.golang.org/grpc/codes)) to indicate the type of error, e.g. `NotFound` or `AlreadyExists`.
If you simply return an error in an RPC it will be wrapped in a `Status` with the code `codes.Unknown`, making it difficult for the receiver to interpret.

For the reasons above we have implemented some helper functions `createStatusError` (in [`server.go`](./grpcfs/fsserver/server.go)) and `translateStatusError` (in [`fs.go`](./grpcfs/fs/fs.go)).
These functions add status codes to well-known Go errors such as `os.ErrNotExist`, and translate them back to well-known Go errors at the receiver.
In the following tasks, you should use these functions to wrap Go errors (sent by the server) and translate gRPC errors (received from the server).
**Failing to do this, several Autograder tests may fail**, since they expect well-known Go errors in certain test cases.

### Task 1: File System Server: Unary RPCs

In this task you will implement the methods for the unary RPCs `Lookup`, `Create`, `Remove` and `Mkdir` on the [server](./grpcfs/fsserver/server.go).
The server uses the [`blang/vfs/memfs` package](https://github.com/blang/vfs/blob/master/memfs/memfs.go) as its underlying file system, which is an in-memory file system.
All file operations should be performed on the `MemFS`, which is in the `FileSystemServer.fs` field.

The operations should do the following:

- `Lookup` looks up a file or directory at the requested `path` in the file system of the server.
  The result depends on whether the path points to a file or a directory:
  
  - For files `size` is the size of the file in bytes, `is_dir` is false, and `files` is empty (the zero value).
  - For directories, `is_dir` is true, `size` is the number of files/subdirectories within the directory, and `files` contains the name of each file/subdirectory within the directory.
  
  *Hint: You may find the `MemFS.Stat` and `MemFS.ReadDir` methods useful.*
  
- `Create` creates a file at the requested `path` in the file system of the server.
  If a file/directory already exists at the path, it returns an error instead.

  *Hint: You may find the `MemFS.OpenFile` method and the constants from the [flags to `OpenFile` in the `os` package](https://golang.org/pkg/os/#pkg-constants) useful.*
  
- `Remove` removes the file or directory at the requested `path` in the file system of the server.
  Contents of directories are recursively removed (similar to the `rm -r` Unix command).

- `Mkdir` makes a directory at the requested `path` in the file system of the server.
  If a file/directory already exists at the path, it returns an error instead.
  The `file_mode` field of `MakeDirRequest` is the same as an [`os.FileMode`](https://golang.org/pkg/os/#FileMode) value.

  *Hint: Since the underlying type of `os.FileMode` is `uint32`, you can safely typecast the `FileMode` field of the gRPC request to the `os.FileMode` type.*

Test your code from the [`fsserver` directory](./grpcfs/fsserver):

```console
go test
```

This will run the `TestCreate`, `TestMkdir`, `TestRemove` and `TestLookup` tests.

### Task 2: File System gRPC Client

In this task you will implement the `Open`, `Mkdir` and `Lookup` methods of the [file system](./grpcfs/fs/fs.go).
The methods use the gRPC client stub `fs.client`, which has already been initialized in `NewFileSystem`, to make RPC requests to the server.
The `Remove` method is already implemented and may be used as a reference.

The `Mkdir` and `Lookup` methods should do the following:

- `Mkdir(path string, perm os.FileMode)` sends an RPC request to make a directory at `path` with the `perm` as the file mode.

- `Lookup(path string)` sends an RPC request to look up the file or directory at `path`.

The `Open` method is a bit more involved.
`Open(name string, flag int)` opens `Reader` and/or `Writer` streams to the server, and initializes a `File` object.
Specification:

- To determine which streams to be opened, an `OpenRead`, `OpenWrite` or `OpenReadWrite` value must be passed as the `flag` parameter to `Open`.
  Any other values for `flag` should return an error.

- When `flag` is `OpenRead`:
  Open the file as read-only by only opening a `Reader` stream.
  The `rClient` field of the resulting `File` object should be set to the `Reader` stream to the gRPC server.
  If the remote file does not exist, `Open` should return an error.
  
- When `flag` is `OpenWrite`:
  Open the file as write-only by only opening a `Writer` stream.
  The `wClient` field of the resulting `File` object should be set to the `Writer` stream to the gRPC server.
  If the remote file does not exist, create it first by sending a `Create` RPC request.

- When `flag` is `OpenReadWrite`:
  Open the file as read-write by opening both a `Reader` stream and a `Writer` stream.
  The `rClient` and `wClient` fields of the resulting `File` object should be set to the respective `Reader` and `Writer` streams to the gRPC server.
  If the remote file does not exist, create it first by sending a `Create` RPC request.

- You may define additional fields in the `FileSystem` and `File` structs as you need.

Test your code from the [`fs` directory](./grpcfs/fs):

```console
$ go test -run TestMkdir
$ go test -run TestOpen
# NOTE: Some tests require File.Write to be implemented. Ignore these for now.
$ go test -run TestLookup
```

### Task 3: File Abstraction and `Reader`/`Writer` Streams

In this task you will implement the `Read`, `Write`, `Seek` and `Close` methods of the [file](./grpcfs/fs/file.go) according to the `Reader`, `Writer`, `Seeker` and `Closer` interfaces in the [`io` package](https://golang.org/pkg/io/).
By implementing these interfaces, your application will be very flexible, as any other Go code that operates on implementers of these interfaces can use your code.
Some examples:

- [`gzip.Reader`](https://golang.org/pkg/compress/gzip/#Reader) and [`gzip.Writer`](https://golang.org/pkg/compress/gzip/#Writer) take implementers of `io.Reader` and `io.Writer` as parameters to read and write gzip-compressed content.
- [`cipher.StreamReader`](https://golang.org/pkg/crypto/cipher/#StreamReader) and [`cipher.StreamWriter`](https://golang.org/pkg/crypto/cipher/#StreamWriter) uses `io.Writer` and `io.Reader` implementers to encrypt/decrypt streams of data.

As mentioned previously, the `File.Read` and `File.Write` are tied to the `Reader` and `Writer` bidirectional streaming RPCs.
Since the `ReadRequest`, `ReadResponse`, `WriteRequest` and `WriteStatus` messages were defined without any fields in [`fs.proto`](./grpcfs/proto/fs.proto), you will have to add any fields you deem necessary to implement the functionality described in this task.

Things to consider for the gRPC message fields:

- Review the [proto3 language guide](https://developers.google.com/protocol-buffers/docs/proto3) to determine how to define the fields.

- Think of each request as a remote call to `Read` or `Write`, which additionally specifies the offset within the file.
  Think of each response as the result of the operation.

- Read and write requests should provide all the information necessary for the server to complete the operation.
  As such you should consider fields for e.g. the path to the file to read from/write to, number of bytes to read, content to write and the offset.

- When the server encounters `io.EOF` errors during requests to the `Reader` or `Writer` streams it *must not* close the stream, but instead inform the client of the event through the stream.
  I.e. you probably need a field for this purpose.

Specification for [`File`](./grpcfs/fs/file.go):

- Implement the `Reader`, `Writer`, `Seeker` and `Closer` interfaces as specified in the [`io` package](https://golang.org/pkg/io/).

- The `Read` and `Write` methods must use the underlying `Reader` and `Writer` streams described in the previous task.

- If `Read` reads `0 < n < len(p)` during an end-of-file condition, it should return `n, io.EOF`.

- The offsets used and set by `Read`, `Write` and `Seek` must be the same, and must be stored on the client side (i.e. in the `File`) rather than the server side.
  E.g., if the offset is 0 and we call `f.Write([]byte("move by 10"))`, the offset will be 10 for the next call to `Read` or `Write`.
  Similarly, if the offset is 3 and we read 5 bytes (assuming there is something stored in the following 5 bytes) by calling `Read(make([]byte, 5))`, the updated offset will be 8 for the next call to `Read` or `Write`.

- The `Close` method must close any underlying `Reader` and `Writer` streams.
  Subsequent calls to `Close` should return an `os.ErrClosed` error.
  Calls to `Read`, `Write` and `Seek` on closed files should also return an `os.ErrClosed` error.

- Seeking to an offset before the start of the file is an error.
  Seeking to any positive offset is legal, but if the offset exceeds the file size, subsequent calls to `Read` or `Write` will cause an `io.EOF` error.

Specification for [`FileSystemServer`](./grpcfs/fsserver/server.go):

- *Hint: For the following tasks, you should look at the [API for `memfs.MemFile`](https://github.com/blang/vfs/blob/master/memfs/memfile.go) for performing file operations.*

- Implement the `Reader` and `Writer` methods.

- The first `ReadRequest` or `WriteRequest` sent on `Reader` or `Writer` stream should include the path to the file, and changing the path in any following request on the same stream is considered an error.

- Further implementation details for the `Reader` method:

  - The `Reader` receives a stream of `ReadRequest`s from gRPC clients by repeatedly calling `stream.Recv()` as long as the stream is open.
    Each `ReadRequest` comes from a call to `File.Read`, and should contain all the necessary information (e.g. path, number of bytes to read, offset) for the server to complete the request.
    The server performs the requested read and sends the result to the client with `stream.Send`.

  - In case of non-recoverable errors (i.e. not end-of-file errors) the server should close the stream by returning the error.

- Further implementation details for the `Writer` method:

  - The `Writer` receives a stream of `WriteRequest`s from gRPC clients by repeatedly calling `stream.Recv()` as long as the stream is open.
    Each `WriteRequest` comes from a call to `File.Write`, and should contain all the necessary information (e.g. path, content to write, offset) for the server to complete the request.
    The server performs the requested write and sends the result to the client with `stream.Send`.

  - In case of errors, the server should close the stream by returning the error.
  
Test your code from the [`fs` directory](./grpcfs/fs):

```console
$ go test -run TestClose
$ go test -run TestWrite
# we can use the regular expression \\b<TestName>\\b to prevent running TestReadWriteSeek as well
$ go test -run \\bTestReadWrite\\b
$ go test -run TestReadWriteSeek
```
