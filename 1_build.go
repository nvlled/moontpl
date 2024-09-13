package moontpl

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type Link string
type Filename string

type siteBuilder struct {
	testBuild   bool
	printOutput bool
	done        map[Link]bool
	buildQueue  []Link

	copyLuaSourceFiles bool
}

func newSiteBuilder() *siteBuilder {
	builder := &siteBuilder{
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
	m.SetPageData(L, pageData)

	return L
}

func (m *Moontpl) queueLink(link string) {
	if !m.builder.done[Link(link)] {
		m.builder.buildQueue = append(m.builder.buildQueue, Link(link))
	}
}

func (m *Moontpl) build(src, dest string) error {
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
	if !m.builder.testBuild {
		// TODO: write output to dest
		println("render", src, "->", dest)
	}
	if m.builder.printOutput {
		header := "---------[" + src + "]---------"
		println(header)
		println(output)
		println(strings.Repeat("-", len(header)))
		println()
	}

	return nil
}

func (m *Moontpl) BuildAll(outputDir string) error {
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

		if err := m.build(src, dest); err != nil {
			panic(err)
		}

		m.builder.done[linkWithParams] = true
	}

	if !m.builder.testBuild {
		if err := m.CopyNonSourceFiles(m.SiteDir, outputDir); err != nil {
			return err
		}
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
