![Tinify API client for Golang logo](assets/tinify-go-logo-pangopher-128x128.png)

# Tinify API client for Golang

[:book: 国内的朋友看这里](http://www.jianshu.com/p/5c4161db4ac8)

---

Golang client for the [Tinify API](https://tinypng.com/developers/reference), used for [TinyPNG](https://tinypng.com) and [TinyJPG](https://tinyjpg.com). Tinify compresses or resizes your images intelligently. Read more at [http://tinify.com](http://tinify.com).

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

Remember, to use it, you need a valid Tinify API Key, passed wither with the `--key=XXXX` flag, or via the environment variable `TINIFY_API_KEY`.

## Usage

-   About the TinyPNG API key

    Get your API key from https://tinypng.com/developers

-   Compress

    ```golang
    func TestCompressFromFile(t *testing.T) {
        Tinify.SetKey(Key)
        source, err := Tinify.FromFile("./test.jpg")
        if err != nil {
            t.Error(err)
            return
        }

        err = source.ToFile("./test_output/CompressFromFile.jpg")
        if err != nil {
            t.Error(err)
            return
        }
        t.Log("Compress successful")
    }
    ```

-   Resize

    ```golang
    func TestResizeFromBuffer(t *testing.T) {
        Tinify.SetKey(Key)

        buf, err := ioutil.ReadFile("./test.jpg")
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

        err = source.ToFile("./test_output/ResizesFromBuffer.jpg")
        if err != nil {
            t.Error(err)
            return
        }
        t.Log("Resize successful")
    }
    ```

-   **_Notice:_**

    `Tinify.ResizeMethod()` supports `scale`, `fit` and `cover`. If you use either fit or cover, you must provide **both a width and a height**. But if you use scale, you must instead provide _either_ a target width _or_ a target height, **but not both**.

-   For further usage, please read the comments in [tinify_test.go](./tinify_test.go)

## Running tests

```shell
cd $GOPATH/src/github.com/gwpp/tinify-go
go test
```

## License

This software is licensed under the MIT License. [View the license](LICENSE).

The logo (a cross-breed between the gopher mascot and the TinyPNG panda!) was provided courtesy of Microsoft's image generative AI, which is currently based on OpenAI's DALL-E technology.
