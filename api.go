package moontpl

import (
	"io/fs"

	lua "github.com/yuin/gopher-lua"
)

var moontpl *Moontpl

func RenderFile(filename string) (string, error) {
	return moontpl.RenderFile(filename)
}

func RenderString(code string) (string, error) {
	return moontpl.RenderString(code)
}

func BuildAll(outputDir string) error {
	return moontpl.BuildAll(outputDir)
}

func CopyNonSourceFiles(srcDir, destDir string) error {
	return moontpl.CopyNonSourceFiles(srcDir, destDir)
}

func Serve(addr string) {
	moontpl.Serve(addr)
}

func SetGlobal(varname string, obj any) {
	moontpl.SetGlobal(varname, obj)
}

func SetModule(moduleName string, modMap ModMap) {
	moontpl.SetModule(moduleName, modMap)
}

func AddLuaPath(pathStr string) {
	moontpl.AddLuaPath(pathStr)
}

func AddLuaDir(dir string) {
	moontpl.AddLuaDir(dir)
}

func GetPages(L *lua.LState) ([]Page, error) {
	// TODO: remove L arg
	return moontpl.GetPages(L)
}

func GetPageFilenames(baseDir string) []PagePath {
	return moontpl.GetPageFilenames(baseDir)
}

func AddFs(fsys fs.FS) {
	moontpl.AddFs(fsys)
}
