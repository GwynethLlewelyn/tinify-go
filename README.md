![Tinify API client for Golang logo](testdata/assets/tinify-go-logo-pangopher-128x128.png)

# Tinify API client for Golang

[:book: 国内的朋友看这里](http://www.jianshu.com/p/5c4161db4ac8)  
[![Go](https://github.com/GwynethLlewelyn/tinify-go/actions/workflows/go.yml/badge.svg)](https://github.com/GwynethLlewelyn/tinify-go/actions/workflows/go.yml)
---

Golang client for the [Tinify API](https://tinypng.com/developers/reference), used for [TinyPNG](https://tinypng.com) and [TinyJPG](https://tinyjpg.com). Tinify compresses or resizes your images intelligently. Read more at <https://tinify.com>.

The code on this repository is the work of volunteers who are neither affiliated with, nor endorsed by Tinify B.V., the makers of the Tinify API and of TinyPNG.

## Documentation

[Go to the documentation for the HTTP client](https://tinypng.com/developers/reference).

## Installation

Install the API client with `go install`.

```shell
go install github.com/gwpp/tinify-go
```

Note that this repository will install two different things:

1. A port of the Tinify API to the Go programming language, which will be placed under its own directory `./tinify`, a stand-alone package/module named `Tinify`;
2. A client application, at the root of the repository, which will compile to an executable binary, using the `Tinify` package as an imported module.

The client application serves both as a testing tool and as a stand-alone binary. It is very loosely based on other, similar client applications [written for other programming languages](https://github.com/tinify/).

Remember, to use it, you need a valid Tinify API Key, passed via the environment variable `TINIFY_API_KEY`.

## Usage

### About the TinyPNG API key

Get your API key from <https://tinypng.com/developers>.

### Compress

```golang
func TestCompressFromFile(t *testing.T) {
    Tinify.SetKey(Key)
    source, err := Tinify.FromFile("./testdata/input/test.jpg")
    if err != nil {
        t.Error(err)
        return
    }

    err = source.ToFile("./testdata/output/CompressFromFile.jpg")
    if err != nil {
        t.Error(err)
        return
    }
    t.Log("Compress successful")
}
```

### Resize

```golang
func TestResizeFromBuffer(t *testing.T) {
    Tinify.SetKey(Key)

    buf, err := ioutil.ReadFile("./testdata/input/test.jpg")
    if err != nil {
        t.Error(err)
        return
    }
    source, err := Tinify.FromBuffer(buf)
    if err != nil {
        t.Error(err)
        return
    }

    err = source.Resize(&Tinify.ResizeOption{
        Method: Tinify.ResizeMethodScale,
        Width:  200,
    })
    if err != nil {
        t.Error(err)
        return
    }

    err = source.ToFile("./testdata/output/ResizesFromBuffer.jpg")
    if err != nil {
        t.Error(err)
        return
    }
    t.Log("Resize successful")
}
```

## ⚠️ Notice:

`Tinify.ResizeMethod()` supports `scale`, `fit`, `cover` and `thumbnail`. If you use `fit`/`cover`/`thumbnail`, you **must** provide **both a width and a height**. But if you use `scale`, you **must** instead provide _either_ a target width _or_ a target height, **but not both**.

For further usage, please read the comments in [tinify_test.go](./tinify_test.go)

## Running tests

```shell
cd $GOPATH/src/github.com/gwpp/tinify-go
go test
```

## Command-line utility

This is a work-in-progress example/demonstration of most of the functionality with a compact CLI, using <https://github.com/urfave/cli/v3> (and `zerolog` for pretty-printing error messages). It was mostly created by Gwyneth Llewelyn to have the ability to tinify *several* images larger than 5 MBytes, using a simple `bash` script and some shell globbing magic.

To build it:

```shell
cd $GOPATH/src/github.com/gwpp/tinify-go
go build -ldflags "-X main.TheBuilder=$USER"
# or, if you prefer, `go install`
```

and then invoke `./tinify-go --help` to get some basic instructions for the CLI.

Remember that you need your `TINIFY_API_KEY`.

To override the logging level, you can either use `--debug`, or even catch some initialisation errors if you set 
`TINIFY_API_DEBUG` to, say, `trace`.

Without arguments, `tinify-go` will read from standard input and write to standard output (with error messages going to standard error). This, however, is designed for automation — if `tinify-go` detects that it is attached to a console (TTY), it will refuse to read from standard input — you *must* supply a file (or an URL for a file) instead. This is deliberate, to avoid typing endless characters in an attempt to "do something", pressing <kbd>Ctrl-D</kbd> by mistake, and sending garbage to the Tinify API endpoint — wasting resources and *possibly* even consuming one of your tokens! 

## License

This software is licensed under the [MIT License](LICENSE).

The logo (a cross-breed between the gopher mascot and the TinyPNG panda!) was provided courtesy of Microsoft's image generative AI, which is currently based on OpenAI's DALL-E technology.

[![tinify-go build with Debian packages](https://github.com/GwynethLlewelyn/tinify-go/actions/workflows/go.yml/badge.svg)](https://github.com/GwynethLlewelyn/tinify-go/actions/workflows/go.yml) [![Codacy Security Scan](https://github.com/GwynethLlewelyn/tinify-go/actions/workflows/codacy.yml/badge.svg)](https://github.com/GwynethLlewelyn/tinify-go/actions/workflows/codacy.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/GwynethLlewelyn/tinify-go)](https://goreportcard.com/report/github.com/GwynethLlewelyn/tinify-go) [![CodeQL](https://github.com/GwynethLlewelyn/tinify-go/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/GwynethLlewelyn/tinify-go/actions/workflows/codeql-analysis.yml)