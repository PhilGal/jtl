VERSION=`git describe --tags`
LDFLAGS=-ldflags "-w -s -X github.com/philgal/jtl/cmd.Version=${VERSION}"
GOBIN=/usr/local/bin
test:
	go test ./... #Run all tests

bin:
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o ./bin/osx/jtl-osx-x64

clean:
	rm -rf bin
	rm -f jtl
	rm -f *.log

install:
	make clean bin
	cp ./bin/osx/jtl-osx-x64 /usr/local/bin/jtl
