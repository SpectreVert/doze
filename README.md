# doze

A modulable and minimalist build system.

## Goals

- [X] Artifacts based
- [X] Easy to operate and to extend
- [ ] Support for expressing source and build directories (todo)
- [ ] Local and remote build caches (todo)

## Design

Doze works with files on disk, which we call artifacts. It turns input artifacts into output artifacts following a build graph defined in a Dozefile. The inputs are turned into outputs following a user-defined procedure written directly in Go and embedded in the doze binary.

A build can be triggered by running doze manually from a terminal like it is usually the case with traditional build systems. Doze can also be launched in the background and left to monitor inputs of a project, automatically rebuilding only the parts that are affected by the operator/developer.

## Usage

doze does not implement a [DSL](https://en.wikipedia.org/wiki/Domain-specific_language) to configure its builds. Instead it uses a very simple `yaml` schema declared in a `Dozefile.yaml` file as described in the example below.

```yaml
rules:
  - do: lang:c:object_file
    inputs: [parse.c, parse.h]
    outputs: [parse.o]

  - do: lang:c:object_file
    inputs: [main.c, parse.h]
    outputs: [main.o]

  - do: lang:c:executable
    inputs: [parse.o, main.o]
    outputs: [exe]
```

A rule defines a procedure to use, a set of input artifacts and a set of outputs. In this repository are provided two example procedures, `lang:c:object_file` and `lang:c:executable` to respectively build C object files and C executables. Their implementation is store in [procedures/lang_c/main.go](./procedures/lang_c/main.go). Registering a procedure in the `init` function of their modules embeds it directly in the doze executable.

There is restriction on the order in which the rules are declared. Doze will only execute rules which are ready to execute; that is those which have all of their inputs ready for processingj.
