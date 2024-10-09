build: build-cli build-web

build-cli: 
	go build -o ./bin/cli.exe ./cmd/cli/

build-web:
	go build -o ./bin/web.exe ./cmd/web/