package moontpl

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func (m *Moontpl) queueLink(link string) {
	if !m.builder.done[Link(link)] {
		m.builder.buildQueue = append(m.builder.buildQueue, Link(link))
	}
}

func (m *Moontpl) build(src, dest string) error {
	L := m.createState(src)
	defer L.Close()

	output, err := m.RenderFile(src)
	if err != nil {
		return err
	}

	if !m.builder.testBuild {
		log.Print("exec ", mustRel(mustGetwd(), src), " -> ", mustRel(mustGetwd(), dest))

		// ignore error
		_ = os.MkdirAll(filepath.Dir(dest), 0755)

		if err := os.WriteFile(dest, []byte(output), 0644); err != nil {
			panic(err)
		}
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
	defer clear(m.builder.done)

	filenames, err := m.GetPageFilenames(m.SiteDir)
	if err != nil {
		return err
	}
	for _, p := range filenames {
		m.queueLink(p.Link)
	}

	for len(m.builder.buildQueue) > 0 {
		linkWithParams := m.builder.buildQueue[0]
		hashIndex := strings.Index(string(linkWithParams), "#")
		if hashIndex >= 0 {
			linkWithParams = linkWithParams[0:hashIndex]
		}

		m.builder.buildQueue = m.builder.buildQueue[1:]

		if _, ok := m.builder.done[linkWithParams]; ok {
			continue
		}

		src := filepath.Join(m.SiteDir, string(linkWithParams)+".lua")
		dest := filepath.Join(outputDir, string(linkWithParams))

		_, actualFilename := extractPathParams(src)
		if !fsExists(actualFilename) {
			log.Printf("LINK NOT FOUND: %s", linkWithParams)
		}

		if err := m.build(src, dest); err != nil {
			panic(err)
		}

		m.builder.done[linkWithParams] = true
	}

	plainFiles, err := m.getNonHtmlLuaFilenames(m.SiteDir)
	if err != nil {
		return err
	}
	for _, p := range plainFiles {
		src := filepath.Join(m.SiteDir, string(p.Link)+".lua")
		dest := filepath.Join(outputDir, string(p.Link))

		// Build only files with extension such as .css.lua,
		// skip .lua  files since they could be just regular modules.
		if filepath.Ext(dest) == "" {
			continue
		}

		if err := m.build(src, dest); err != nil {
			panic(err)
		}
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

		src := filepath.Join(srcDir, p)
		dest := filepath.Join(destDir, p)

		// ignore error
		_ = os.MkdirAll(filepath.Dir(dest), 0755)

		log.Print("copy ", mustRel(mustGetwd(), src), " -> ", mustRel(mustGetwd(), dest))

		inputFile, err := os.Open(src)
		if err != nil {
			return err
		}
		defer inputFile.Close()

		outputFile, err := os.Create(dest)
		if err != nil {
			return err
		}
		defer outputFile.Close()

		if _, err = io.Copy(outputFile, inputFile); err != nil {
			return err
		}

		return inputFile.Sync()
	})
}
