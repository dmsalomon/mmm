package main

import (
	"io"
	"os"
)

func contains(c []string, e string) bool {
	for _, v := range c {
		if e == v {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func fexists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return false
	}

	return !fi.IsDir()
}

func moveFile(source, destination string) error {
	if DryRun {
		return nil
	}

	err := os.Rename(source, destination)
	if err == nil {
		return nil
	}

	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	fi, err := src.Stat()
	if err != nil {
		return err
	}

	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	perm := fi.Mode() & os.ModePerm
	dst, err := os.OpenFile(destination, flag, perm)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		dst.Close()
		os.Remove(destination)
		return err
	}

	err = dst.Close()
	if err != nil {
		return err
	}
	err = src.Close()
	if err != nil {
		return err
	}
	err = os.Remove(source)
	if err != nil {
		return err
	}
	return nil
}
