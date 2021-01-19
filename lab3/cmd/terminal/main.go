package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Task 7: Simple Shell
//
// This task focuses on building a simple shell that accepts
// commands that run certain OS functions or programs. For OS
// functions refer to golang's built-in OS and ioutil packages.
//
// The shell should be implemented through a command line
// application; allowing the user to execute all the functions
// specified in the task. Info such as [path] are command arguments
//
// Important: The prompt of the shell should print the current directory.
// For example, something like this:
//   /Users/meling/Dropbox/work/opsys/2020/meling-stud-labs/lab3>
//
// We suggest using a space after the > symbol.
//
// Your program should be able to at least the following functions:
// 	- exit
// 		- exit the program
// 	- cd [path]
// 		- change directory to a specified path
// 	- ls
// 		- list items and files in the current path
// 	- mkdir [path]
// 		- create a directory with the specified path
// 	- rm [path]
// 		- remove a specified file or folder
// 	- create [path]
// 		- create a file with a specified name
// 	- cat [file]
// 		- show the contents of a specified file
// 			- any file, you can use the 'hello.txt' file to check if your
// 			  implementation works
// 	- help
// 		- show a list of available commands
//
// You may also implement any number of optional functions, here are some ideas:
// 	- help [command]
// 		- give additional info on a certain command
// 	- ls [path]
// 		- make ls allow for a specified path parameter
// 	- rm -r
// 		WARNING: Be aware of where you are when you try to execute this command
// 		- recursively remove a directory
// 			- meaning that if the directory contains files, remove
// 			  all the files within the directory first, then the
// 			  directory itself
// 	- calc [expression]
// 		- Simple calculator program that can calculate a given expression
// 			- example expressions could be + - * \ pow
// 	- ipconfig
// 		- show ip interfaces
// 	- history
// 		- show command history
// 		- Alternatively implement this together with pressing up on your
// 		  keypad to load the previous command
// 		- clrhistory to clear history
// 	- tail [n]
// 		- show last n lines of a file
// 	- head [n]
// 		- show first n lines of a file
// 	- writefile [text]
// 		- write specified text to a specified file
//
// 	Or, alternatively, implement your own functionality not specified as you please
//
// Additional notes:
// 	- If you want to use colors in your terminal program you can see the package
// 		"github.com/fatih/color"
//
// 	- Helper functions may lead to cleaner code
//

// Terminal contains
type Terminal struct {
	state int
	path  string
}

// Execute executes a given command
func (t *Terminal) Execute(command string) {
	cmd := strings.Split(command, " ")

	if cmd[0] == "exit" {
		os.Exit(0)
	}

	if cmd[0] == "ls" {
		files, _ := ioutil.ReadDir(t.path)
		for i := range files {
			fmt.Println(files[i].Name())
		}
	}

	if cmd[0] == "help" {

		fmt.Println("- exit				- exit the program")
		fmt.Println("- cd [folder] 			- change directory to a specified path")
		fmt.Println("- ls				- list items and files in the current path")
		fmt.Println("- mkdir [folder]		- create a directory with the specified path")
		fmt.Println("- rm [file/folder]		- remove a specified file or folder")
		fmt.Println("- create [file]			- create a file with a specified name")
		fmt.Println("- cat [file]			- show the contents of a specified file")
		fmt.Println("- help				- show a list of available commands")

	}

	if cmd[0] == "cd" {
		try := ""
		if strings.HasPrefix(cmd[1], "/") {
			try = cmd[1]
		} else {
			try = t.path + "/" + cmd[1]
		}
		_, err := os.Stat(try)
		if err == nil {
			t.path = try
		} else {
			fmt.Println(err)
		}
	}

	if cmd[0] == "mkdir" {
		ny_dir := os.Mkdir(t.path+"/"+cmd[1], 0700)
		if ny_dir != nil {
			fmt.Println(ny_dir)
		}
	}

	if cmd[0] == "cat" {
		n, err := ioutil.ReadFile(t.path + "/" + cmd[1])
		if err == nil {
			fmt.Println(string(n))
		} else {
			fmt.Println(err)
		}
	}

	if cmd[0] == "rm" {
		ny_rm := os.Remove(t.path + "/" + cmd[1])
		if ny_rm != nil {
			fmt.Println(ny_rm)
		}
	}

	if cmd[0] == "create" {
		_, err := os.Create(t.path + "/" + cmd[1])
		if err != nil {
			fmt.Println(err)
		}
	}
}

// This is the main function of the application.
// User input should be continuously read and checked for commands
// for all the defined operations.
// See https://golang.org/pkg/bufio/#Reader and especially the ReadLine
// function.
func main() {
	nv_mappe, _ := os.Getwd()
	terminal := Terminal{path: nv_mappe, state: 1}
	var commando string
	var argument string
	for terminal.state == 1 {
		fmt.Print(terminal.path + "> ")
		fmt.Scanln(&commando, &argument)
		if argument == "" {
			terminal.Execute(commando)
		} else {
			nytt_ord := commando + " " + argument
			terminal.Execute(nytt_ord)
		}
	}

}
