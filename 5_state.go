package moontpl

import (
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/laher/mergefs"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

// TODO: don't use global variables since it makes it harder to test
//       - use a separate lua state for GetPages

const (
	registryPrefix = "moontpl."
)

const (
	CommandNone = iota
	CommandRun
	CommandServe
	CommandBuild
)

var (
	SiteDir = ""
	Command = CommandNone

	luaModules  = map[string]ModMap{}
	luaGlobals  = map[string]any{}
	fileSystems = []fs.FS{}

	cachedPages []Page
)

type ModMap map[string]any

type LuaModule interface {
	LMod()
}

func AddFs(fsys fs.FS) {
	fileSystems = append(fileSystems, fsys)
}

func SetGlobal(varname string, obj any) {
	luaGlobals[varname] = obj
}
func SetModule(moduleName string, modMap ModMap) {
	luaModules[moduleName] = modMap
}

func AddLuaPath(pathStr string) {
	var sep string
	var path = strings.TrimSpace(lua.LuaPathDefault)

	if len(path) > 0 {
		if path[len(path)-1] == ';' {
			sep = ""
		} else {
			sep = ";"
		}
	}

	lua.LuaPathDefault = path + sep + pathStr
}

func AddLuaDir(dir string) {
	AddLuaPath(path.Join(dir, "?.lua"))
}

func createState(filename string) *lua.LState {
	L := lua.NewState(lua.Options{
		SkipOpenLibs: true,
	})

	openLibs(L)

	initAddedGlobals(L)
	initAddedModules(L)
	initEnvModule(L, filename)
	initPathModule(L, filename)

	fsys := mergefs.Merge(fileSystems...)
	initFsLoader(L, fsys)

	return L
}

func openLibs(L *lua.LState) {
	var luaLibs = []struct {
		libName string
		libFunc lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage},
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},

		// Disable these libraries:
		//{lua.DebugLibName, lua.OpenDebug},
		//{lua.IoLibName, lua.OpenIo},
		//{lua.OsLibName, lua.OpenOs},
		//{lua.ChannelLibName, lua.OpenChannel},
		//{lua.CoroutineLibName, lua.OpenCoroutine},
	}
	for _, lib := range luaLibs {
		L.Push(L.NewFunction(lib.libFunc))
		L.Push(lua.LString(lib.libName))
		L.Call(1, 0)
	}

	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		if GetInternalVar(L, "recursed") != lua.LNil {
			return 0
		}

		top := L.GetTop()
		for i := 1; i <= top; i++ {
			fmt.Print(L.ToStringMeta(L.Get(i)).String())
			if i != top {
				fmt.Print("\t")
			}
		}
		fmt.Println("")
		return 0
	}))

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

		L.SetField(mod, "getPageFilenames", L.NewFunction(func(L *lua.LState) int {
			paths := GetPageFilenames(SiteDir)
			var filenames []string
			for _, p := range paths {
				filenames = append(filenames, p.Link)
			}

			L.Push(luarFromArray(L, filenames))
			return 1
		}))

		L.SetField(mod, "getPages", L.NewFunction(func(L *lua.LState) int {
			if GetInternalVar(L, "recursed") != lua.LNil {
				L.Push(L.NewTable())
				return 1
			}
			SetInternalVar(L, "recursed", lua.LTrue)
			defer SetInternalVar(L, "recursed", lua.LNil)

			if Command == CommandBuild {
				if cachedPages == nil {
					pages, err := GetPages(L)
					if err != nil {
						panic(err)
					}
					cachedPages = pages
				}

				L.Push(luarFromArray(L, cachedPages))
				return 1
			} else {
				ck := "GetPages"
				if lv := getStateCache(L, ck); lv != lua.LNil {
					L.Push(lv)
					return 1
				}

				pages, err := GetPages(L)
				if err != nil {
					panic(err)
				}

				lv := luarFromArray(L, pages)
				setStateCache(L, ck, lv)

				L.Push(lv)
				return 1
			}
		}))

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

func getStateCache(L *lua.LState, key string) lua.LValue {
	return GetInternalVar(L, "cache."+key)
}

func setStateCache(L *lua.LState, key string, value lua.LValue) {
	SetInternalVar(L, "cache."+key, value)
}

func luarFromArray[T any](L *lua.LState, items []T) lua.LValue {
	t := L.NewTable()
	for _, x := range items {
		t.Append(luar.New(L, x))
	}
	return t
}

func SetInternalVar(L *lua.LState, key string, val lua.LValue) {
	L.G.Registry.RawSetString(registryPrefix+key, val)
}

func GetInternalVar(L *lua.LState, key string) lua.LValue {
	return L.G.Registry.RawGetString(registryPrefix + key)
}
