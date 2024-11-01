require("web")
local LAYOUT = require("layout")

local images = {}
for i = 0, 2000 do
    table.insert(images, SPAN {
        style="padding-right:" .. tostring(i / 2) .. "px ";
        class="img-container" .. (i == 1804 and " highlight" or "");
        IMG {
            style="width:" .. tostring(i/2  ) .. "px ";
            src="white-whale.jpg"
        };
        B(tostring(i));
    })
end

return LAYOUT {
    BUTTON {
        id="top";
        "scroll down"
    };
    BR();

    FRAGMENT(images);

    STYLE {
        CSS ".img-container" {
            position="relative";
            CSS "b" {font_size="18px"; position="absolute"; top="0"; left="0"};
            [".highlight"]={filter="hue-rotate(140deg)"};
        };
        CSS "img" {width="200px"};
    };

    BR();
    BUTTON {
        id="bottom";
        "scroll up"
    };
    BR();

    P {
        "Go "; A  {href="index.html", "back"}; "."
    };

    SCRIPT [[
        function sleep(n) { return new Promise(f => setTimeout(f, n)); }
        const images = document.querySelectorAll("img");
        document.querySelector("button#top").onclick = async function() {
            for (let i = 0; i < images.length; i++) {
                images[i].scrollIntoView({behaviour: "smooth", block: "center"});
                await sleep(2);
            }
            window.scrollTo(0, window.scrollMaxY);
        }
        document.querySelector("button#bottom").onclick = async function() {
            for (let i = images.length-1; i >= 0; i--) {
                images[i].scrollIntoView({behaviour: "smooth", block: "nearest"});
                await sleep(2);
            }
            window.scrollTo(0, 0);
        }
    ]]
}
