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

func cpEntry(srcDirEntry os.DirEntry, src, dst string) error {
	switch srcDirEntry.Type() {
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
		return dstF.Close()
	case os.ModeSymlink:
		linkTarget, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(linkTarget, dst)
	case os.ModeDir:
		stat, err := srcDirEntry.Info()
		if err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		if err := os.Mkdir(dst, 0700); err != nil {
			return err
		}
		for _, entry := range entries {
			name := entry.Name()
			if err := cpEntry(entry, filepath.Join(src, name), filepath.Join(dst, name)); err != nil {
				return err
			}
		}
		return os.Chmod(dst, stat.Mode().Perm())
	default:
		return fmt.Errorf("unknown file type %s", srcDirEntry.Type())
	}
}

func cp(src, dst string) error {
	stat, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if err := cpEntry(fs.FileInfoToDirEntry(stat), src, dst); err != nil {
		os.RemoveAll(dst)
		return err
	}
	return nil
}

func writeTrashInfo(fileName, trash, now string) (_ string, err error) {
	absPath, err := filepath.Abs(fileName)
	if err != nil {
		return "", fmt.Errorf("find absolute path: %s", err)
	}
	info, err := os.CreateTemp(filepath.Join(trash, "info"), filepath.Base(fileName)+"."+now+".*.trashinfo")
	if err != nil {
		return "", fmt.Errorf("create trashinfo: %s", err)
	}
	defer func() {
		if err != nil {
			info.Close()
			os.Remove(info.Name())
		}
	}()
	escapedPath := strings.Split(absPath, string(filepath.Separator))
	for i, pathSegment := range escapedPath {
		escapedPath[i] = url.PathEscape(pathSegment)
	}
	_, err = fmt.Fprintf(info, `[Trash Info]
Path=%s
DeletionDate=%s
`,
		strings.Join(escapedPath, string(filepath.Separator)),
		now)
	if err != nil {
		return "", fmt.Errorf("write trashinfo: %s", err)
	}
	if err := info.Close(); err != nil {
		return "", fmt.Errorf("write trashinfo: %s", err)
	}
	return info.Name(), err
}

func trashFile(fileName, trash, now string) error {
	trashInfo, err := writeTrashInfo(fileName, trash, now)
	if err != nil {
		return fmt.Errorf("%s: %s", fileName, err)
	}
	dst := filepath.Join(trash, "files", strings.TrimSuffix(filepath.Base(trashInfo), ".trashinfo"))
	if err := os.Rename(fileName, dst); err == nil {
		return nil
	}
	if err := cp(fileName, dst); err != nil {
		os.Remove(trashInfo)
		return fmt.Errorf("%s: %s", fileName, err)
	}
	if err := os.RemoveAll(fileName); err != nil {
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
