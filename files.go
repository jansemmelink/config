package config

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/pkg/errors"
)

var (
	bufferSize = int64(1024)
)

//FileCopy a file (modified from https://github.com/mactsouk/opensource.com/blob/master/cp3.go)
//Copies file src to file dst
//When allowOverwrite == false, this function will fail if dst already exists
func FileCopy(src, dst string, allowOverwrite bool) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return errors.Wrapf(err, "Failed to stat source %s", src)
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("Source %s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return errors.Wrapf(err, "Failed to open source %s", src)
	}
	defer source.Close()

	if !allowOverwrite {
		_, err = os.Stat(dst)
		if err == nil {
			return fmt.Errorf("Destination %s already exists", dst)
		}
	}

	destination, err := os.Create(dst)
	if err != nil {
		return errors.Wrapf(err, "Failed to create destination %s", dst)
	}
	defer destination.Close()

	buf := make([]byte, bufferSize)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return errors.Wrapf(err, "Failed to read from source %s", src)
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return errors.Wrapf(err, "Failed to write to destination %s", dst)
		}
	}
	return nil
} //FileCopy()

//Mkdir makes a directory and all its parents
func Mkdir(name string, perm os.FileMode) error {
	//if exists, return success
	info, err := os.Stat(name)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("Mkdir(%s,%v) exists but is not a directory", name, perm)
		}
		return nil
	}

	parent := path.Dir(name)
	if parent != "" {
		if err := Mkdir(parent, perm); err != nil {
			return err
		}
	}

	//parent exists
	if err := os.Mkdir(name, perm); err != nil {
		return errors.Wrapf(err, "Mkdir(%s,%v)", name, perm)
	}

	//created
	return nil
} //Mkdir()
