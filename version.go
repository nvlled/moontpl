package moontpl

import (
	"bytes"
	"runtime/debug"
)

// Set the Version variable with ldflags.
// For example: go build -ldflags="-X github.com/nvlled/moontpl.Version=v0.1.0"
// If ldflags is not set, version will be read from buildInfo.
var Version = ""

func init() {
	if Version != "" {
		return
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		var buf bytes.Buffer
		buf.WriteString(buildInfo.Main.Version)
		for _, e := range buildInfo.Settings {
			switch e.Key {
			case "vcs.revision", "vcs.time":
				buf.WriteRune(' ')
				buf.WriteString(e.Value)
			}
		}
		Version = buf.String()
	}
}
