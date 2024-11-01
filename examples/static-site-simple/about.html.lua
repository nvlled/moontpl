require("web")
local LAYOUT = require("layout")

return LAYOUT {
    P {
        [[Here's a picture of a whale floating in the moonlit sky:]];
        A {href="more.html"; IMG {src="white-whale.jpg"}};
        [[
        Click image to see more.
        ]]
    };

    P {"Go "; A {href="index.html"; "back"}; "."};
}
