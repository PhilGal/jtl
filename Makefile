test: 
	go test ./... #Run all tests

bin:
	GOOS=linux GOARCH=amd64 go build -o ./bin/linux/jtl
	GOOS=windows GOARCH=amd64 go build -o ./bin/win/jtl.exe
	GOOS=darwin GOARCH=amd64 go build -o ./bin/osx/jtl

clean:
	rm -rf bin
	rm -f jtl
	rm -f *.log

install:
	make clean bin
	ln ./bin/osx/jtl jtl