local page = {}

-- Used when a lua html page is used as a
-- HTML/CSS templating engine. The page
-- can access the page.input to alter the output.
page.input = {}

-- The page data that is set by the lua page itself.
-- This serves as a metadata about the page itself.
-- The data set here is included in the page.list().
-- example:
--   local page = require("page")
--   page.data.title = "Home page"
--   page.data.desc = "Welcome"
page.data = {}

-- The link to the page currently being run or rendered.
page.PAGE_LINK = ""

-- The filename of thepage currently being run or rendered.
page.PAGE_FILENAME = ""

-- Returns the list of pages
-- found in the SITEDIR.
function page.list()
    -- stub
    return {}
end

-- Returns the list of filenames
-- found in the SITEDIR.
function page.files()
    -- stub
    return {}
end

return page
