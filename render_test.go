package moontpl

import (
	"fmt"
	"strings"
	"testing"
)

func TestSimple(t *testing.T) {
	testRender(t, Data{code: `
require("html").importGlobals()
return DIV {
	P "Hello, world"
}
	`, expected: `
<div>
    <p>Hello, world</p>
</div>
	`})
}

func TestNested(t *testing.T) {
	testRender(t, Data{code: `
require("html").importGlobals()
return DIV {
	H1 "this is a heading";
	H2 "this is a subheading";

	P [[
		A paragraph of text is here. Sentence follows naturally.
	]];
	P {
	   [[ Another paragraph starts. ]] ;
	   I "Italic, but not italian" ;
	   B "Bold, and emphasized, barely a sentence.";
	};

	DIV {
		id="provided";
		class="middle",
		"A div with a middle class and a provided id";
	}
}
	`, expected: `
<div>
    <h1>this is a heading</h1>
    <h2>this is a subheading</h2>
    <p> A paragraph of text is here. Sentence follows naturally.
    </p>
    <p> Another paragraph starts. <i>Italic, but not italian</i><b>Bold, and emphasized, barely a sentence.</b></p>
    <div id="provided" class="middle">A div with a middle class and a provided id</div>
</div>
	`})

}

func TestHTMLEscape(t *testing.T) {
	testRender(t, Data{code: `
require("html").importGlobals()
return DIV {
	P {
		"<script>alert('not okay')</script>"
	};
	P {
		["--noHTMLEscape"]=true;
		"<script>alert('okay')</script>"
	};
}
	`, expected: `
<div>
    <p>&lt;script&gt;alert('not okay')&lt;/script&gt;</p>
    <p>
        <script>
            alert('okay')
        </script>
    </p>
</div>
	`})
}

func TestFragment(t *testing.T) {
	testRender(t, Data{
		code: `
require("html").importGlobals()
return DIV {
	FRAGMENT(SPAN "one");
	FRAGMENT{SPAN "two", SPAN "three"};
	function() 
		return SPAN "four";
	end;
	{
		SPAN "six";
		SPAN "seven";
	};
}
	`,
		expected: `
<div>
    <span>one</span><span>two</span><span>three</span><span>four</span><span>six</span><span>seven</span>
</div>
	`})

}

func TestOperator(t *testing.T) {
	testRender(t, Data{
		code: `
require("html").importGlobals()
return FRAGMENT {
	P "111"
	/ I "222"
	/ B "333";
	
	P ^ I ^ B "333"
}
	`,
		expected: `
<p>111<i>222</i><b>333</b></p>
<p>
    <i><b>333</b></i>
</p>
	`})
}

func TestMultilineParagraphs(t *testing.T) {
	// TODO: collapse whitespace
	testRender(t, Data{
		code: `
require("html").importGlobals()
return FRAGMENT {
	PP {class="para"}
	/ [[This is a paragraph.

		This is another paragraph.
	    This one is a ]]
	/ A {href="localhost", "link"}
	/ [[.

		Third paragraph.

		Fourth ]] / EM "paragraph.";
}
	`,
		expected: `
<p class="para">This is a paragraph.</p>
<p class="para">
    This is another paragraph.
    This one is a <a href="localhost">link</a>.
</p>
<p class="para">
    Third paragraph.
</p>
<p class="para">
    Fourth <em>paragraph.</em>
</p>
	`})

}

func TestBuildHook(t *testing.T) {
	output, err := New().RenderFile("test/data/hook-build.html.lua")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "[hook]") {
		t.Error("failed to add render hook")
		t.Logf("output: \n%s", output)
	}
}

func printComparison(expected, actual string) {
	s := fmt.Sprintf("\n------[expected]------ \n%s\n------[ actual ]------\n%s", expected, actual)
	println(s)
}

type Data struct {
	code     string
	expected string
}

func testRender(t *testing.T, data Data) {
	output, err := New().RenderString(data.code)
	if err != nil {
		t.Errorf("failed to execute lua: %s\n%s", err, data.code)
	}
	output = strings.TrimSpace(output)
	expected := strings.TrimSpace(data.expected)

	if expected != output {
		t.Errorf("unexpected output")
		printComparison(expected, output)
	}
}
