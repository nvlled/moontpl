package moontpl

import (
	"bytes"
	"fmt"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

const dependencyIndex = lua.LNumber(-9988002)

type dependentTable lua.LTable

func (dt *dependentTable) GetModules() []string {
	t := (*lua.LTable)(dt)
	var result []string
	t.ForEach(func(k, _ lua.LValue) {
		if str, ok := k.(lua.LString); ok {
			result = append(result, string(str))
		}
	})
	return result
}

func (dt *dependentTable) AddDependentOf(L *lua.LState, parentModule, dependentModule string) {
	t := (*lua.LTable)(dt)
	deps, ok := t.RawGetString(parentModule).(*lua.LTable)
	if !ok || deps == lua.LNil {
		deps = L.NewTable()
		t.RawSetString(parentModule, deps)
	}
	deps.RawSetString(dependentModule, lua.LTrue)
}

func (dt *dependentTable) RemoveDependents(parentModule string) {
	t := (*lua.LTable)(dt)
	t.RawSetString(parentModule, lua.LNil)
}

func (dt *dependentTable) GetDependentsOf(moduleName string) []string {
	t := (*lua.LTable)(dt)
	dependents, ok := t.RawGetString(moduleName).(*lua.LTable)
	var result []string
	if !ok {
		return result
	}
	dependents.ForEach(func(k, _ lua.LValue) {
		if s, ok := k.(lua.LString); ok {
			result = append(result, string(s))
		}
	})
	return result
}

func (dt *dependentTable) ClearParent(moduleName string) {
	t := (*lua.LTable)(dt)
	t.ForEach(func(k, v lua.LValue) {
		if dependents, ok := v.(*lua.LTable); ok {
			dependents.RawSetString(moduleName, lua.LNil)
		}
	})
}

func (dt *dependentTable) String() string {
	t := (*lua.LTable)(dt)

	var buffer bytes.Buffer
	t.ForEach(func(k, v lua.LValue) {
		if lname, ok := k.(lua.LString); ok {
			name := string(lname)
			buffer.WriteString(fmt.Sprintf("%v -> %+v\n", lname, dt.GetDependentsOf(name)))
		}
	})
	return buffer.String()
}

func requireWithDependencyTree(moontpl *Moontpl, L *lua.LState) lua.LGFunction {
	level := 0
	lineage := []string{}
	dependents := (*dependentTable)(L.NewTable())
	require := L.GetGlobal("require").(*lua.LFunction)
	L.G.Registry.RawSet(dependencyIndex, (*lua.LTable)(dependents))

	return func(L *lua.LState) int {
		name := L.ToString(1)
		if !moontpl.disableLuaPool {
			if len(lineage) >= 1 {
				dependents.AddDependentOf(L, name, lineage[len(lineage)-1])
			}

			level++
			lineage = append(lineage, name)

			defer func() {
				level--
				lineage = lineage[:len(lineage)-1]
			}()
		}

		top := L.GetTop()
		L.Push(require)
		L.Push(lua.LString(name))
		L.Call(1, lua.MultRet)

		return L.GetTop() - top
	}
}

type lStatePool struct {
	m     sync.Mutex
	saved []*lua.LState
}

func (pl *lStatePool) Get() *lua.LState {
	pl.m.Lock()
	defer pl.m.Unlock()

	n := len(pl.saved)
	if n == 0 {
		return nil
	}
	x := pl.saved[n-1]
	pl.saved = pl.saved[0 : n-1]
	return x
}

func (pl *lStatePool) Put(L *lua.LState) {
	pl.m.Lock()
	defer pl.m.Unlock()
	pl.saved = append(pl.saved, L)
}

func (pl *lStatePool) Clear() {
	pl.m.Lock()
	defer pl.m.Unlock()
	for _, L := range pl.saved {
		L.Close()
	}
	pl.saved = nil
}
func (pl *lStatePool) resetLoadedPoolModules(moduleName string) {
	pl.m.Lock()
	defer pl.m.Unlock()

	for _, L := range pl.saved {
		pl.resetModuleDependents(L, moduleName)
	}
}

func (pl *lStatePool) resetModuleDependents(L *lua.LState, moduleName string) {
	var dependents *dependentTable
	if t, ok := L.G.Registry.RawGet(dependencyIndex).(*lua.LTable); !ok {
		return
	} else {
		dependents = (*dependentTable)(t)
	}

	queue := []string{moduleName}
	visited := map[string]struct{}{}
	for len(queue) > 0 {
		name := queue[len(queue)-1]
		queue = queue[:len(queue)-1]

		if _, ok := visited[name]; ok {
			continue
		} else {
			visited[name] = struct{}{}
		}

		queue = append(queue, dependents.GetDependentsOf(name)...)
		dependents.RemoveDependents(name)

		loadedModules := L.GetField(L.Get(lua.RegistryIndex), "_LOADED").(*lua.LTable)
		loadedModules.RawSetString(name, lua.LNil)
	}

	dependents.ClearParent(moduleName)
}

func (pl *lStatePool) printLoadedModules() {
	pl.m.Lock()
	defer pl.m.Unlock()

	for _, L := range pl.saved {
		if t, ok := L.G.Registry.RawGet(dependencyIndex).(*lua.LTable); ok {
			print((*dependentTable)(t).String())
		}

	}
}
