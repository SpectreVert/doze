# Doze

A modulable and minimalist file-processing system.

## Overview

Doze is an executable and a library that allow you to compose file-processing workflows using Go.

A [directed acyclic graph](https://en.wikipedia.org/wiki/Directed_acyclic_graph) keeps track of the dependencies between files and runs in order the different processing functions (procedures) if the rule is out-of-date. A unit of work in Doze is called a rule. A rule must have at least one input and must create at least one output. A rule is bound to exactly one procedure. Inputs and outputs of rules are also called artifacts.

Doze also has a caching interface implemented for local and remote caching of artifacts.
