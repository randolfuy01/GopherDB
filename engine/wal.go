package engine

import (
	"bufio"
	"os"
)

type WALEntryType int

const (
	EntryTypeWrite WALEntryType = iota
	EntryTypeCommit
)

type WriteAheadLogEntry struct {
	TxnID   uint64
	Type    WALEntryType
	PageID  PageID
	Offset  uint32
	OldData []byte
	NewData []byte
}

type WriteAheadLogs struct {
	file   *os.File
	writer *bufio.Writer
}

type WriteAheadLog interface {
	Append(entry *WriteAheadLogEntry) error
	Flush() error
	Replay() ([]WriteAheadLogEntry, error)
	Close() error
}
