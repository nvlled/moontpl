package moontpl

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"runtime/debug"
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

func respondInternalError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "error: %v", err)
	log.Print(err)
	debug.PrintStack()
}
