package moontpl

import (
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
		L.SetField(mod, "hook", L.NewFunction(func(L *lua.LState) int {
			hookFn := L.ToFunction(1)
			L.SetGlobal("__HOOK_FUNC__", hookFn)
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

func (b *siteBuilder) Build(src, dest string, params PathParams) error {
	L := b.createState(src, params)
	defer L.Close()

	if err := L.DoFile(src); err != nil {
		return err
	}

	lv := L.Get(-1)
	if fn, ok := L.GetGlobal("__HOOK_FUNC__").(*lua.LFunction); ok && fn != nil {
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

	println("render", src, "->", dest, "\n", output)
	// TODO: write output to dest

	return nil
}

func (b *siteBuilder) CopyNonSourceFiles(srcDir, destDir string) error {
	// TODO:
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

		params, link := ExtractPathParams(string(linkWithParams))

		src := filepath.Join(b.srcDir, string(link)+".lua")
		dest := filepath.Join(b.destDir, string(link))

		if err := b.Build(src, dest, params); err != nil {
			panic(err)
		}

		b.done[linkWithParams] = true
	}

	return nil
}
