![Tinify API client for Golang logo](assets/tinify-go-logo-pangopher-128x128.png)

# Tinify API client for Golang

[:book: 国内的朋友看这里](http://www.jianshu.com/p/5c4161db4ac8)

---

Golang client for the [Tinify API](https://tinypng.com/developers/reference), used for [TinyPNG](https://tinypng.com) and [TinyJPG](https://tinyjpg.com). Tinify compresses or resize your images intelligently. Read more at [http://tinify.com](http://tinify.com).

## Documentation

[Go to the documentation for the HTTP client](https://tinypng.com/developers/reference).

## Installation

Install the API client with `go get`.

```shell
go get -u github.com/gwpp/tinify-go
```

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
