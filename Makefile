.PHONY: test test-raft test-raft-election test-raft-log test-raft-persist fmt fmt-raft vet vet-raft clean

GO_ENV := GO111MODULE=off GOPATH=$(CURDIR)

test: test-raft

test-raft:
	$(GO_ENV) go test raft

test-raft-election:
	$(GO_ENV) go test raft -run 'TestInitialElection|TestReElection'

test-raft-log:
	$(GO_ENV) go test raft -run 'TestBasicAgree|TestFailAgree|TestFailNoAgree|TestConcurrentStarts|TestRejoin|TestBackup|TestCount'

test-raft-persist:
	$(GO_ENV) go test raft -run 'TestPersist'

fmt: fmt-raft

fmt-raft:
	gofmt -w src/raft

vet: vet-raft

vet-raft:
	$(GO_ENV) go vet raft

clean:
	rm -rf bin
