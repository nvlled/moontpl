package moontpl

import (
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/laher/mergefs"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

const (
	registryPrefix = "moontpl."
)

const (
	CommandNone = iota
	CommandRun
	CommandServe
	CommandBuild
)

type Moontpl struct {
	SiteDir     string
	Command     int
	luaModules  map[string]ModMap
	luaGlobals  map[string]any
	fileSystems []fs.FS
	cachedPages []Page
	runtags     map[string]struct{}

	builder   *siteBuilder
	fsWatcher *FsWatcher
}

type PageData map[string]any

func New() *Moontpl {
	self := &Moontpl{
		SiteDir: "",
		Command: CommandNone,

		luaModules:  map[string]ModMap{},
		luaGlobals:  map[string]any{},
		fileSystems: []fs.FS{},
		cachedPages: []Page{},
		runtags:     make(map[string]struct{}),

		builder:   newSiteBuilder(),
		fsWatcher: newFsWatcher(),
	}

	self.AddFs(embedded)
	self.AddLuaPath("./lua/?.lua")

	return self
}

type ModMap map[string]any

type LuaModule interface {
	LMod()
}

func (m *Moontpl) AddFs(fsys fs.FS) {
	m.fileSystems = append(m.fileSystems, fsys)
}

func (m *Moontpl) SetGlobal(varname string, obj any) {
	m.luaGlobals[varname] = obj
}
func (m *Moontpl) SetModule(moduleName string, modMap ModMap) {
	m.luaModules[moduleName] = modMap
}

func (m *Moontpl) AddLuaPath(pathStr string) {
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

func (m *Moontpl) AddLuaDir(dir string) {
	m.AddLuaPath(path.Join(dir, "?.lua"))
}

func (m *Moontpl) SetPageData(L *lua.LState, pageData PageData) {
	// make sure page module is loaded
	L.DoString(`require "page"`)

	mod := getLoadedModule(L, "page")
	if mod != lua.LNil {
		t := pageDataToLValue(L, pageData)
		L.SetField(mod, "input", t)
	}
}

func (m *Moontpl) AddRunTags(tags ...string) {
	for _, t := range tags {
		m.runtags[t] = struct{}{}
	}
}

func (m *Moontpl) RemoveRunTags(tags ...string) {
	for _, t := range tags {
		delete(m.runtags, t)
	}
}

func (m *Moontpl) createState(filename string, initModules ...bool /* = true */) *lua.LState {
	L := lua.NewState(lua.Options{
		SkipOpenLibs: true,
	})

	openLibs(L)

	if len(initModules) == 0 || initModules[0] {
		m.initAddedGlobals(L)
		m.initAddedModules(L)
		m.initPageModule(L, filename)
		m.initPathModule(L, filename)
		m.initHookModule(L)
		m.initBuildModule(L)
		m.initTagsModule(L)
		m.initSiteModule(L)
	}

	// allow loading lua modules from fs.Fs (mainly for embedded files)
	fsys := mergefs.Merge(m.fileSystems...)
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
}

func (m *Moontpl) initAddedGlobals(L *lua.LState) {
	for varname, v := range m.luaGlobals {
		L.SetGlobal(varname, luar.New(L, v))
	}
}

func (m *Moontpl) initAddedModules(L *lua.LState) {
	for moduleName, modMap := range m.luaModules {
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

func (m *Moontpl) initPathModule(L *lua.LState, filename string) {
	L.PreloadModule("path", func(L *lua.LState) int {
		mod := L.NewTable()
		L.SetField(mod, "getParams", luar.New(L, getPathParams))
		L.SetField(mod, "setParams", luar.New(L, setPathParams))
		L.SetField(mod, "hasParams", luar.New(L, hasPathParams))

		L.SetField(mod, "absolute", L.NewFunction(func(L *lua.LState) int {
			targetLink := L.CheckString(1)
			pagePath := m.getPagePath(filename)
			L.Push(lua.LString(relativeFrom(targetLink, pagePath.Link)))
			return 1
		}))

		L.SetField(mod, "absolute", L.NewFunction(func(L *lua.LState) int {
			link := L.CheckString(1)
			if link[0] == '/' {
				L.Push(lua.LString(link))
				return 1
			}
			pagePath := m.getPagePath(filename)
			result := path.Dir(pagePath.Link)
			result = path.Join(result, link)
			result = path.Clean(result)

			L.Push(lua.LString(result))

			return 1
		}))

		L.Push(mod)
		return 1
	})
}

func (m *Moontpl) initPageModule(L *lua.LState, filename string) {
	L.PreloadModule("page", func(L *lua.LState) int {
		mod := L.NewTable()
		pagePath := m.getPagePath(filename)

		L.SetField(mod, "input", L.NewTable())
		L.SetField(mod, "data", L.NewTable())
		L.SetField(mod, "PAGE_FILENAME", lua.LString(pagePath.AbsFile))
		L.SetField(mod, "PAGE_LINK", lua.LString(pagePath.Link))

		L.SetField(mod, "files", L.NewFunction(func(L *lua.LState) int {
			paths := m.GetPageFilenames(m.SiteDir)
			var filenames []string
			for _, p := range paths {
				filenames = append(filenames, p.Link)
			}

			L.Push(luarFromArray(L, filenames))
			return 1
		}))

		L.SetField(mod, "list", L.NewFunction(func(L *lua.LState) int {
			pages, err := m.GetPages()
			if err != nil {
				panic(err)
			}

			lv := luarFromArray(L, pages)
			L.Push(lv)

			return 1
		}))

		L.Push(mod)
		return 1
	})
}

func (m *Moontpl) initHookModule(L *lua.LState) {
	L.PreloadModule("hook", func(L *lua.LState) int {
		mod := L.NewTable()
		L.SetField(mod, "onPageRender", L.NewFunction(func(L *lua.LState) int {
			return 0
		}))

		L.Push(mod)
		return 1
	})
}

func (m *Moontpl) initBuildModule(L *lua.LState) {
	L.PreloadModule("build", func(L *lua.LState) int {
		mod := L.NewTable()
		L.SetField(mod, "queue", luar.New(L, m.queueLink))
		L.Push(mod)
		return 1
	})
}

func (m *Moontpl) initTagsModule(L *lua.LState) {
	L.PreloadModule("runtags", func(L *lua.LState) int {
		mod := L.NewTable()
		for k := range m.runtags {
			L.SetField(mod, k, lua.LTrue)
		}

		L.Push(mod)
		return 1
	})
}

func (m *Moontpl) initSiteModule(L *lua.LState) {
	L.PreloadModule("site", func(L *lua.LState) int {
		mod := L.NewTable()
		L.SetField(mod, "files", L.NewFunction(func(L *lua.LState) int {
			options := L.OptTable(1, L.NewTable())

			dir := lua.LString("/")
			var filter lua.LValue = lua.LNil
			var includeLua lua.LValue = lua.LFalse

			if options != lua.LNil {
				if s, ok := options.RawGetString("dir").(lua.LString); ok {
					dir = s
				}
				if f, ok := options.RawGetString("filter").(*lua.LFunction); ok {
					filter = f
				}
				if val := options.RawGetString("lua"); val != lua.LNil {
					includeLua = val
				}
			}

			result := L.NewTable()
			err := fs.WalkDir(os.DirFS(m.SiteDir), ".", func(p string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.IsDir() {
					return nil
				}

				p = "/" + p
				if !isSubDirectory(string(dir), p) {
					return nil
				}

				if filter != lua.LNil {
					err = L.CallByParam(lua.P{
						Fn:      filter,
						NRet:    1,
						Protect: true,
					}, lua.LString(p))

					if err != nil {
						panic(err)
					}

					if ret := L.Get(-1); ret == lua.LNil || ret == lua.LFalse {
						return nil
					}
				}

				if strings.HasSuffix(p, ".lua") {
					if wholeExt(p) == ".lua" {
						if includeLua == lua.LFalse {
							return nil // skip plain .lua files
						}
					} else {
						p = p[0 : len(p)-4] // remove .lua
					}
				}

				result.Append(lua.LString(p))

				return nil
			})

			if err != nil {
				panic(err)
			}

			L.Push(result)
			return 1
		}))

		L.Push(mod)
		return 1
	})
}

func mapToLtable[T any](L *lua.LState, items map[string]T) lua.LValue {
	t := L.NewTable()
	for k, v := range items {
		var lv lua.LValue
		var x any = v
		switch v := x.(type) {
		case float32:
			lv = lua.LNumber(v)
		case float64:
			lv = lua.LNumber(v)
		case int:
			lv = lua.LNumber(v)
		case int8:
			lv = lua.LNumber(v)
		case int16:
			lv = lua.LNumber(v)
		case int32:
			lv = lua.LNumber(v)
		case int64:
			lv = lua.LNumber(v)

		case string:
			lv = lua.LString(v)
		case bool:
			lv = lua.LBool(v)
		default:
			lv = luar.New(L, v)
		}
		t.RawSetString(k, lv)
	}
	return t
}

func arrayToLTable[T any](L *lua.LState, items []T) *lua.LTable {
	t := L.NewTable()
	for _, v := range items {
		var lv lua.LValue
		var x any = v
		switch v := x.(type) {
		case float32:
			lv = lua.LNumber(v)
		case float64:
			lv = lua.LNumber(v)
		case int:
			lv = lua.LNumber(v)
		case int8:
			lv = lua.LNumber(v)
		case int16:
			lv = lua.LNumber(v)
		case int32:
			lv = lua.LNumber(v)
		case int64:
			lv = lua.LNumber(v)
		case string:
			lv = lua.LString(v)
		case bool:
			lv = lua.LBool(v)
		default:
			lv = luar.New(L, v)
		}
		t.Append(lv)
	}
	return t
}

func pageDataToLValue(L *lua.LState, data PageData) lua.LValue {
	t := L.NewTable()
	for k, v := range data {
		var lv lua.LValue
		switch v := v.(type) {
		case float32:
			lv = lua.LNumber(v)
		case float64:
			lv = lua.LNumber(v)
		case int:
			lv = lua.LNumber(v)
		case int8:
			lv = lua.LNumber(v)
		case int16:
			lv = lua.LNumber(v)
		case int32:
			lv = lua.LNumber(v)
		case int64:
			lv = lua.LNumber(v)

		case string:
			lv = lua.LString(v)
		case bool:
			lv = lua.LBool(v)
		default:
			lv = luar.New(L, v)
		}
		t.RawSetString(k, lv)
	}
	return t
}

func getLoadedModule(L *lua.LState, moduleName string) lua.LValue {
	lv := L.GetField(L.GetField(L.Get(lua.EnvironIndex), "package"), "loaded")
	if loaded, ok := lv.(*lua.LTable); !ok {
		L.RaiseError("package.loaded must be a table")
		return lua.LNil
	} else {
		return loaded.RawGetString(moduleName)
	}
}
