package moontpl

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type pathParams map[string]string

func getPathParams(filename string) pathParams {
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

func extractPathParams(filename string) (pathParams, string) {
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

func SetPathParams(filename string, params pathParams) string {
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

func hasPathParams(filename string) bool {
	matched, err := regexp.MatchString(`\[.+?\]`, filename)
	if err != nil {
		panic(err)
	}
	return matched
}

func relativeFrom(targetLink, srcPage string) string {
	if !filepath.IsAbs(targetLink) {
		return targetLink
	}
	return mustRel(filepath.Dir(srcPage), targetLink)
}
