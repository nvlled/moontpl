package moontpl

import (
	"log"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

/*
# LUA_PATH is set to current directory by default
$ goom render index.html.lua
$ ls
pages index.html.lua includes
$ ls includes
lib1.lua lib2.lua

TODO:
$ goom run -path includes/ index.html.lua
$ goom run -path nvlled.github.io/pages/ nvlled.github.io/pages/index.html.lua
$ goom build [-path nvlled.github.io/pages] nvlled.github.io/pages
$ goom serve nvlled.github.io/pages

TODO:
moontpl.RenderString(filename, code)
moontpl.RenderFile(filename, data)
moontpl.Build(srcFile, destFile)
moontpl.BuildAll(srcDir, destDir)
moontpl.AddLuaPath(path)
moontpl.AddFs(fs fs.Fs)
moontpl.Serve(dir)
moontpl.ExecuteCLI()

moontpl.AddGlobal("ReadLogs", ReadLogs)
moontpl.AddGlobal("ReadImages", ReadImages)

	moontpl.AddModule("mymodule", ModMap{
		"Foo": func() {},
		"Bar": func() {},
	})

moontpl.ExecuteCLI()
*/
var (
	// if there's only one file, set to the directory containing the file
	SiteDir    = "./site"
	luaModules = map[string]ModMap{}
	luaGlobals = map[string]any{}
)

type ModMap map[string]any

type LuaModule interface {
	LMod()
}

func SetGlobal(varname string, obj any) {
	luaGlobals[varname] = obj
}
func SetModule(moduleName string, modMap ModMap) {
	luaModules[moduleName] = modMap
}

func createState(filename string) *lua.LState {
	L := lua.NewState()

	// TODO: remove, convert into a script that generates all page data
	L.SetGlobal("GetPageFilenames", luar.New(L, func() []PagePath {
		return GetPageFilenames(".")
	}))
	L.SetGlobal("GetPages", luar.New(L, func() []Page {
		pages, err := GetPages(L)
		if err != nil {
			L.RaiseError("failed to get pages: %v", err)
			log.Println(err)
			return nil
		}
		return pages
	}))
	// ----------------------------------------------------------------

	initAddedGlobals(L)
	initAddedModules(L)
	initEnvModule(L, filename)
	initPathModule(L, filename)

	return L
}

func initAddedGlobals(L *lua.LState) {
	for varname, v := range luaGlobals {
		L.SetGlobal(varname, luar.New(L, v))
	}
}

func initAddedModules(L *lua.LState) {
	for moduleName, modMap := range luaModules {
		L.PreloadModule(moduleName, func(L *lua.LState) int {
			mod := L.NewTable()

			for varname, val := range modMap {
				L.SetField(mod, varname, luar.New(L, val))
			}

			L.Push(mod)
			return 1
		})
	}
}

func initEnvModule(L *lua.LState, filename string) {
	L.PreloadModule("env", func(L *lua.LState) int {
		mod := L.NewTable()
		pagePath := getPagePath(filename)
		L.SetField(mod, "PAGE_FILENAME", lua.LString(pagePath.AbsFile))
		L.SetField(mod, "PAGE_LINK", lua.LString(pagePath.Link))
		L.Push(mod)
		return 1
	})
}

func initPathModule(L *lua.LState, filename string) {
	L.PreloadModule("path", func(L *lua.LState) int {
		mod := L.NewTable()
		L.SetField(mod, "getParams", luar.New(L, GetPathParams))
		L.SetField(mod, "setParams", luar.New(L, SetPathParams))
		L.SetField(mod, "hasParams", luar.New(L, HasPathParams))
		L.SetField(mod, "relative", luar.New(L, func(targetLink string) string {
			pagePath := getPagePath(filename)
			return RelativeFrom(targetLink, pagePath.Link)
		}))

		L.Push(mod)
		return 1
	})
}
