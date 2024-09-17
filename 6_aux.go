package moontpl

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/samber/lo"
)

func fsExists(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Print(err)
		}
		return false
	}
	defer file.Close()
	return true
}

func fsStat(filename string) (fs.FileInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil

		}
		return nil, err
	}
	defer file.Close()
	return file.Stat()
}

func mustGetwd() string {
	return lo.Must(os.Getwd())
}

func mustAbs(filename string) string {
	return lo.Must(filepath.Abs(filename))
}

func mustRel(basePath, targetPath string) string {
	basePath = mustAbs(basePath)
	targetPath = mustAbs(targetPath)
	return lo.Must(filepath.Rel(basePath, targetPath))
}

func isDirectory(path string) bool {
	stat := lo.Must(fsStat(path))
	if stat != nil {
		return stat.IsDir()
	}
	return false
}

func isSubDirectory(base, sub string) bool {
	if len(base) > 0 && base[len(base)-1] != os.PathSeparator {
		base += string(os.PathSeparator)
	}
	return strings.HasPrefix(sub, base)
}

func respondInternalError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "error: %v", err)
	log.Print(err)
	debug.PrintStack()
}

func apply1[A, B any](x A, f1 func(A) B) B {
	return f1(x)
}
func apply2[A, B, C any](x A, f1 func(A) B, f2 func(B) C) C {
	return f2(f1(x))
}
func apply3[A, B, C, D any](x A, f1 func(A) B, f2 func(B) C, f3 func(C) D) D {
	return f3(f2(f1(x)))
}

// Similar to filepath.Ext, but returns the whole file extension.
// Example:
//
//	wholeExt("/dir/filename.html.lua")     == ".html.lua"
//	filepath.Ext("/dir/filename.html.lua") == ".lua"
func wholeExt(path string) string {
	foundIndex := -1
	for i := len(path) - 1; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			foundIndex = i
		}
	}
	if foundIndex >= 0 {
		return path[foundIndex:]
	}
	return ""
}
