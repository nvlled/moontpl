package moontpl

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/samber/lo"
)

//go:embed lua/*
var embedded embed.FS

type buildCmd struct {
	SiteDir    string `arg:"required,positional" help:"directory that contains the source lua files"`
	OutputDir  string `arg:"positional" help:"directory where the rendered html files will be placed" default:"output"`
	CopyFiles  bool   `help:"copy and include static files (such as images) in the output" default:"true"`
	CopySource bool   `help:"copy and include source lua files in the output" default:"false"`
	Test       bool   `help:"runs only the lua files, but do not write or copy files" default:"false"`
	Print      bool   `help:"prints the output of each file to STDOUT" default:"false"`
}

type runCmd struct {
	Filename string `arg:"positional,required" help:"run a lua file and show ouput on the STDOUT"`
	SiteDir  string `arg:"-d" help:"for nested site directories, set to explicitly indicate where the site root is"`
	Watch    bool   `arg:"-w" help:"watch file for changes" default:"false"`
	Force    bool   `arg:"-f" help:"try to run file even it doesn't have .lua file extension" default:"false"`
}

func (*runCmd) Epilogue() string {
	return `SITEDIR is automatically inferred based on the current working directory (CWD).
For example, in $ moontpl run ./site/foo/bar/page.html.lua
SITEDIR will be set to ./site since it's the closest directory from CWD to page.html.lua
that contains a .lua file. This matters when require()'ing files found on the SITEDIR.
The alternative is to set the LUADIR.
`
}

type serveCmd struct {
	SiteDir string `arg:"required,positional" help:"directory that contains the source lua files to serve in a web server"`
	Port    int    `help:"HTTP port to use" default:"9876"`
}

type luaDocCmd struct {
	Module string `arg:"positional" help:"show documentation for the module"`
}

type cliArgs struct {
	Build  *buildCmd  `arg:"subcommand:build"`
	Run    *runCmd    `arg:"subcommand:run"`
	Serve  *serveCmd  `arg:"subcommand:serve"`
	LuaDoc *luaDocCmd `arg:"subcommand:luadoc"`

	LuaDir []string `arg:"-l,separate" help:"directories where to find lua files with require(), automatically includes SITEDIR"`
	RunTag []string `arg:"-t,separate" help:"runtime tags to include in the lua environment"`

	Version bool `arg:"-v" help:"show version number"`
}

var args cliArgs

func showHelp(parser *arg.Parser) {
	parser.WriteHelp(os.Stdout)

	if args.Run != nil {
		println()
		println(args.Run.Epilogue())
	}
	os.Exit(0)
}

