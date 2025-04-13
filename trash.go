package main

import (
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func cp(srcRoot string, dstRoot string) (err error) {
	defer func() {
		if err != nil {
			os.RemoveAll(dstRoot)
		}
	}()

	type roDir struct {
		path string
		mode os.FileMode
	}
	var roDirs []roDir

	srcRoot = filepath.Clean(srcRoot)
	if err := filepath.WalkDir(srcRoot, func(src string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dst := filepath.Join(dstRoot, strings.TrimPrefix(src, srcRoot))
		switch d.Type() {
		case 0: // regular file
			srcF, err := os.Open(src)
			if err != nil {
				return err
			}
			defer srcF.Close()
			stat, err := srcF.Stat()
			if err != nil {
				return err
			}
			dstF, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, stat.Mode().Perm())
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
		case os.ModeSymlink:
			linkTarget, err := os.Readlink(src)
			if err != nil {
				return err
			}
			return os.Symlink(linkTarget, dst)
		case os.ModeDir:
			stat, err := d.Info()
			if err != nil {
				return err
			}
			perm := stat.Mode().Perm()
			if perm&0300 != 0300 {
				roDirs = append(roDirs, roDir{dst, perm})
				// Make sure we can create directory contents.
				perm |= 0300
			}
			return os.MkdirAll(dst, perm)
		default:
			return fmt.Errorf("unknown file type %s", d.Type())
		}
	}); err != nil {
		return err
	}
	// Iterate backwards to process directory contents before the parent directory.
	for i := len(roDirs) - 1; i >= 0; i-- {
		dir := roDirs[i]
		if err := os.Chmod(dir.path, dir.mode); err != nil {
			return err
		}
	}
	return nil
}

func mv(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := cp(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

func trashFile(fileName, trash, now string) (err error) {
	absPath, err := filepath.Abs(fileName)
	if err != nil {
		return fmt.Errorf("%s: find absolute path: %s", fileName, err)
	}
	info, err := os.CreateTemp(filepath.Join(trash, "info"), filepath.Base(fileName)+"."+now+".*.trashinfo")
	if err != nil {
		return fmt.Errorf("%s: create trashinfo: %s", fileName, err)
	}
	defer func() {
		info.Close()
		if err != nil {
			os.Remove(info.Name())
		}
	}()
	escapedPath := strings.Split(absPath, string(filepath.Separator))
	for i, pathSegment := range escapedPath {
		escapedPath[i] = url.QueryEscape(pathSegment)
	}
	_, err = fmt.Fprintf(info, `[Trash Info]
Path=%s
DeletionDate=%s
`,
		filepath.Join(escapedPath...),
		now)
	if err != nil {
		return fmt.Errorf("%s: write trashinfo: %s", fileName, err)
	}
	if err := info.Close(); err != nil {
		return fmt.Errorf("%s: write trashinfo: %s", fileName, err)
	}
	if err := mv(fileName, filepath.Join(trash, "files", strings.TrimSuffix(filepath.Base(info.Name()), ".trashinfo"))); err != nil {
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

	trash := filepath.Join(os.Getenv("HOME"), ".local", "share", "Trash")
	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		trash = filepath.Join(xdgDataHome, "Trash")
	}
	if err := os.MkdirAll(filepath.Join(trash, "files"), 0755); err != nil {
		fmt.Fprintln(os.Stderr, "trash:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Join(trash, "info"), 0755); err != nil {
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
