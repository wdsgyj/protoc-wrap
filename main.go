// protoc-wrap project main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "no input...")
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "micro":
		err = runJavaMicroProgram(os.Args[2:])
	case "nano":
		err = runJavaNanoProgram(os.Args[2:])
	case "lite":
		err = runJavaLiteProgram(os.Args[2:])
	default:
		err = runJavaProgram(os.Args[1:])
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMicroOrNanoPrograme(args []string, fn func(string, []string, []string) *programeState) error {
	fs := flag.NewFlagSet("protoc-wrap", flag.ExitOnError)

	flagProtoDir := fs.String("proto", ".", "directory include *.proto.1 files")
	flagProtocArgs := fs.String("arg", "", "args give protoc")
	flagRuntimeJars := fs.String("runtime", "", "protobuf.jar & json.jar")

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	program := fn(*flagProtoDir,
		_B(*flagProtocArgs == "", []string{},
			strings.Split(*flagProtocArgs, ",")).([]string),
		_B(*flagRuntimeJars == "", []string{},
			strings.Split(*flagRuntimeJars, string(os.PathListSeparator))).([]string))
	//	program.debug = true
	return program.run()
}

func runJavaMicroProgram(args []string) error {
	return runMicroOrNanoPrograme(args, NewJavaMicroPrograme)
}

func runJavaNanoProgram(args []string) error {
	return runMicroOrNanoPrograme(args, NewJavaNanoPrograme)
}

func runJavaOrLiteProgram(args []string, lite bool) error {
	fs := flag.NewFlagSet("protoc-wrap", flag.ExitOnError)

	flagProtoDir := fs.String("proto", ".", "directory include *.proto.1 files")
	flagRuntimeJars := fs.String("runtime", "", "protobuf.jar & json.jar")

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	program := NewJavaPrograme(*flagProtoDir,
		_B(*flagRuntimeJars == "", []string{},
			strings.Split(*flagRuntimeJars, string(os.PathListSeparator))).([]string),
		lite)

	return program.run()
}

func runJavaLiteProgram(args []string) error {
	return runJavaOrLiteProgram(args, true)
}

func runJavaProgram(args []string) error {
	return runJavaOrLiteProgram(args, false)
}
