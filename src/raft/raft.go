package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"labrpc"
	"math/rand"
	"sync"
	"time"
)

// import "bytes"
// import "encoding/gob"

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make().
type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

type LogEntry struct {
	Index   int
	Term    int
	Command interface{}
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex
	peers     []*labrpc.ClientEnd
	persister *Persister
	me        int // index into peers[]

	// Your data here.
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	// role of the server: follower, candidate, or leader
	role string

	// Persistent state on all servers:
	currentTerm int
	votedFor    int
	log         []LogEntry

	// Volatile state on all servers:
	commitIndex int
	lastApplied int

	// Volatile state on leaders:
	nextIndex  []int
	matchIndex []int

	// timer for election timeout
	electionTimer   *time.Timer
	lastResetTime   time.Time
	electionTimeOut time.Duration
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here.
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term = rf.currentTerm
	isleader = rf.role == "leader"
	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
func (rf *Raft) persist() {
	// Your code here.
	// Example:
	// w := new(bytes.Buffer)
	// e := gob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	// Your code here.
	// Example:
	// r := bytes.NewBuffer(data)
	// d := gob.NewDecoder(r)
	// d.Decode(&rf.xxx)
	// d.Decode(&rf.yyy)
}

// example RequestVote RPC arguments structure.
type RequestVoteArgs struct {
	// Your data here.
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

// example RequestVote RPC reply structure.
type RequestVoteReply struct {
	// Your data here.
	Term        int
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	defer rf.persist()
	reply.Success = false
	reply.Term = rf.currentTerm
	if args.Term < rf.currentTerm {
		return
	}
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.votedFor = -1
	}
	reply.Term = rf.currentTerm
	rf.role = "follower"
	rf.resetElectionTimer()

	// check logs
	if args.PrevLogIndex > len(rf.log) {
		// no such log
		return
	}

	if args.PrevLogIndex > 0 && rf.log[args.PrevLogIndex-1].Term != args.PrevLogTerm {
		// leader should reduce nextIndex
		return
	}
	errIdx := -1
	for i := range args.Entries {
		entry := args.Entries[i]
		if entry.Index > len(rf.log) {
			errIdx = i
			break
		}
		index := rf.log[entry.Index-1].Index
		term := rf.log[entry.Index-1].Term
		if entry.Index == index && entry.Term == term {
			continue
		} else {
			errIdx = i
			break
		}
	}
	if errIdx != -1 {
		// log conflict or miss
		entry := args.Entries[errIdx]
		if entry.Index <= len(rf.log) {
			// conflict
			rf.log = rf.log[:entry.Index-1]
		}
		rf.log = append(rf.log, args.Entries[errIdx:]...)
	}

	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = min(args.LeaderCommit, len(rf.log))
	}

	reply.Success = true
}

func (rf *Raft) sendAppendEntries(server int, args AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here.
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply.Term = rf.currentTerm
	reply.VoteGranted = false
	// Reply false if term < currentTerm
	if args.Term < rf.currentTerm {
		return
	}
	// update receiver
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.votedFor = -1
		rf.role = "follower"
		rf.persist()
	}
	// if votedFor is null or candidateId, and candidate’s log is at
	// least as up-to-date as receiver’s log, grant vote
	if (rf.votedFor == -1 || rf.votedFor == args.CandidateId) && rf.isUpToDate(args.LastLogIndex, args.LastLogTerm) {
		reply.Term = rf.currentTerm
		reply.VoteGranted = true
		rf.votedFor = args.CandidateId
		rf.resetElectionTimer()
		rf.persist()
		return
	}
	return
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// returns true if labrpc says the RPC was delivered.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term = rf.currentTerm
	isLeader = rf.role == "leader"
	if !isLeader {
		return index, term, isLeader
	}
	index = len(rf.log) + 1
	newEntry := LogEntry{
		Index:   index,
		Term:    term,
		Command: command,
	}
	rf.log = append(rf.log, newEntry)
	rf.persist()
	return index, term, isLeader
}

// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
func (rf *Raft) Kill() {
	// Your code here, if desired.
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here.
	rf.role = "follower"
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.log = make([]LogEntry, 0)
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))
	for i := range rf.peers {
		rf.nextIndex[i] = 1
		rf.matchIndex[i] = 0
	}

	rf.lastResetTime = time.Now()
	rf.electionTimeOut = randomElectionTimeout()
	rf.electionTimer = time.NewTimer(rf.electionTimeOut)

	go rf.ticker()

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	return rf
}

func (rf *Raft) ticker() {
	for {
		<-rf.electionTimer.C
		rf.mu.Lock()
		if rf.role == "leader" {
			// if it's leader, reset election timer and continue
			rf.resetElectionTimer()
			rf.mu.Unlock()
			continue
		}
		// if it's follower or candidate, start election
		rf.mu.Unlock()
		rf.startElection()
	}
}

