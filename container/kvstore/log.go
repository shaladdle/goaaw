package kvstore

import (
	"os"
	"path"
)

type walog struct {
}

type logEntry struct {
    id uint64
    data kvpair
}

func (p *kvstore) logFilePath() string {
	return path.Join(p.path, "-log")
}

func (p *kvstore) loadLog() (*walog, error) {
	f, err := os.Open(p.logFilePath())
	if err != nil {
		return nil, err
	}
	defer f.Close()

    return nil, nil
}

type logManWrite kvpair
type logManAppended uint64
type logManShutdown chan bool

type appWrite logManWrite
type appShutdown chan uint64
type appDone uint64

type cmtSetLimit uint64
type cmtCommitDone struct {}
type cmtShutdown chan bool

func (p *kvstore) logCommitter(msgs chan interface{}) {
    commit := func(limit uint64, done chan<- interface{}) {
        // commit should write logEntries to the actual data file up until the
        // limit transaction id
        done <- cmtCommitDone{}
    }

    var last, limit uint64
    committing := false
	for {
		switch msg := (<-msgs).(type) {
        case cmtSetLimit:
            limit = uint64(msg)
            if limit > last && !committing {
                go commit(limit, msgs)
            }
        case cmtCommitDone:
            if limit > last {
                go commit(limit, msgs)
            }
		}
	}
}

func (p *kvstore) logAppender(msgs chan interface{}, manMsgs chan<- interface{}) {
    appender := func(pairs []kvpair, amsgs chan<- interface{}) {
        // append the pairs to the end of the file along with their id
        amsgs <- 0
    }

    f, err := os.Open(p.logFilePath())
    if err != nil {
        panic(err)
    }
    defer f.Close()

    var last uint64
    // Walk the file until we get to the end, keeping track of what the last
    // sequence number is. Set last here.

    pendingWrites := []kvpair{}
    appending := false
	for {
		switch msg := (<-msgs).(type) {
        case appWrite:
            if appending {
                pendingWrites = append(pendingWrites, kvpair(msg))
            } else {
                go appender(append(pendingWrites, kvpair(msg)), msgs)
                appending = true

                if len(pendingWrites) > 0 {
                    pendingWrites = []kvpair{}
                }
            }
        case appDone:
            last = uint64(msg)

            // have to do this async or there's a deadlock
            go func () { manMsgs <- logManAppended(last) } ()

            if len(pendingWrites) > 0 {
                go appender(pendingWrites, msgs)
                pendingWrites = []kvpair{}
            } else {
                appending = false
            }
        case appShutdown:
            if appending {
                // wait for appender to finish
                msg <- (<-msgs).(uint64)
            }
		}
	}
}

func (p *kvstore) logManager(msgs chan interface{}) {
	_, err := p.loadLog()
	if err != nil {
		panic(err)
	}

    appender := func(manMsgs chan<- interface{}) {
        // append new data to the end of the log
        manMsgs <- logManAppended(100)
    }

    appending := false
    committing := false

	for {
		switch msg := (<-msgs).(type) {
        case logManAppended:
            // also need to notify kvstore at large that it's ok to insert
            // this into the in-mem cache?
		case logManShutdown:
            break
		}
	}
}

func (p *kvstore) startLogManager() chan<- interface{} {
	msgs := make(chan interface{})
	go p.logManager(msgs)
	return msgs
}
