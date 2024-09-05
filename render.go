package moontpl

import lua "github.com/yuin/gopher-lua"

func renderFile(L *lua.LState, filename string) (string, error) {
	if err := L.DoFile(filename); err != nil {
		return "", err
	}

	lv := L.Get(-1)
	if lv.Type() == lua.LTNil {
		return "", nil
	}

	return L.ToStringMeta(lv).String(), nil
}

func RenderFile(filename string) (string, error) {
	L := createState(filename)
	defer L.Close()
	return renderFile(L, filename)
}

func RenderString(luaCode string) (string, error) {
	L := createState("inline.lua")
	defer L.Close()
	if err := L.DoString(luaCode); err != nil {
		return "", err
	}

	lv := L.Get(-1)
	if lv.Type() == lua.LTNil {
		return "", nil
	}
	return L.ToStringMeta(lv).String(), nil
}
