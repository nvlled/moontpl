# How to use a module

To use a module, remove the file extension
and pass it on require().

Example:
```lua
local build = require("build") -- search for build.lua
```

# Stub functions
Some functions in the module have `-- stub` labels in them.
These functions will be implemented by the host environment.