module github.com/gwpp/tinify-go

go 1.24.6

require (
	github.com/GwynethLlewelyn/justify v0.2.1
	github.com/joho/godotenv v1.5.1
	github.com/rs/zerolog v1.34.0
	github.com/urfave/cli/v3 v3.6.1
	golang.org/x/term v0.37.0
)

require (
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.38.0 // indirect
)

replace github.com/gwpp/tinify-go v0.0.0-20170613055357-77b9df15f343 => github.com/GwynethLlewelyn/tinify-go v0.1.1-0.20231112021032-de06fee9c2ac
