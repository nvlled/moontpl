package moontpl

import (
	"embed"
	"fmt"
	"log"
	"os"
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
	Watch    bool   `arg:"-w" help:"watch file for changes"`
}

func (*runCmd) Epilogue() string {
	return `SITEDIR is automatically inferred based on the current working directory (CWD).
For example, in $ moontpl run ./site/foo/bar/page.html.lua
SITEDIR will be set to ./site since it's the closest directory from CWD to page.html.lua
This matters when require()'ing files found on the SITEDIR. The alternative is to set the LUADIR.
`
}

type serveCmd struct {
	SiteDir string `arg:"required,positional" help:"directory that contains the source lua files to serve in a web server"`
	Port    int    `help:"HTTP port to use" default:"9876"`
}

type cliArgs struct {
	Build *buildCmd `arg:"subcommand:build"`
	Run   *runCmd   `arg:"subcommand:run"`
	Serve *serveCmd `arg:"subcommand:serve"`

	LuaDir []string `arg:"-l,separate" help:"directories where to find lua files with require(), automatically includes SITEDIR"`
	RunTag []string `arg:"-l,separate" help:"runtime tags to include in the lua environment"`
}

var args cliArgs

func ExecuteCLI() {
	p, err := arg.NewParser(arg.Config{}, &args)
	if err != nil {
		panic(err)
	}

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
		moontpl.AddLuaDir(mustAbs(p))
	}

	switch {
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

			moontpl.AddLuaPath(fmt.Sprintf("%s/?.lua", moontpl.SiteDir))
			moontpl.AddRunTags("run")

			run := func() {
				output, err := moontpl.RenderFile(args.Run.Filename)

				if err != nil {
					log.Println(err)
				} else {
					println(output)
					fmt.Printf(" --------------------[ output %s ]--------------------", time.Now().Local().Format("15:04:05"))
				}
			}

			if args.Run.Watch {
				moontpl.fsWatcher.On(func(string) {
					run()
				})
			}

			go moontpl.startFsWatch()
			run()
			<-make(chan struct{})
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

				os.MkdirAll(outputDir, 0755)

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
			moontpl.AddRunTags("serve", "autoreload")

			if !isDirectory(moontpl.SiteDir) {
				println("error: SITEDIR must be a directory")
				os.Exit(1)
			}

			go moontpl.startFsWatch()
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
