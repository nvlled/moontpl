local runtags

---@type { [string]: boolean|nil }
runtags = {} ---
--- Runtags are like environment variables,
--- except it doesn't include environment variables.
--- Command names (build, run or serve) are automatically added as tags.
--- Additional tags can be added on the CLI.
--- The purpose of runtags is to change the behaviour on different
--- environments.
---
--- Example:
--- $ moontpl build -t foo # runtags include: build, foo 
--- $ moontpl run -t foo   # runtags include: run, foo 


return runtags
