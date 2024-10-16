package moontpl

import (
	"testing"
)

func TestGetPathParams(t *testing.T) {
	type entry struct {
		filename string
		params   pathParams
	}
	for _, entry := range []entry{
		{"/test[x=1].html", pathParams{"x": "1"}},
		{"/test[aa=bb,cc=dddd].html", pathParams{"aa": "bb", "cc": "dddd"}},
		{"/test[key   = value , a = 123].html", pathParams{"key": "value", "a": "123"}},
		{"/test[].html", pathParams{}},
		{"/test.html", pathParams{}},
	} {
		params := getPathParams(entry.filename)
		if len(params) != len(entry.params) {
			t.Errorf("expected: %v, got %v", entry.params, params)
			t.Fail()
		}
		for k, v := range entry.params {
			if params[k] != v {
				t.Errorf("expected: %v, got %v", entry.params, params)
				t.Fail()
			}
		}
	}
}

func TestSetPathParams(t *testing.T) {
	type entry struct {
		input    string
		params   pathParams
		expected string
	}
	for _, entry := range []entry{
		{"", pathParams{}, ""},
		{"/test[x=1].html", pathParams{"x": "2"}, "/test[x=2].html"},
		{"/test[x=1]", pathParams{"x": "2"}, "/test[x=2]"},
		{"/test", pathParams{"x": "2"}, "/test[x=2]"},
		{"/test", pathParams{}, "/test"},
		{"/test[x=1,y=2].html", pathParams{"x": "3"}, "/test[x=3,y=2].html"},
	} {
		filename := setPathParams(entry.input, entry.params, false)
		if filename != entry.expected {
			t.Errorf("expected: %v, got %v", entry.expected, filename)
		}
	}
}

func TestRelativePath(t *testing.T) {
	for _, entry := range [][]string{
		{"/a.png", "/dir1/index.html", "../a.png"},
		{"/dir2/a.png", "/dir1/index.html", "../dir2/a.png"},
		{"/dir1/a.png", "/dir1/index.html", "a.png"},
		{"/dir1/dir3/a.png", "/dir1/index.html", "dir3/a.png"},
		{"a.png", "/dir1/index.html", "a.png"},
	} {

		actual := relativeFrom(entry[0], entry[1])
		if actual != entry[2] {
			t.Errorf("expected: %v, got %v", entry[2], actual)
		}
	}
}
