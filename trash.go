package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func trashFile(fileName, trash, now string) error {
	absPath, err := filepath.Abs(fileName)
	if err != nil {
		return fmt.Errorf("%s: find absolute path: %s", fileName, err)
	}
	info, err := os.CreateTemp(trash+"/info", filepath.Base(fileName)+"."+now+".*.trashinfo")
	if err != nil {
		return fmt.Errorf("%s: create trashinfo: %s", fileName, err)
	}
	escapedPath := strings.Split(absPath, "/")
	for i, pathSegment := range escapedPath {
		escapedPath[i] = url.QueryEscape(pathSegment)
	}
	_, err = fmt.Fprintf(info, `[Trash Info]
Path=%s
DeletionDate=%s
`,
		strings.Join(escapedPath, "/"),
		now)
	if err != nil {
		info.Close()
		os.Remove(info.Name())
		return fmt.Errorf("%s: write trashinfo: %s", fileName, err)
	}
	if err := info.Close(); err != nil {
		os.Remove(info.Name())
		return fmt.Errorf("%s: write trashinfo: %s", fileName, err)
	}
	if err := os.Rename(fileName, trash+"/files/"+strings.TrimSuffix(filepath.Base(info.Name()), ".trashinfo")); err != nil {
		os.Remove(info.Name())
		return fmt.Errorf("%s: %s", fileName, err)
	}
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: trash [FILE]...")
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "trash: missing operand")
		flag.Usage()
		os.Exit(1)
	}

	trash := os.Getenv("HOME") + "/.local/share/Trash"
	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		trash = xdgDataHome + "/Trash"
	}
	if err := os.MkdirAll(trash+"/files", 0755); err != nil {
		fmt.Fprintln(os.Stderr, "trash: ", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(trash+"/info", 0755); err != nil {
		fmt.Fprintln(os.Stderr, "trash: ", err)
		os.Exit(1)
	}

	now := time.Now().Format("2006-01-02T15:04:05")
	success := true
	for _, fileName := range flag.Args() {
		if err := trashFile(fileName, trash, now); err != nil {
			fmt.Fprintln(os.Stderr, "trash: ", err)
			success = false
		}
	}
	if !success {
		os.Exit(1)
	}
}
