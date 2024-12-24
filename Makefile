build: build-cli build-web

build-cli: 
	go build -o ./bin/sqli-scanner.exe ./cmd/cli/

build-web:
	go build -o ./bin/web.exe ./cmd/web/

build-script:
	go build -o ./bin/script.exe ./scripts/

run-cli: build-cli
	./bin/cli.exe