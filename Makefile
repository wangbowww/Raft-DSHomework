.PHONY: test test-raft test-raft-election test-raft-log test-raft-persist \
	test-raft-initial-election test-raft-re-election test-raft-basic-agree \
	test-raft-fail-agree test-raft-fail-no-agree test-raft-concurrent-starts \
	test-raft-rejoin test-raft-backup test-raft-count test-raft-persist-1 \
	test-raft-persist-2 test-raft-persist-3 test-raft-figure-8 \
	test-raft-unreliable-agree test-raft-figure-8-unreliable \
	test-raft-reliable-churn test-raft-unreliable-churn \
	fmt fmt-raft vet vet-raft clean

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

test-raft-initial-election:
	$(GO_ENV) go test raft -run '^TestInitialElection$$'

test-raft-re-election:
	$(GO_ENV) go test raft -run '^TestReElection$$'

test-raft-basic-agree:
	$(GO_ENV) go test raft -run '^TestBasicAgree$$'

test-raft-fail-agree:
	$(GO_ENV) go test raft -run '^TestFailAgree$$'

test-raft-fail-no-agree:
	$(GO_ENV) go test raft -run '^TestFailNoAgree$$'

test-raft-concurrent-starts:
	$(GO_ENV) go test raft -run '^TestConcurrentStarts$$'

test-raft-rejoin:
	$(GO_ENV) go test raft -run '^TestRejoin$$'

test-raft-backup:
	$(GO_ENV) go test raft -run '^TestBackup$$'

test-raft-count:
	$(GO_ENV) go test raft -run '^TestCount$$'

test-raft-persist-1:
	$(GO_ENV) go test raft -run '^TestPersist1$$'

test-raft-persist-2:
	$(GO_ENV) go test raft -run '^TestPersist2$$'

test-raft-persist-3:
	$(GO_ENV) go test raft -run '^TestPersist3$$'

test-raft-figure-8:
	$(GO_ENV) go test raft -run '^TestFigure8$$'

test-raft-unreliable-agree:
	$(GO_ENV) go test raft -run '^TestUnreliableAgree$$'

test-raft-figure-8-unreliable:
	$(GO_ENV) go test raft -run '^TestFigure8Unreliable$$'

test-raft-reliable-churn:
	$(GO_ENV) go test raft -run '^TestReliableChurn$$'

test-raft-unreliable-churn:
	$(GO_ENV) go test raft -run '^TestUnreliableChurn$$'

fmt: fmt-raft

fmt-raft:
	gofmt -w src/raft

vet: vet-raft

vet-raft:
	$(GO_ENV) go vet raft

clean:
	rm -rf bin
