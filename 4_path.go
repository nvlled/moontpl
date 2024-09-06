package moontpl

import (
	"io/fs"
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/samber/lo"
	lua "github.com/yuin/gopher-lua"
)

const NO_RECURSE = "__NO_RECURSE__"

type PathParams map[string]string
type PagePath struct {
	AbsFile string
	RelFile string
	Link    string
}

type Page struct {
	PagePath
	Data map[string]any
}

func GetPageFilenames(baseDir string) []PagePath {
	var result []PagePath
	filepath.WalkDir(baseDir, func(filename string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(filename, ".html.lua") {
			result = append(result, getPagePath(filename))
		}
		return nil
	})
	return result
}

func getPagePath(filename string) PagePath {
	filename = lo.Must(filepath.Abs(filename))

	link := lo.Must(filepath.Rel(SiteDir, filename))
	link = "/" + strings.TrimSuffix(link, ".lua")

	return PagePath{
		AbsFile: filename,
		RelFile: lo.Must(filepath.Rel(SiteDir, filename)),
		Link:    link,
	}
}

func GetPages(L *lua.LState) ([]Page, error) {
	if L.GetGlobal(NO_RECURSE) != lua.LNil {
		return []Page{}, nil
	}

	L.SetGlobal(NO_RECURSE, lua.LNumber(1))
	defer L.SetGlobal(NO_RECURSE, lua.LNil)

	result := []Page{}
	for _, entry := range GetPageFilenames(".") {
		data, err := GetPageData(L, entry.AbsFile)
		if err != nil {
			log.Printf("failed to get pages: %v", err)
			return nil, err
		}
		result = append(result, Page{
			PagePath: entry,
			Data:     data,
		})
	}

	return result, nil
}

func GetPageData(L *lua.LState, filename string) (map[string]any, error) {
	if err := L.DoFile(filename); err != nil {
		return map[string]any{}, err
	}
	lv := L.Get(-1)

	if lv.Type() != lua.LTTable {
		return map[string]any{}, nil
	}

	table := lv.(*lua.LTable)
	result := map[string]any{}

	if data, ok := table.RawGet(lua.LString("data")).(*lua.LTable); ok {
		data.ForEach(func(k, v lua.LValue) {
			switch v.Type() {
			case lua.LTBool:
				result[k.String()] = lua.LVAsBool(v)
			case lua.LTNumber:
				result[k.String()] = lua.LVAsNumber(v)
			case lua.LTString:
				result[k.String()] = lua.LVAsString(v)
			}
		})
	}

	return result, nil
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
	return lo.Must(filepath.Rel(filepath.Dir(srcPage), targetLink))
	//return lo.Must(filepath.Rel(targetLink, filepath.Dir(srcPage)))
}
