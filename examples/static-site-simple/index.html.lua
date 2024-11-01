require("web")
local LAYOUT = require("layout")

return LAYOUT {
    H1 "Hello";
    P [[
        This is a simple example site that demonstrates
        how to use basic stuffs with moontpl.
    ]];
    
    HR();

    P {
        class="fancy";
        "This is a text with some "; EM "CSS"; " style."
    };
    
    
    H2 "A list of random things";
    UL {
        LI "one";
        LI "apple";
        LI "chicken";
        LI "pie";
    };
    
    P {
        "See "; A  {href="about.html", "other page"}; " for another example."
    };
    
    STYLE {
        CSS ".fancy" {
            color="orange";
            border="5px double teal";
            display="inline-block";
        }
    }
}