# 🗑️ Trash

A **blazingly fast** and **safe** alternative to `rm`, written in Go! 🚀

## ✨ Features

- 🛡️ **Safer deletions**: No more accidental data loss! Files and directories
  are moved to your system's trash instead of being permanently deleted.
- ⚡ **Blazingly fast**: Leverages the power of Go for
  lightning-fast performance.
- 📜 **XDG Trash Specification**: Fully compatible with the XDG Trash spec for
  cross-platform support.
- 💡 **Intuitive**: Seamlessly integrates into your workflow as a drop-in
  replacement for `rm`.

---

## 🚀 Installation

### Use a pre-built release binary (recommended)

Get the latest release from the
[releases page](https://github.com/rhogenson/trash/releases). Pre-built binaries
are available for Linux, Mac OS, and Windows.

### Install from source

Or install from source:

```
go install roseh.moe/cmd/trash@latest
```

---

## 🛠️ Usage

Replace `rm` with `trash` in your terminal commands for a safer experience:

```bash
trash somefile1.txt somefile2.txt somefile3.txt
trash somedirectory/
```

### Recovering Files

Files can be recovered from your system's trash in the
`~/.local/share/Trash/files/` directory:

```bash
ls ~/.local/share/Trash/files  # List all trashed files
mv ~/.local/share/Trash/files/somefile1.txt.* ./somefile1.txt  # Restore a trashed file
```

---

## 🌟 Why Use Trash CLI?

- 🧘 **Peace of mind**: No more oops moments when using `rm`.
- 🏃 **Fast and efficient**: Written in Go, optimized for performance.
- 🌎 **Portable**: Compatible with the XDG Trash spec.

---

## 💻 Compatibility

Trash CLI is compatible with:
- Linux 🐧
- macOS 🍎
- WSL 🌐

---

## 📄 License

This project is licensed under GPLv3 or later. See the [COPYING](COPYING) file for details.