func (rf *Raft) startElection() {
	rf.mu.Lock()
	rf.role = "candidate"
	rf.currentTerm++
	rf.votedFor = rf.me
	rf.persist()
	term := rf.currentTerm
	lastLogIndex := 0
	lastLogTerm := 0
	if len(rf.log) > 0 {
		lastLogIndex = rf.log[len(rf.log)-1].Index
		lastLogTerm = rf.log[len(rf.log)-1].Term
	}
	rf.resetElectionTimer()
	rf.mu.Unlock()

	var votes int32 = 1
	for i := range rf.peers {
		if i == rf.me {
			continue
		}
		// send RequestVote RPCs to all other servers concurrently
		go func(server int) {
			args := RequestVoteArgs{
				Term:         term,
				CandidateId:  rf.me,
				LastLogIndex: lastLogIndex,
				LastLogTerm:  lastLogTerm,
			}
			var reply RequestVoteReply
			if rf.sendRequestVote(server, args, &reply) {
				rf.mu.Lock()
				if reply.Term > rf.currentTerm {
					// if RPC reply contains term T > currentTerm:
					// set currentTerm = T, convert to follower
					rf.currentTerm = reply.Term
					rf.role = "follower"
					rf.votedFor = -1
					rf.persist()
					rf.resetElectionTimer()
				} else if reply.VoteGranted {
					// count votes
					if rf.role != "candidate" || rf.currentTerm != term {
						// if it's not candidate or term has changed, ignore the reply
						rf.mu.Unlock()
						return
					}
					votes++
					if votes > int32(len(rf.peers)/2) {
						rf.role = "leader"
						// reinitialize nextIndex and matchIndex for each server
						for i := range rf.peers {
							rf.nextIndex[i] = len(rf.log) + 1
							rf.matchIndex[i] = 0
						}
						rf.resetElectionTimer()
						go rf.startHeartBeatLoop(rf.currentTerm)
					}
				}
				rf.mu.Unlock()
			}
		}(i)
	}
}

func (rf *Raft) broadcastHeartBeat() {
	rf.mu.Lock()
	if rf.role != "leader" {
		rf.mu.Unlock()
		return
	}
	term := rf.currentTerm
	rf.mu.Unlock()
	for i := 0; i < len(rf.peers); i++ {
		if i == rf.me {
			continue
		}
		// send HeartBeat and logs to peer servers
		go func(server int) {
			rf.mu.Lock()
			nextIndex := rf.nextIndex[server]
			prevLogIndex := nextIndex - 1
			prevLogTerm := 0
			if prevLogIndex > 0 {
				prevLogTerm = rf.log[prevLogIndex-1].Term
			}
			entries := make([]LogEntry, len(rf.log[nextIndex-1:]))
			copy(entries, rf.log[nextIndex-1:])
			rf.mu.Unlock()
			args := AppendEntriesArgs{
				Term:         term,
				LeaderId:     rf.me,
				PrevLogIndex: prevLogIndex,
				PrevLogTerm:  prevLogTerm,
				Entries:      entries,
				LeaderCommit: rf.commitIndex,
			}
			var reply AppendEntriesReply
			if rf.sendAppendEntries(server, args, &reply) {
				// now we get reply from other server
				rf.mu.Lock()
				defer rf.mu.Unlock()
				replyTerm := reply.Term
				if replyTerm > rf.currentTerm {
					rf.currentTerm = replyTerm
					rf.role = "follower"
					rf.votedFor = -1
					rf.persist()
					rf.resetElectionTimer()
				}
			}
		}(i)
	}
}

func (rf *Raft) startHeartBeatLoop(term int) {
	for {
		rf.mu.Lock()
		if rf.role != "leader" || rf.currentTerm != term {
			rf.mu.Unlock()
			return
		}
		rf.mu.Unlock()
		rf.broadcastHeartBeat()
		time.Sleep(100 * time.Millisecond)
	}
}

// Helper Functions
// determine if candidate's log is at least as up-to-date as receiver's log
func (rf *Raft) isUpToDate(lastLogIndex int, lastLogTerm int) bool {
	if len(rf.log) == 0 {
		return true
	}
	lastEntry := rf.log[len(rf.log)-1]
	if lastLogTerm != lastEntry.Term {
		return lastLogTerm > lastEntry.Term
	}
	return lastLogIndex >= lastEntry.Index
}

func (rf *Raft) resetElectionTimer() {
	// assume we are already holding the lock
	rf.lastResetTime = time.Now()
	rf.electionTimeOut = randomElectionTimeout()
	if !rf.electionTimer.Stop() {
		select {
		case <-rf.electionTimer.C:
		default:
		}
	}
	rf.electionTimer.Reset(rf.electionTimeOut)
}

func randomElectionTimeout() time.Duration {
	return time.Duration(300+rand.Intn(200)) * time.Millisecond
}
