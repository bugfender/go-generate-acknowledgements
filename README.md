# go-generate-acknowledgements
Generates acknowledgement text from a go.mod file.

This tool parses the given `go.mod` files and uses `github.com/google/licenseclassifier` to identify the license for each module, then output the text to the standard output, which you can capture.

## Installation

```shell
go install github.com/bugfender/go-generate-acknowledgements
```

## Usage

You can generate an acknowledgements file for your project this:

```shell
go-generate-acknowledgements > open-source-licenses.txt
```

## Configuration

You can provide one or multiple `go.mod` files as arguments, like this:

```shell
go-generate-acknowledgements project-a/go.mod project-b/go.mod
```

You can also provide a configuration file. By default, the file `go-generate-acknowledgements.yaml` is read (you can also prefix it with a `.`, or use the `yml` or `json` extensions). You can provide a configuration file as argument, like this:

```shell
go-generate-acknowledgements -config your-config.yaml
```

Example configuration:

```yaml
# by default, no licenses are allowed (you can run go-generate-acknowledgements to discover the licenses you use and approve them)
allowedlicenses:
  - Apache-2.0
  - MIT
# you can specify the go.mod files to parse (by default "go.mod")
gomodfiles:
  - module-a/go.mod
  - module-b/go.mod
# you can override licenses for specific modules if the license classifier can not identify them correctly
licenseoverrides:
  - moduleversion:
      path: github.com/dchest/uniuri
      version: v1.2.0
    spdx: CC0-1.0
```

# Support

This tool is open source and free to use (see `LICENSE`). We do not provide support.

We may implement new features, but we will likely only implement features that we need. If you'd like to contribute, check `CONTRIBUTING.md`.
