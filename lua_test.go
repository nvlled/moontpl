package moontpl

import (
	"testing"
)

func TestGetPathParams(t *testing.T) {
	type entry struct {
		filename string
		params   PathParams
	}
	for _, entry := range []entry{
		{"/test[x=1].html", PathParams{"x": "1"}},
		{"/test[aa=bb,cc=dddd].html", PathParams{"aa": "bb", "cc": "dddd"}},
		{"/test[key   = value , a = 123].html", PathParams{"key": "value", "a": "123"}},
		{"/test[].html", PathParams{}},
		{"/test.html", PathParams{}},
	} {
		params := GetPathParams(entry.filename)
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
		params   PathParams
		expected string
	}
	for _, entry := range []entry{
		{"", PathParams{}, ""},
		{"/test[x=1].html", PathParams{"x": "2"}, "/test[x=2].html"},
		{"/test[x=1]", PathParams{"x": "2"}, "/test[x=2]"},
		{"/test", PathParams{"x": "2"}, "/test[x=2]"},
		{"/test", PathParams{}, "/test"},
		{"/test[x=1,y=2].html", PathParams{"x": "3"}, "/test[x=3,y=2].html"},
	} {
		filename := SetPathParams(entry.input, entry.params)
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

		actual := RelativeFrom(entry[0], entry[1])
		if actual != entry[2] {
			t.Errorf("expected: %v, got %v", entry[2], actual)
		}
	}
}
