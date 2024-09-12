package moontpl

import (
	"io/fs"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

type Link string
type Filename string

type siteBuilder struct {
	running    bool
	done       map[Link]bool
	buildQueue []Link

	copyLuaSourceFiles bool
}

func newSiteBuilder() *siteBuilder {
	builder := &siteBuilder{
		running:    false,
		done:       map[Link]bool{},
		buildQueue: []Link{},
	}
	return builder
}

func (m *Moontpl) createBuildState(filename string, params pathParams) *lua.LState {
	pageData := PageData{}
	for k, v := range params {
		pageData[k] = v
	}

	L := m.createState(filename)
	m.initPageModule(L, pageData)

	return L
}

func (m *Moontpl) queueLink(link string) {
	if !m.builder.done[Link(link)] {
		m.builder.buildQueue = append(m.builder.buildQueue, Link(link))
	}
}

func (m *Moontpl) Build(src, dest string) error {
	params, src := extractPathParams(src)

	L := m.createBuildState(src, params)
	defer L.Close()

	lv, err := m.renderFile(L, src)
	if err != nil {
		return err
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

func (m *Moontpl) BuildAll(outputDir string) error {
	m.builder.running = true
	defer func() { m.builder.running = false }()

	for _, p := range m.GetPageFilenames(m.SiteDir) {
		m.queueLink(p.Link)
	}

	for len(m.builder.buildQueue) > 0 {
		linkWithParams := m.builder.buildQueue[0]
		m.builder.buildQueue = m.builder.buildQueue[1:]

		if _, ok := m.builder.done[linkWithParams]; ok {
			continue
		}

		src := filepath.Join(m.SiteDir, string(linkWithParams)+".lua")
		dest := filepath.Join(outputDir, string(linkWithParams))

		println("render", src, "->", dest)
		if err := m.Build(src, dest); err != nil {
			panic(err)
		}

		m.builder.done[linkWithParams] = true
	}

	if err := m.CopyNonSourceFiles(m.SiteDir, outputDir); err != nil {
		return err
	}

	return nil
}

func (m *Moontpl) CopyNonSourceFiles(srcDir, destDir string) error {
	return fs.WalkDir(os.DirFS(srcDir), ".", func(p string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dir.IsDir() {
			return nil
		}
		if filepath.Ext(p) == ".lua" && !m.builder.copyLuaSourceFiles {
			return nil
		}
		println("copy  ", filepath.Join(srcDir, p), "->", filepath.Join(destDir, p))
		return nil
	})
}
