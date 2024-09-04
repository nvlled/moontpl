package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/samber/lo"
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
*/
var (
	// if there's only one file, set to the directory containing the file
	SiteDir = "./site"
)

func init() {
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("failed to get working directory with os.Getwd: %v", err)
		return
	}
	SiteDir = filepath.Join(cwd, "site")

	pathPtr := flag.String("luapath", SiteDir, "add directory to LUA_PATH")
	flag.Parse()

	luaPath := os.Getenv("LUA_PATH")
	lua.LuaPathDefault = ";./lua/?.lua;" + luaPath

	if *pathPtr != "" {
		if !filepath.IsAbs(*pathPtr) {
			*pathPtr = lo.Must(filepath.Abs(*pathPtr))
		}
		SiteDir = *pathPtr
		// LuaPathDefault is set to something like:
		//   ./?.lua;/usr/local/share/lua/5.1/?.lua;/usr/local/share/lua/5.1/?/init.lua
		// If LUA_PATH env is set, LuaPathDefault is ignored.
		// Syntax of LUA_PATH is similiar to LuaPathDefault shown above.
		//os.Setenv("LUA_PATH", fmt.Sprintf("./lua/?.lua;%s/?.lua", BaseDir))
		lua.LuaPathDefault = fmt.Sprintf("./lua/?.lua;%s/?.lua;%s", SiteDir, luaPath)
	}

	args := flag.Args()
	if len(args) < 1 {
		println("command required")
		return
	}

	cmd := args[0]
	args = args[1:]

	switch cmd {
	case "run":
		for _, filename := range args {
			output, err := RenderFile(filename)
			if err != nil {
				panic(err)
			}
			println(output)
		}
	case "serve":
		{
			server := &http.Server{
				Addr:    "localhost:8080",
				Handler: createHTTPHandler(),
			}

			if err := server.ListenAndServe(); err != nil {
				panic(err)
			}

		}

	case "build":
		{
			builder := NewBuilder(SiteDir, "output")
			builder.BuildAll()
		}

	default:
		println("invalid command: " + cmd)
	}
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

	initEnvModule(L, filename)
	initPathModule(L, filename)

	return L
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

func renderFile(L *lua.LState, filename string) (string, error) {
	if err := L.DoFile(filename); err != nil {
		return "", err
	}

	lv := L.Get(-1)
	if lv.Type() == lua.LTNil {
		return "", nil
	}

	return L.ToStringMeta(lv).String(), nil
}

func RenderFile(filename string) (string, error) {
	L := createState(filename)
	defer L.Close()
	return renderFile(L, filename)
}

func RenderString(luaCode string) (string, error) {
	L := createState("inline.lua")
	defer L.Close()
	if err := L.DoString(luaCode); err != nil {
		return "", err
	}

	lv := L.Get(-1)
	if lv.Type() == lua.LTNil {
		return "", nil
	}
	return L.ToStringMeta(lv).String(), nil
}
