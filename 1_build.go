package moontpl

import (
	"io/fs"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

type Link string
type Filename string

type siteBuilder struct {
	done       map[Link]bool
	buildQueue []Link
	srcDir     string
	destDir    string

	copyLuaSourceFiles bool
}

func newBuilder(srcDir, destDir string) *siteBuilder {
	builder := &siteBuilder{
		done:       map[Link]bool{},
		buildQueue: []Link{},
		srcDir:     srcDir,
		destDir:    destDir,
	}
	return builder
}

func (b *siteBuilder) createState(filename string, params PathParams) *lua.LState {
	L := createState(filename)

	L.PreloadModule("page", func(L *lua.LState) int {
		mod := L.NewTable()

		lv := L.NewTable()
		for k, v := range params {
			lv.RawSetString(k, lua.LString(v))
		}
		L.SetField(mod, "params", lv)
		L.Push(mod)
		return 1
	})

	L.PreloadModule("build", func(L *lua.LState) int {
		mod := L.NewTable()
		L.SetField(mod, "queue", luar.New(L, b.queueLink))
		L.SetField(mod, "onPageRender", L.NewFunction(func(L *lua.LState) int {
			hookFn := L.ToFunction(1)
			globals := L.Get(lua.GlobalsIndex)
			L.SetTable(globals, lua.LNumber(hookIndex), hookFn)
			return 0
		}))

		L.Push(mod)
		return 1
	})

	return L
}

func (b *siteBuilder) queueLink(link string) {
	if !b.done[Link(link)] {
		b.buildQueue = append(b.buildQueue, Link(link))
	}
}

func (b *siteBuilder) Build(src, dest string) error {
	params, src := ExtractPathParams(src)

	L := b.createState(src, params)
	defer L.Close()

	if err := L.DoFile(src); err != nil {
		return err
	}

	lv := L.Get(-1)
	globals := L.Get(lua.GlobalsIndex)
	if fn, ok := L.GetTable(globals, lua.LNumber(hookIndex)).(*lua.LFunction); ok && fn != nil {
		err := L.CallByParam(lua.P{
			Fn:      fn,
			NRet:    1,
			Protect: false,
		}, lv)
		if err != nil {
			return err
		}
		if ret := L.Get(-1); ret != nil {
			lv = ret
		}
	}

	if lv.Type() == lua.LTNil {
		return nil
	}

	output := L.ToStringMeta(lv).String()
	println("----------------------------------------------------")
	println(output)
	println("----------------------------------------------------")
	// TODO: write output to dest

	return nil
}

func (b *siteBuilder) BuildAll() error {
	for _, p := range GetPageFilenames(b.srcDir) {
		b.queueLink(p.Link)
	}

	for len(b.buildQueue) > 0 {
		linkWithParams := b.buildQueue[0]
		b.buildQueue = b.buildQueue[1:]

		if _, ok := b.done[linkWithParams]; ok {
			continue
		}

		src := filepath.Join(b.srcDir, string(linkWithParams)+".lua")
		dest := filepath.Join(b.destDir, string(linkWithParams))

		println("render", src, "->", dest)
		if err := b.Build(src, dest); err != nil {
			panic(err)
		}

		b.done[linkWithParams] = true
	}

	if err := b.CopyNonSourceFiles(b.srcDir, b.destDir); err != nil {
		return err
	}

	return nil
}

func (b *siteBuilder) CopyNonSourceFiles(srcDir, destDir string) error {
	return fs.WalkDir(os.DirFS(srcDir), ".", func(p string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dir.IsDir() {
			return nil
		}
		if filepath.Ext(p) == ".lua" && !b.copyLuaSourceFiles {
			return nil
		}
		println("copy  ", filepath.Join(srcDir, p), "->", filepath.Join(destDir, p))
		return nil
	})
}
