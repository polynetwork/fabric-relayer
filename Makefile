# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

relayer:
	$(GOBUILD) -o build/bin/relayer cmd/*

build-linux-robot:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o build/target/linux-robot cmd/main.go

run:
	@echo test case $(t)
	./build/bin/relayer -config=build/target/config.json -t=$(t)

clean:
	rm -rf build/bin/*