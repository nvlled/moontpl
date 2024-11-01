require("web")
local tags = require "runtags"

return function(args)
    return HTML {
        HEAD {
            TITLE="Simple example site";
            LINK {rel="stylesheet"; href="/style.css"};
            STYLE {
                CSS "a" {
                    color="red";
                };
            }
        };
        BODY {args};
    }
end
