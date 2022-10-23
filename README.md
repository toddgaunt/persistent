# Persistent

[![Go](https://github.com/toddgaunt/persistent/actions/workflows/go.yml/badge.svg)](https://github.com/toddgaunt/persistent/actions/workflows/go.yml)

`persistent` is the top-level package for packages which provide
implementations of Clojure's persistent data structures (namely: lists,
vectors, and maps) for Go, using generic types.

## Vectors Benchmarks

![Assoc Performance Graph](./vectors/benchmark/assoc.png)

![Conj Performance Graph](./vectors/benchmark/conj.png)

![Nth Performance Graph](./vectors/benchmark/nth.png)

## For Developers
This section is intended as guidance for developers and contributors to this
project.
### Update package version
Simply run `./version.sh increment` and follow the prompts. The VERSION.txt
file will be updated and comitted to  Git automatically. After the commit, a
tag using the incremented version number is created.
### Publishing package version
After updating the version locally with `./version.sh increment`, the new version
can be published with `./version.sh publish`. This pushes the Git tag which matches
the current VERSION.txt file and publishes documentation to pkg.go.dev.
