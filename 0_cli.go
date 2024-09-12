package moontpl

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/samber/lo"
)

//go:embed lua/*
var embedded embed.FS

// TODO: don't use global variables since it makes it harder to test
//       - use a separate lua state for GetPages
//       - remove Command constants
//       - do not use separate struct for site builder
//       - always load build module

type BuildCmd struct {
	SiteDir    string `arg:"required,positional" help:"directory that contains the source lua files"`
	OutputDir  string `arg:"positional" help:"directory where the rendered html files will be placed" default:"output"`
	CopyFiles  bool   `help:"copy and include static files (such as images) in the output" default:"true"`
	CopySource bool   `help:"copy and include source lua files in the output" default:"false"`
}

type RunCmd struct {
	Filename string `arg:"positional,required" help:"run a lua file and show ouput on the STDOUT"`
	SiteDir  string `arg:"-d" help:"for nested site directories, set to explicitly indicate where the site root is"`
}

func (*RunCmd) Epilogue() string {
	return `SITEDIR is automatically inferred based on the current working directory (CWD).
For example, in $ moontpl run ./site/foo/bar/page.html.lua
SITEDIR will be set to ./site since it's the closest directory from CWD to page.html.lua
This matters when require()'ing files found on the SITEDIR. The alternative is to set the LUADIR.
`
}

type ServeCmd struct {
	SiteDir string `arg:"required,positional" help:"directory that contains the source lua files to serve in a web server"`
	Port    int    `help:"HTTP port to use" default:"9876"`
}

type Args struct {
	Build *BuildCmd `arg:"subcommand:build"`
	Run   *RunCmd   `arg:"subcommand:run"`
	Serve *ServeCmd `arg:"subcommand:serve"`

	LuaDir []string `arg:"-l,separate" help:"directories where to find lua files with require(), automatically includes SITEDIR"`
}

var args Args

func ExecuteCLI() {
	p, err := arg.NewParser(arg.Config{}, &args)
	if err != nil {
		panic(err)
	}

	moon := New()

	err = p.Parse(os.Args[1:])
	switch {
	case err == arg.ErrHelp:
		{
			p.WriteHelp(os.Stdout)

			if args.Run != nil {
				println()
				println(args.Run.Epilogue())
			}
			os.Exit(0)
		}

	case err != nil:
		{
			fmt.Printf("error: %v\n", err)
			p.WriteUsage(os.Stdout)
			os.Exit(1)
		}
	}

	for _, p := range args.LuaDir {
		if !lo.Must(fsStat(p)).IsDir() {
			println("error: luadir must be a DIRECTORY")
			os.Exit(-1)
		}
		moon.AddLuaDir(mustAbs(p))
	}

	switch {
	case args.Run != nil:
		{
			moon.Command = CommandRun

			if args.Run.SiteDir != "" {
				moon.SiteDir = lo.Must(filepath.Abs(args.Run.SiteDir))
			} else {
				path := mustRel(mustGetwd(), mustAbs(args.Run.Filename))
				subDir, found := findFirstSubDirWithLuaFile(path)
				if found {
					moon.SiteDir = mustAbs(subDir)
				} else {
					moon.SiteDir = mustAbs(filepath.Dir(args.Run.Filename))
				}
			}

			moon.AddLuaPath(fmt.Sprintf("%s/?.lua", moon.SiteDir))
			output, err := moon.RenderFile(args.Run.Filename)

			if err != nil {
				panic(err)
			} else {
				println(output)
			}
		}

	case args.Build != nil:
		{
			moon.Command = CommandBuild
			moon.SiteDir = lo.Must(filepath.Abs(args.Build.SiteDir))
			outputDir := lo.Must(filepath.Abs(args.Build.OutputDir))
			moon.AddLuaDir(moon.SiteDir)

			if !isDirectory(moon.SiteDir) {
				println("error: SITEDIR must be a directory")
				os.Exit(1)
			}

			os.MkdirAll(outputDir, 0644)

			if !isDirectory(outputDir) {
				println("error: OUTPUTDIR must be a directory")
				os.Exit(1)
			}

			if moon.SiteDir == outputDir {
				println("error: SITEDIR and OUTPUTDIR must not be the same")
				os.Exit(1)
			}

			if isSubDirectory(moon.SiteDir, outputDir) || isSubDirectory(outputDir, moon.SiteDir) {
				println("error: SITEDIR and OUTPUTDIR must be subdirectories of each other")
				println("  SITEDIR:", moon.SiteDir)
				println("  OUTPUTDIR:", outputDir)
				os.Exit(1)
			}

			if err := moon.BuildAll(outputDir); err != nil {
				println("error:", err.Error())
			}
		}

	case args.Serve != nil:
		{
			moon.Command = CommandServe
			moon.SiteDir = lo.Must(filepath.Abs(args.Serve.SiteDir))
			moon.AddLuaDir(moon.SiteDir)
			moon.Serve("localhost:" + strconv.Itoa(args.Serve.Port))
		}
	}

}

func findFirstSubDirWithLuaFile(path string) (string, bool) {
	ps := ""

	for _, p := range strings.Split(path, string(os.PathSeparator)) {
		if p == "" {
			continue
		}

		ps = filepath.Join(ps, p)
		for range lo.Must(filepath.Glob(filepath.Join(ps, "*.lua"))) {

			return mustAbs(ps), true
		}
	}

	return "", false
}
