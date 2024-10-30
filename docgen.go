package moontpl

import (
	"bytes"
	"fmt"
	"io/fs"
	"iter"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

const luaDocMarker = "---"

func (m *Moontpl) extractDocumentation(filename string) (string, bool, error) {
	contents, err := fs.ReadFile(m.fsys, filename)
	if err != nil {
		return "", false, err
	}

	moduleName := filepath.Base(filename)
	moduleName = strings.TrimSuffix(moduleName, filepath.Ext(moduleName))

	pat := fmt.Sprintf(`function %s\.\w+()`, moduleName)
	funcRe := regexp.MustCompile(pat)

	var buf bytes.Buffer

	includeLast := false
	for line := range getLines(string(contents)) {
		line = strings.TrimSpace(line)
		include := false
		if strings.HasPrefix(line, luaDocMarker) {
			line = "" + line[len(luaDocMarker):]
			if line != "" {
				if unicode.IsSpace(rune(line[0])) {
					line = line[1:]
					line = "  | " + line
				} else {
					line = "" + line
				}
			} else {
				line = "  | "
			}
			include = true
		} else if strings.HasSuffix(line, luaDocMarker) {
			include = true
			line = line[:len(line)-len(luaDocMarker)]

		} else if funcRe.MatchString(line) {
			include = true
		}

		if include {
			buf.WriteString(line)
			buf.WriteString("\n")
		} else if includeLast {
			buf.WriteString("\n")
		}
		includeLast = include
	}

	buf.WriteRune('\n')

	result := strings.TrimRightFunc(buf.String(), unicode.IsSpace)

	header := "--------------------[ module: " + moduleName + "]--------------------\n\n"

	return header + result, result != "", nil
}

func getLines(s string) iter.Seq[string] {
	return func(yield func(string) bool) {
		a := 0
		for a < len(s) {
			b := a
			for ; b < len(s); b++ {
				if s[b] == '\n' {
					b++
					break
				}
			}
			yield(s[a:b])
			a = b
		}
	}
}
