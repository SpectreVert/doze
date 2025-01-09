package internal

import "runtime/debug"

// Version reports the version of the main package of the binary.
//
// As of Go 1.21, we still need to use two different approaches to be able to
// report version information:
// 1. to support "go build", we must use the linker.
// 2. to support "go install with remote path", we must use the `debug` package.
// 3. as far as I understand it, "go install with local path" does not work.
// See https://github.com/golang/go/issues/50603, to be able to use the  `debug`
// package for all use cases.
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
