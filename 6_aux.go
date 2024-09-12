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
