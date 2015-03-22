// state
package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	_PROTO_1              = ".proto.1"
	_PROTO_GEN_NAME       = "proto"
	_JAVA_SOURCE_GEN_NAME = "src"
	_CLASS_GEN_NAME       = "classes"
	_JAR_GEN_NAME         = "classes.jar"
)

type programeState struct {
	proto1Dir string
	protoDir  string

	genArgs       []string
	javaSourceDir string

	runtimes   []string
	classesDir string

	protocPath     string
	basicOutputDir string
	debug          bool
	outArg         string

	jarFile string
	islite  bool
}

func NewJavaPrograme(proto1Dir string,
	runtimes []string, lite bool) *programeState {
	return &programeState{
		proto1Dir: proto1Dir,
		basicOutputDir: _B(lite, fmt.Sprintf("out%c%s", os.PathSeparator, "javalite"),
			fmt.Sprintf("out%c%s", os.PathSeparator, "java")).(string),
		outArg:   "java_out",
		runtimes: runtimes,
		islite:   lite,
	}
}

func NewJavaMicroPrograme(proto1Dir string,
	genArgs, runtimes []string) *programeState {
	return &programeState{
		proto1Dir:      proto1Dir,
		basicOutputDir: fmt.Sprintf("out%c%s", os.PathSeparator, "javamicro"),
		outArg:         "javamicro_out",
		genArgs:        genArgs,
		runtimes:       runtimes,
	}
}

func NewJavaNanoPrograme(proto1Dir string,
	genArgs, runtimes []string) *programeState {
	return &programeState{
		proto1Dir:      proto1Dir,
		basicOutputDir: fmt.Sprintf("out%c%s", os.PathSeparator, "javanano"),
		outArg:         "javanano_out",
		genArgs:        genArgs,
		runtimes:       runtimes,
	}
}

func (s *programeState) printfln(format string, a ...interface{}) {
	if s.debug {
		fmt.Printf(format, a...)
		fmt.Println()
	}
}

func (s *programeState) run() error {
	var err error

	actions := []func() error{
		s.precondition,
		s.copyProtoFiles,
		s.genJavaSources,
		s.genClasses,
		s.getJar,
	}

	for _, action := range actions {
		err = action()
		if err != nil {
			return err
		}
	}
	fmt.Printf("Success! -> %s\n", s.jarFile)

	return nil
}

func (s *programeState) precondition() error {
	var err error

	// check *protoc* exist or not
	s.protocPath, err = exec.LookPath(_C("protoc"))
	if err != nil {
		return err
	}

	// check *javac* exist or not
	_, err = exec.LookPath(_C("javac"))
	if err != nil {
		return err
	}

	// check *jar* exist or not
	_, err = exec.LookPath(_C("jar"))
	if err != nil {
		return err
	}

	// check *.proto.1* dir exist or not
	if !isDir(s.proto1Dir) {
		return errors.New("proto.1 dir not found")
	}

	// check runtime jar file exist or not
	if len(s.runtimes) > 0 {
		for _, runtime := range s.runtimes {
			if !isExist(runtime) {
				return fmt.Errorf("%s not exist", runtime)
			}
		}
	}

	return nil
}

