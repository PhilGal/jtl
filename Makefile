test: 
	go test ./... #Run all tests

bin:
	GOOS=linux GOARCH=amd64 go build -o ./bin/linux/jtl-linux-x64
	GOOS=windows GOARCH=amd64 go build -o ./bin/win/jtl-win-x64.exe
	GOOS=darwin GOARCH=amd64 go build -o ./bin/osx/jtl-osx-x64

clean:
	rm -rf bin
	rm -f jtl
	rm -f *.log

install:	
	make clean bin
	ln ./bin/osx/jtl-osx-x64 jtl