require("html").importGlobals() ---
require("css").importGlobals() ---
--- web is a module that imports globals from html and css modules.
---
--- Example:
---     require("web")
---     -- DIV, P, CSS are now avaiable for use.
---     DIV {
---         STYLE {
---             CSS 'div' { color="red"};
---             CSS 'body' { color="blue"};
---         };
---         P "test paragraph";
---     }
