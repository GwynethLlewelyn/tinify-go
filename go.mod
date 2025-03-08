module github.com/gwpp/tinify-go

go 1.23.0

toolchain go1.24.1

require (
	github.com/joho/godotenv v1.5.1
	github.com/rs/zerolog v1.33.0
	github.com/spf13/pflag v1.0.6
)

require (
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.31.0 // indirect
)

replace github.com/gwpp/tinify-go v0.0.0-20170613055357-77b9df15f343 => github.com/GwynethLlewelyn/tinify-go v0.1.1-0.20231112021032-de06fee9c2ac
