.PHONY: test-raft test-raft-election \
		test-raft-fail-no-agree test-raft-concurrent-starts \
		test-raft-rejoin test-raft-backup \ 
		test-raft-persist \
		fmt-raft vet-raft clean

GO_ENV := GO111MODULE=off GOPATH=$(CURDIR)

test-raft:
	$(GO_ENV) go test raft

test-raft-election:
	$(GO_ENV) go test raft -run Election

test-raft-fail-no-agree:
	$(GO_ENV) go test raft -run FailNoAgree

test-raft-concurrent-starts:
	$(GO_ENV) go test raft -run ConcurrentStarts

test-raft-rejoin:
	$(GO_ENV) go test raft -run Rejoin

test-raft-backup:
	$(GO_ENV) go test raft -run Backup

test-raft-persist:
	$(GO_ENV) go test raft -run Persist

fmt-raft:
	gofmt -w src/raft

vet-raft:
	$(GO_ENV) go vet raft

clean:
	rm -rf bin
