package moontpl

import (
	"log"
	"testing"

	"github.com/samber/lo"
	lua "github.com/yuin/gopher-lua"
)

func TestCssRender(t *testing.T) { // TODO:
}

func TestHTMLRender(t *testing.T) {
	actual, err := RenderString(`
		local html = require("html")
		html.importGlobals()
		return DIV {
			x=1,
			y=2,
			z="zzz",
			BR,
			BR{},
			BR(),
			function()
				return "x"
			end,
			function()
				return P {"paragraph"}
			end,
			data={
				someData=789
			},
			--function()
			--	return B { "TODO" }
			--end,
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	expected := `<div x="1" y="2" z="zzz"><br/><br/><br/>x<p>paragraph</p></div>`
	if actual != expected {
		t.Errorf("output mismatch:\nactual  =%v\nexpected=%v", actual, expected)
	}
}

func TestLoading(t *testing.T) {
	code := `
	local TITLE = "title foo" + "blah"
	local DESC = "description"
	
	local X = 100
	local Y = {}
	local Z = "z"
	function f() 
		println("blah")
		f() 
		local blah = 1
	end
	function g() 
	end
	
	local foo = "huh"
	`

	L := lua.NewState()
	fn := lo.Must(L.LoadString(code))
	log.Printf("num constants  defined: %v", len(fn.Proto.Constants))

	i := 0
	for _, lv := range fn.Proto.Constants {
		log.Printf("%v: %v, %v", i, lv.String(), lv.Type().String())
		//if lv.Type() != lua.LTFunction {
		//}
		i++
	}

	println()
	i = 0
	for _, lv := range fn.Proto.DbgLocals {
		log.Printf("%v: %v, %v", i, lv.Name, lv.StartPc)
		i++
	}
}