func ExecuteCLI() {
	p, err := arg.NewParser(arg.Config{}, &args)
	if err != nil {
		panic(err)
	}

	err = p.Parse(os.Args[1:])
	switch {
	case err == arg.ErrHelp:
		showHelp(p)

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
		moontpl.AddLuaDir(mustAbs(p))
	}

	moontpl.AddRunTags(args.RunTag...)

	switch {
	default:
		showHelp(p)

	case args.Version:
		if Version != "" {
			fmt.Printf("version: %s\n", Version)
		} else {
			fmt.Printf("version: development\n")
		}

	case args.LuaDoc != nil:
		if args.LuaDoc.Module != "" {
			module := args.LuaDoc.Module
			filename := filepath.Join("lua", module+".lua")

			doc, ok, err := moontpl.extractDocumentation(filename)
			if err != nil {
				println("failed to get documentation:", err.Error())
				os.Exit(-1)
			}
			if ok {
				fmt.Print(doc)
			} else {
				println("no documentation found for", module)
			}
		} else {
			entries, err := fs.Glob(moontpl.fsys, "lua/*.lua")
			if err != nil {
				println("failed to get read dir:", err.Error())
				os.Exit(-1)
			}
			for _, filename := range entries {
				doc, ok, err := moontpl.extractDocumentation(filename)
				if err != nil {
					println("failed to get documentation:", err.Error())
					os.Exit(-1)
				}
				if ok {
					fmt.Println(doc)
					fmt.Print("\n\n")
				}
			}
		}

	case args.Run != nil:
		{
			moontpl.Command = CommandRun

			if args.Run.SiteDir != "" {
				moontpl.SiteDir = lo.Must(filepath.Abs(args.Run.SiteDir))
			} else {
				path := mustRel(mustGetwd(), mustAbs(args.Run.Filename))
				subDir, found := findFirstSubDirWithLuaFile(path)
				if found {
					moontpl.SiteDir = mustAbs(subDir)
				} else {
					moontpl.SiteDir = mustAbs(filepath.Dir(args.Run.Filename))
				}
			}

			if filepath.Ext(args.Run.Filename) != ".lua" && !args.Run.Force {
				println("file must have a .lua file extension")
				os.Exit(-1)
			}

			run := func() {
				output, err := moontpl.RenderFile(args.Run.Filename)

				if err != nil {
					log.Println(err)
				} else {
					fmt.Println(output)
				}
			}

			moontpl.AddLuaPath(fmt.Sprintf("%s/?.lua", moontpl.SiteDir))
			moontpl.AddRunTags("run")

			if args.Run.Watch {
				moontpl.fsWatcher.On(func(filename string) {
					var modname string
					if !isSubDirectory(moontpl.SiteDir, filename) {
						filename = path.Base(filename)
						modname = getModuleName(moontpl.SiteDir, filename)
					} else {
						modname = getModuleName(moontpl.SiteDir, filename)
					}
					moontpl.luaPool.resetLoadedPoolModules(modname)

					now := time.Now().Local().Format("15:04:05")
					fmt.Printf(" --------------------[ start output %s ]--------------------\n", now)
					run()
					fmt.Printf(" --------------------[ end output   %s ]--------------------\n", now)
				})

				_ = moontpl.startFsWatch()
				run()
				fmt.Printf(" --------------------[ output %s ]--------------------\n", time.Now().Local().Format("15:04:05"))

				<-make(chan struct{})
			} else {
				moontpl.disableLuaPool = true
				run()
			}

		}

	case args.Build != nil:
		{
			moontpl.Command = CommandBuild
			moontpl.SiteDir = lo.Must(filepath.Abs(args.Build.SiteDir))
			moontpl.AddLuaDir(moontpl.SiteDir)
			moontpl.AddRunTags("build")

			outputDir := lo.Must(filepath.Abs(args.Build.OutputDir))

			if !args.Build.Test {
				if !isDirectory(moontpl.SiteDir) {
					println("error: SITEDIR must be a directory")
					os.Exit(1)
				}

				_ = os.MkdirAll(outputDir, 0755)

				if !isDirectory(outputDir) {
					println("error: OUTPUTDIR must be a directory")
					os.Exit(1)
				}

				if moontpl.SiteDir == outputDir {
					println("error: SITEDIR and OUTPUTDIR must not be the same")
					os.Exit(1)
				}

				if isSubDirectory(moontpl.SiteDir, outputDir) || isSubDirectory(outputDir, moontpl.SiteDir) {
					println("error: SITEDIR and OUTPUTDIR must be subdirectories of each other")
					println("  SITEDIR:", moontpl.SiteDir)
					println("  OUTPUTDIR:", outputDir)
					os.Exit(1)
				}
			}

			moontpl.builder.testBuild = args.Build.Test
			moontpl.builder.printOutput = args.Build.Print

			if err := moontpl.BuildAll(outputDir); err != nil {
				println("error:", err.Error())
			}
		}

	case args.Serve != nil:
		{
			moontpl.Command = CommandServe
			moontpl.SiteDir = lo.Must(filepath.Abs(args.Serve.SiteDir))
			moontpl.AddLuaDir(moontpl.SiteDir)
			moontpl.AddRunTags("serve")

			moontpl.fsWatcher.On(func(filename string) {
				var modname string
				if !isSubDirectory(moontpl.SiteDir, filename) {
					filename = path.Base(filename)
					modname = getModuleName(moontpl.SiteDir, filename)
				} else {
					modname = getModuleName(moontpl.SiteDir, filename)
				}

				moontpl.luaPool.resetLoadedPoolModules(modname)

			})

			if !isDirectory(moontpl.SiteDir) {
				println("error: SITEDIR must be a directory")
				os.Exit(1)
			}

			_ = moontpl.startFsWatch()
			moontpl.Serve("localhost:" + strconv.Itoa(args.Serve.Port))
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
