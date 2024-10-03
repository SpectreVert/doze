package internal

import "runtime/debug"

// Reports the version of the main package of the binary.
//
// To get version information we need to support two different approaches.
// "go install remote path" works with the `debug` package
// "go build" is supported with the linker
// See https://github.com/golang/go/issues/50603
func Version(linkerVersion string) string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(stripped) " + linkerVersion
	}
	mod := &info.Main
	if mod.Replace != nil {
		mod = mod.Replace
	}

	if mod.Version != "" && mod.Version != "(devel)" {
		return mod.Path + " " + mod.Version
	}
	return mod.Path + " " + linkerVersion
}
