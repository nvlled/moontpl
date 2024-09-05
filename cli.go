package moontpl

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/samber/lo"
	lua "github.com/yuin/gopher-lua"
)

func ExecuteCLI() {
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
		{
			for _, filename := range args {
				output, err := RenderFile(filename)
				if err != nil {
					panic(err)
				}
				println(output)
			}
		}

	case "serve":
		Serve("localhost:8080")

	case "build":
		{
			outputDir := "output"
			if len(args) > 0 {
				outputDir = args[0]
			}
			builder := newBuilder(SiteDir, outputDir)
			builder.BuildAll()
		}

	default:
		println("invalid command: " + cmd)
	}
}
