VERSION=`git describe --tags`
LDFLAGS=-ldflags "-w -s -X github.com/philgal/jtl/cmd.Version=${VERSION}"
GOBIN=/usr/local/bin
test:
	go test ./... #Run all tests

bin:
	# GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ./bin/linux/jtl-linux-x64
	# GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ./bin/win/jtl-win-x64.exe
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o ./bin/osx/jtl-osx-x64

clean:
	rm -rf bin
	rm -f jtl
	rm -f *.log

install:
	make clean bin
	cp ./bin/osx/jtl-osx-x64 /usr/local/bin/jtl
