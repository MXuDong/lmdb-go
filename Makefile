
.PHONY: deps all test race bin

BRANCH=`git rev-parse --abbrev-ref HEAD`
COMMIT=`git rev-parse --short HEAD`
MASTER_COMMIT=`git rev-parse --short origin/master`
GOLDFLAGS="-X main.branch $(BRANCH) -X main.commit $(COMMIT)"

deps:
	go get -d ./...

bin:
	mkdir -p bin
	GOBIN=${PWD}/bin go install ./exp/cmd/...
	GOBIN=${PWD}/bin go install ./cmd/...

all: deps check race bin

test:
	go test -cover ./lmdb ./exp/lmdbpool

race:
	go test -race ./lmdb ./exp/lmdbpool

lint:
	./build/bin/golangci-lint run --new-from-rev=$(MASTER_COMMIT) ./...

lintci-deps:
	rm -f ./build/bin/golangci-lint
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b ./build/bin v1.30.0

check:
	which goimports > /dev/null
	find . -name '*.go' | xargs goimports -d | tee /dev/stderr | wc -l | xargs test 0 -eq
	which golint > /dev/null
	golint ./... | tee /dev/stderr | wc -l | xargs test 0 -eq

install:
	cd lmdb && make clean && make install # use DESTDIR