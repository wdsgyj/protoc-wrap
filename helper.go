package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func absPath(path string) string {
	rs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}

	return rs
}

func isExist(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		// 所有错误都认为是不存在的
		return false
	}

	return info.IsDir() || info.Mode().IsRegular()
}

func runExec(exe string, args ...string) (string, error) {
	cmd := exec.Command(exe, args...)
	bs, err := cmd.CombinedOutput()
	return string(bs), err
}

func isProtocClarkVersion() bool {
	output, err := runExec("protoc", "--version")
	if err == nil {
		return strings.Contains(output, "(clark modify version)")
	}

	return false
}

func isDir(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func _C(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	} else {
		return name
	}
}

func _B(condition bool, a interface{}, b interface{}) interface{} {
	if condition {
		return a
	} else {
		return b
	}
}
