package moontpl

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	lua "github.com/yuin/gopher-lua"
)

type PagePath struct {
	AbsFile string
	RelFile string
	Link    string
}

type Page struct {
	PagePath
	Data *lua.LTable
}

func (m *Moontpl) GetPageFilenames(baseDir string) []PagePath {
	var result []PagePath
	filepath.WalkDir(baseDir, func(filename string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(filename, ".html.lua") {
			result = append(result, m.getPagePath(filename))
		}
		return nil
	})
	return result
}

func (m *Moontpl) GetPages(L *lua.LState) ([]Page, error) {
	result := []Page{}
	for _, entry := range m.GetPageFilenames(m.SiteDir) {
		data, err := getReturnedPageData(L, entry.AbsFile)
		if err != nil {
			return nil, err
		}
		result = append(result, Page{
			PagePath: entry,
			Data:     data,
		})
	}

	return result, nil
}

func (m *Moontpl) getPagePath(filename string) PagePath {
	filename = lo.Must(filepath.Abs(filename))

	link := mustRel(m.SiteDir, filename)
	link = "/" + strings.TrimSuffix(link, ".lua")

	return PagePath{
		AbsFile: filename,
		RelFile: mustRel(m.SiteDir, filename),
		Link:    link,
	}
}

func getReturnedPageData(L *lua.LState, filename string) (*lua.LTable, error) {
	if err := L.DoFile(filename); err != nil {
		return L.NewTable(), err
	}
	lv := L.Get(-1)

	if lv.Type() != lua.LTTable {
		return L.NewTable(), nil
	}

	table := lv.(*lua.LTable)

	if data, ok := table.RawGet(lua.LString("data")).(*lua.LTable); ok {
		return data, nil
	}

	return L.NewTable(), nil
}
