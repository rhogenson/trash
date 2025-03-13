package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func mv(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	src = filepath.Clean(src)
	if err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(path, src)
		dstPath := filepath.Join(dst, relPath)
		switch d.Type() {
		case fs.ModeDir:
			stat, err := d.Info()
			if err != nil {
				return err
			}
			return os.Mkdir(dstPath, stat.Mode().Perm())
		case fs.ModeSymlink:
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(linkTarget, dstPath)
		case 0:
			srcF, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcF.Close()
			stat, err := srcF.Stat()
			if err != nil {
				return err
			}
			dstF, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, stat.Mode().Perm())
			if err != nil {
				return err
			}
			defer dstF.Close()
			if _, err := io.Copy(dstF, srcF); err != nil {
				return err
			}
			if err := dstF.Close(); err != nil {
				return err
			}
			return nil
		default:
			return fmt.Errorf("unknown file type %s", d.Type())
		}
	}); err != nil {
		os.RemoveAll(dst)
		return err
	}
	return os.RemoveAll(src)
}

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
	if err := mv(fileName, trash+"/files/"+strings.TrimSuffix(filepath.Base(info.Name()), ".trashinfo")); err != nil {
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
		fmt.Fprintln(os.Stderr, "trash:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(trash+"/info", 0755); err != nil {
		fmt.Fprintln(os.Stderr, "trash:", err)
		os.Exit(1)
	}

	now := time.Now().Format("2006-01-02T15:04:05")
	success := true
	for _, fileName := range flag.Args() {
		if err := trashFile(fileName, trash, now); err != nil {
			fmt.Fprintln(os.Stderr, "trash:", err)
			success = false
		}
	}
	if !success {
		os.Exit(1)
	}
}
