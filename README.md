# doze

A modulable and minimalist build system.

## Goals

- Artifact based
- Easy to operate and to extend
- Local and remote build caches
- Not slow with large codebases

In addition here are some stretch goals:

- Remote build execution

## Design

Doze works with files on disk, which we call artifacts. It turns input artifacts into output artifacts following a build graph defined in a Dozefile. The inputs are turned into outputs following a user-defined procedure written directly in Go.

A build can be triggered by running doze manually from a terminal like it is usually the case with traditional build systems. Doze can also be launched in the background and left to monitor inputs of a project, automatically rebuilding only the parts that are affected by the operator/developer.
