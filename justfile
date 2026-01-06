version := `cat VERSION`

export CC := "zig cc"
export CXX := "zig c++"

build:
    go build -ldflags "-X main.version={{version}}" -o bin/tapshow ./cmd/tapshow

run: build
    ./bin/tapshow

test path="./..." *flags="":
    go test {{flags}} {{path}}

test-verbose:
    go test -v ./...

test-coverage:
    go test -cover ./...

build-release:
    go build -ldflags "-s -w -X main.version={{version}}" -o bin/tapshow ./cmd/tapshow
    upx --best bin/tapshow

clean:
    rm -rf bin/
