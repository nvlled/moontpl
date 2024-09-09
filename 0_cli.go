package moontpl

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/samber/lo"
)

//go:embed lua/*
var embedded embed.FS

func ExecuteCLI() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("failed to get working directory with os.Getwd: %v", err)
		return
	}
	SiteDir = filepath.Join(cwd, "site")

	pathPtr := flag.String("luadir", SiteDir, "add directory to LUA_PATH")
	flag.Parse()

	AddFs(embedded)
	AddLuaPath("./lua/?.lua")
	AddLuaPath(fmt.Sprintf("./%s/?.lua", SiteDir))

	luaPath := os.Getenv("LUA_PATH")
	if luaPath != "" {
		AddLuaPath(luaPath)
	}

	if *pathPtr != "" {
		if !filepath.IsAbs(*pathPtr) {
			*pathPtr = lo.Must(filepath.Abs(*pathPtr))
		}
		AddLuaPath(filepath.Join(*pathPtr, "?.lua"))
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
		if len(args) == 1 {
			SiteDir = lo.Must(filepath.Abs(args[0]))
			AddLuaDir(SiteDir)
		} else {
			println("usage: serve [site-dir]")
		}
		Serve("localhost:8080")

	case "build":
		{
			if len(args) > 0 {
				SiteDir = lo.Must(filepath.Abs(args[0]))
				AddLuaDir(SiteDir)
			}

			outputDir := "output"
			if len(args) > 1 {
				outputDir = args[1]
			}
			outputDir = lo.Must(filepath.Abs(outputDir))

			builder := newBuilder(SiteDir, outputDir)
			builder.BuildAll()
		}

	default:
		println("invalid command: " + cmd)
	}
}
