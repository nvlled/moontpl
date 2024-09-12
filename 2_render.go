package moontpl

import lua "github.com/yuin/gopher-lua"

func (m *Moontpl) RenderFile(filename string) (string, error) {
	L := m.createState(filename)
	defer L.Close()

	lv, err := m.renderFile(L, filename)
	if err != nil {
		return "", err
	}
	return L.ToStringMeta(lv).String(), nil
}

func (m *Moontpl) RenderString(luaCode string) (string, error) {
	L := m.createState("inline.lua")
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

func (m *Moontpl) renderFile(L *lua.LState, filename string) (lua.LValue, error) {
	if err := L.DoFile(filename); err != nil {
		return lua.LNil, err
	}

	lv := L.Get(-1)
	if lv.Type() == lua.LTNil {
		return lua.LNil, nil
	}

	if hook, ok := getLoadedModule(L, "hook").(*lua.LTable); ok {
		onPageRender, isFunc := hook.RawGetString("onPageRender").(*lua.LFunction)
		if isFunc {
			err := L.CallByParam(lua.P{
				Fn:      onPageRender,
				NRet:    1,
				Protect: false,
			}, lv)

			if err != nil {
				return lua.LNil, err
			}

			if ret := L.Get(-1); ret != lua.LNil {
				lv = ret
			}
		}
	}

	return lv, nil
}
