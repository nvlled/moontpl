package moontpl

import "testing"

func TestExtract(t *testing.T) {
	docs, ok, err := extractDocumentation("./lua/html.lua")
	if err != nil {
		panic(err)
	}
	if !ok {
		t.Errorf("failed to generate docs for html")
		println(docs)
	}
}
