package moontpl

import lua "github.com/yuin/gopher-lua"

func init() {
	// Remove default lua path
	lua.LuaPathDefault = ""

	moontpl = New()
}
