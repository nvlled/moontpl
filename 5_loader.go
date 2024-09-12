package moontpl

import (
	"io/fs"
	"os"
	"path"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type Loader struct {
	fsys fs.FS
}

func initFsLoader(L *lua.LState, fsys fs.FS) {
	loader := &Loader{fsys}
	pkg := L.GetField(L.Get(lua.EnvironIndex), "package").(*lua.LTable)
	loaders := L.GetField(pkg, "loaders").(*lua.LTable)
	loaders.Append(L.NewFunction(loader.LoadFile))
}

func (l *Loader) LoadFile(L *lua.LState) int {
	name := L.CheckString(1)
	path, msg := l.loFindFile(L, name, "path")
	if len(path) == 0 {
		L.Push(lua.LString(msg))
		return 1
	}

	bytes, err := fs.ReadFile(l.fsys, path)
	if err != nil {
		L.Push(lua.LString(msg))
		return 1
	}

	fn, err1 := L.LoadString(string(bytes))
	if err1 != nil {
		L.RaiseError(err1.Error())
	}
	L.Push(fn)
	return 1
}

func (l *Loader) loFindFile(L *lua.LState, name, pname string) (string, string) {
	name = strings.Replace(name, ".", string(os.PathSeparator), -1)
	lv := L.GetField(L.GetField(L.Get(lua.EnvironIndex), "package"), pname)
	lpath, ok := lv.(lua.LString)
	if !ok {
		L.RaiseError("package.%s must be a string", pname)
	}
	messages := []string{}
	for _, pattern := range strings.Split(string(lpath), ";") {
		luapath := path.Clean(strings.Replace(pattern, "?", name, -1))
		if _, err := fs.Stat(l.fsys, luapath); err == nil {
			return luapath, ""
		} else if !os.IsNotExist(err) {
			messages = append(messages, err.Error())
		}
	}
	return "", strings.Join(messages, "\n\t")
}
