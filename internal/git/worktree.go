package git

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsRepo reports whether dir is a git repository.
func IsRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

// ChangedFiles returns .go files changed vs base ref.
func ChangedFiles(dir, base string) ([]string, error) {
	cmd := exec.Command("git", "-C", dir, "diff", "--name-only", base+"...HEAD")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, ln := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		ln = strings.TrimSpace(ln)
		if strings.HasSuffix(ln, ".go") {
			files = append(files, filepath.Join(dir, ln))
		}
	}
	// include unstaged
	cmd2 := exec.Command("git", "-C", dir, "diff", "--name-only")
	if out2, err := cmd2.Output(); err == nil {
		for _, ln := range strings.Split(strings.TrimSpace(string(out2)), "\n") {
			ln = strings.TrimSpace(ln)
			if strings.HasSuffix(ln, ".go") {
				files = append(files, filepath.Join(dir, ln))
			}
		}
	}
	return files, nil
}

// CopyTree recursively copies src into dst, skipping .git.
func CopyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		if strings.HasPrefix(rel, ".git") || strings.Contains(rel, string(os.PathSeparator)+".git") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
