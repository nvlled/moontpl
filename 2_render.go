package moontpl

import (
	"github.com/nvlled/htmlformat"
	lua "github.com/yuin/gopher-lua"
)

func (m *Moontpl) RenderFile(filename string) (string, error) {
	L := m.getState(filename)
	defer m.luaPool.Put(L)

	lv, err := m.renderFile(L, filename)
	if err != nil {
		return "", err
	}

	if lv.Type() == lua.LTNil {
		return "", nil
	}

	output := L.ToStringMeta(lv).String()
	if wholeExt(filename) == ".html.lua" {
		output = htmlformat.Format(output)
	}

	return output, nil
}

func (m *Moontpl) RenderString(luaCode string) (string, error) {
	L := m.getState("-")
	defer m.luaPool.Put(L)

	if err := L.DoString(luaCode); err != nil {
		return "", err
	}

	lv := L.Get(-1)
	if lv.Type() == lua.LTNil {
		return "", nil
	}

	output := L.ToStringMeta(lv).String()
	output = htmlformat.Format(output)

	return output, nil
}

func (m *Moontpl) renderFile(L *lua.LState, filename string) (lua.LValue, error) {
	if hasPathParams(filename) {
		var params pathParams
		params, filename = extractPathParams(filename)
		pageData := PageData{}
		for k, v := range params {
			pageData[k] = v
		}
		m.SetPageData(L, pageData)
	}

	if err := L.DoFile(filename); err != nil {
		return lua.LNil, err
	}

	lv := L.Get(-1)
	if lv.Type() == lua.LTNil {
		return lua.LNil, nil
	}

	if mod, ok := getLoadedModule(L, "page").(*lua.LTable); ok {
		onPageRender, isFunc := mod.RawGetString("onRender").(*lua.LFunction)
		if !isFunc {
			onPageRender, isFunc = mod.RawGetString("onRenderDefault").(*lua.LFunction)
		}
		if isFunc {
			err := L.CallByParam(lua.P{
				Fn:      onPageRender,
				NRet:    1,
				Protect: true,
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
