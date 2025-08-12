package engine

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"os"
)

type WALEntryType int

const ENTRY_SIZE int = 8216

const (
	EntryTypeWrite WALEntryType = iota
	EntryTypeCommit
)

type WriteAheadLogEntry struct {
	TxnID   uint64
	Type    WALEntryType
	PageID  PageID
	Offset  uint32
	OldData [PageSize]byte
	NewData [PageSize]byte
}

type WriteAheadLogs struct {
	FilePath string
	File     *os.File
	Writer   *bufio.Writer
}

type WriteAheadLog struct {
	FilePath string
	File     *os.File
	Writer   *bufio.Writer
}

type WALInterface interface {
	Create() error
	Append(entry *WriteAheadLogEntry) error
	Flush() error
	Replay() ([]WriteAheadLogEntry, error)
	Close() error
}

func (wal *WriteAheadLog) Create() error {
	_, err := os.Stat(wal.File.Name())
	if err != nil {
		file, err := os.OpenFile(wal.FilePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		wal.File = file
	}
	return nil
}

func (wal *WriteAheadLog) Append(entry *WriteAheadLogEntry) error {
	err := wal.Create()
	if err != nil {
		return err
	}
	serialized, err := SerializeData(entry)
	if err != nil {
		return err
	}

	_, err = wal.Writer.Write(serialized)
	if err != nil {
		return err
	}

	return nil
}

func (wal *WriteAheadLog) Flush() error {
	panic("Implement this functino")
}

func (wal *WriteAheadLog) Replay() ([]WriteAheadLogEntry, error) {
	var entries []WriteAheadLogEntry

	for {
		serialized_entry := make([]byte, ENTRY_SIZE)
		_, err := wal.File.Read(serialized_entry)
		if err == io.EOF {
			break
		}
		deserialized, err := DeserializeData(serialized_entry, ENTRY_SIZE)
		if err != nil {
			return entries, fmt.Errorf("unable to deserialize data")
		}
		entries = append(entries, *deserialized)
	}

	return entries, nil
}

func (wal *WriteAheadLog) Close() {
	panic("Implement this function")
}

// HELPER FUNCTIONS FOR SERIALIZING AND DESERIALIZING DATA
func SerializeData[T any](data T) ([]byte, error) {
	var bytes_buffer bytes.Buffer
	encoder := gob.NewEncoder(&bytes_buffer)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}
	return bytes_buffer.Bytes(), nil
}

func DeserializeData(data []byte, size int) (*WriteAheadLogEntry, error) {
	entry := &WriteAheadLogEntry{}
	if len(data) != size {
		return entry, fmt.Errorf("invalid entry size")
	}

	buf := bytes.NewReader(data)

	binary.Read(buf, binary.LittleEndian, &entry.TxnID)
	binary.Read(buf, binary.LittleEndian, &entry.Type)
	binary.Read(buf, binary.LittleEndian, &entry.PageID)
	binary.Read(buf, binary.LittleEndian, &entry.Offset)
	buf.Read(entry.OldData[:])
	buf.Read(entry.NewData[:])

	return entry, nil
}