func (s *programeState) copyProtoFiles() error {
	s.printfln("remove %s ...", s.basicOutputDir)
	err := os.RemoveAll(s.basicOutputDir)
	if err != nil {
		return err
	}

	s.protoDir = filepath.Join(s.basicOutputDir, _PROTO_GEN_NAME)
	s.printfln("mkdir %s ...", s.protoDir)
	err = os.MkdirAll(s.protoDir, 0774)
	if err != nil {
		return err
	}

	s.printfln("copy from %s to %s ...", s.proto1Dir, s.protoDir)
	src, err := os.Open(s.proto1Dir)
	if err != nil {
		return err
	}
	defer src.Close()

	infos, err := src.Readdir(-1)
	if err != nil {
		return err
	}

	for _, info := range infos {
		name := info.Name()
		if info.Mode().IsRegular() && strings.HasSuffix(name, _PROTO_1) {
			err = copyOneProtoFile(filepath.Join(s.proto1Dir, name),
				filepath.Join(s.protoDir, name[:len(name)-2]), s.islite)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *programeState) genJavaSources() error {
	s.javaSourceDir = filepath.Join(s.basicOutputDir, _JAVA_SOURCE_GEN_NAME)
	s.printfln("mkdir %s ...", s.javaSourceDir)
	err := os.MkdirAll(s.javaSourceDir, 0774)
	if err != nil {
		return err
	}

	s.printfln("prepare protoc args ...")
	args := []string{}
	args = append(args, fmt.Sprintf("--proto_path=%s", s.protoDir))

	protocArgs := new(bytes.Buffer)
	if len(s.genArgs) > 0 {
		protocArgs.WriteString(strings.Join(s.genArgs, ","))
		protocArgs.WriteString(":")
	}
	protocArgs.WriteString(s.javaSourceDir)
	args = append(args, fmt.Sprintf("--%s=%s", s.outArg, protocArgs))

	err = filepath.Walk(s.protoDir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		if info.Mode().IsRegular() {
			args = append(args, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	s.printfln("%s %s", s.protocPath, strings.Join(args, " "))
	output, err := runExec(s.protocPath, args...)
	if err != nil {
		return fmt.Errorf("%s\n%s", err, output)
	}

	return nil
}

func (s *programeState) genClasses() error {
	s.classesDir = filepath.Join(s.basicOutputDir, _CLASS_GEN_NAME)
	s.printfln("mkdir %s ...", s.classesDir)
	err := os.MkdirAll(s.classesDir, 0774)
	if err != nil {
		return err
	}

	var commands = []string{
		"-g:none", "-encoding", "UTF-8", "-target", "1.6", "-source", "1.6",
		"-d", s.classesDir,
	}
	if len(s.runtimes) > 0 {
		commands = append(commands, "-cp", strings.Join(s.runtimes, string(os.PathListSeparator)))
	}
	err = filepath.Walk(s.javaSourceDir,
		func(path string, info os.FileInfo, e error) error {
			if e != nil {
				return e
			}

			name := info.Name()
			if info.Mode().IsRegular() && strings.HasSuffix(name, ".java") {
				commands = append(commands, path)
			}

			return nil
		})
	if err != nil {
		return err
	}

	s.printfln("%s %s", _C("javac"), strings.Join(commands, " "))
	output, err := runExec(_C("javac"), commands...)
	if err != nil {
		return fmt.Errorf("%s\n%s", err, output)
	}

	return nil
}

func (s *programeState) getJar() error {
	s.printfln("gen jar file ...")
	s.jarFile = filepath.Join(s.basicOutputDir, _JAR_GEN_NAME)
	output, err := runExec(_C("jar"),
		"cvf",
		s.jarFile,
		"-C",
		s.classesDir+string(os.PathSeparator), ".")
	if err != nil {
		return fmt.Errorf("%s\n%s", err, output)
	}
	return nil
}

// ---------------------------------------------------------
func copyOneProtoFile(src string, dst string, islite bool) error {
	const HEADER = `option java_package="com.baidu.entity.pb";
option java_multiple_files=true;
`
	const LITE_OPTION = "option optimize_for=LITE_RUNTIME;\n"

	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer w.Close()

	var flushAll = false
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := scanner.Text()

		if !flushAll {
			if strings.HasPrefix(s, "message") {
				w.WriteString(HEADER)
				if islite {
					w.WriteString(LITE_OPTION)
				}
				flushAll = true
			} else if !strings.HasPrefix(s, "import") {
				continue
			}
		}

		w.WriteString(s)
		w.WriteString("\n")
	}

	if e := scanner.Err(); e != nil {
		return e
	}

	return nil
}
