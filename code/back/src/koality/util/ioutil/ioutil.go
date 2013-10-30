package ioutil

import (
	"bytes"
	"io"
	"sync"
)

type combinedReader struct {
	buffer    *bytes.Buffer
	readers   []io.Reader
	locker    sync.Mutex
	exitError error
}

func CombineReaders(readers ...io.Reader) io.Reader {
	combinedOut := &combinedReader{
		buffer:  new(bytes.Buffer),
		locker:  new(sync.Mutex),
		readers: readers,
	}
	combinedOut.combine()
	return combinedOut
}

func (reader *combinedReader) Read(buffer []byte) (n int, err error) {
	numBytes, err := reader.Buffer.read(buffer)
	if err != io.EOF {
		return numBytes, err
	}
	// It is possible to return (0, nil), which is defined as a no-op, when the readers haven't been fully read
	return numBytes, reader.exitError
}

func (reader *combinedReader) combine() {
	doneChan := make(chan error, len(reader.readers))
	for _, subReader := range reader.readers {
		go func(subReader io.Reader) {
			doneChan <- reader.syncCopy(subReader)
		}(subReader)
	}

	go reader.trackDone(doneChan)
	return combinedOut
}

func (reader *combinedReader) trackDone(doneChan chan<- error) {
	doneCount := 0
	for doneCount < len(reader.readers) {
		err := <-doneChan
		doneCount++
		if err != io.EOF {
			reader.exitError = err
			return
		}
	}
	reader.exitError = io.EOF
}

func (reader *combinedReader) syncCopy(subReader io.Reader) error {
	buffer := make([]byte, 1024)
	for {
		numBytes, err := subReader.Read(buffer)
		if err != nil {
			return err
		}
		if numBytes > 0 {
			err = reader.syncWrite(buffer[:numBytes])
			if err != nil {
				return err
			}
		}
	}
}

func (reader *combinedReader) syncWrite(buffer []byte) error {
	reader.locker.Lock()
	defer reader.locker.Unlock()

	_, err := reader.Write(buffer)
	return err
}
