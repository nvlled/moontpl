package moontpl

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/samber/lo"
	lua "github.com/yuin/gopher-lua"
)

type PathParams map[string]string
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

func (m *Moontpl) GetPages(L *lua.LState) ([]Page, error) {
	result := []Page{}
	for _, entry := range m.GetPageFilenames(m.SiteDir) {
		data, err := GetPageData(L, entry.AbsFile)
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

func GetPageData(L *lua.LState, filename string) (*lua.LTable, error) {
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

func GetPathParams(filename string) PathParams {
	re := regexp.MustCompile(`\[.*?\]`)
	matched := re.FindString(filename)
	result := map[string]string{}
	if len(matched) == 0 {
		return result
	}
	for _, s := range strings.Split(matched[1:len(matched)-1], ",") {
		i := strings.Index(s, "=")
		if i < 0 {
			continue
		}
		key := strings.TrimSpace(s[0:i])
		value := strings.TrimSpace(s[i+1:])
		result[key] = value
	}
	return result
}

func ExtractPathParams(filename string) (PathParams, string) {
	re := regexp.MustCompile(`\[.*?\]`)
	matched := re.FindString(filename)
	result := map[string]string{}
	if len(matched) == 0 {
		return result, filename
	}
	for _, s := range strings.Split(matched[1:len(matched)-1], ",") {
		i := strings.Index(s, "=")
		if i < 0 {
			continue
		}
		key := strings.TrimSpace(s[0:i])
		value := strings.TrimSpace(s[i+1:])
		result[key] = value
	}

	return result, re.ReplaceAllString(filename, "")
}

func SetPathParams(filename string, params PathParams) string {
	if len(params) == 0 {
		return filename
	}
	re := regexp.MustCompile(`\[.*?\]`)
	matched := re.FindString(filename)
	filename = re.ReplaceAllString(filename, "")
	tmp := map[string]string{}

	if len(matched) > 0 {
		for _, s := range strings.Split(matched[1:len(matched)-1], ",") {
			i := strings.Index(s, "=")
			if i < 0 {
				continue
			}
			key := strings.TrimSpace(s[0:i])
			value := strings.TrimSpace(s[i+1:])
			tmp[key] = value
		}
	}

	for k, v := range params {
		tmp[k] = v
	}

	keys := []string{}
	for k := range tmp {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	buf := []string{}
	for _, k := range keys {
		buf = append(buf, k+"="+tmp[k])
	}
	paramsStr := "[" + strings.Join(buf, ",") + "]"
	dotIndex := strings.Index(filename, ".")
	if dotIndex < 0 {
		dotIndex = len(filename)
	}
	result := filename[0:dotIndex] + paramsStr + filename[dotIndex:]

	return result
}

func HasPathParams(filename string, params PathParams) bool {
	matched, err := regexp.MatchString(`\[.+?\]`, filename)
	if err != nil {
		panic(err)
	}
	return matched
}

func RelativeFrom(targetLink, srcPage string) string {
	if !filepath.IsAbs(targetLink) {
		return targetLink
	}
	return mustRel(filepath.Dir(srcPage), targetLink)
}
