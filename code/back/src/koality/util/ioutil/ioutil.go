package ioutil

import (
	"bytes"
	"io"
	"sync"
)

type combinedReader struct {
	buffer    *bytes.Buffer
	readers   []io.Reader
	locker    sync.Locker
	doneChan  chan error
	doneCount int
	writeCond *sync.Cond
}

func CombineReaders(readers ...io.Reader) io.Reader {
	combinedOut := &combinedReader{
		buffer:    new(bytes.Buffer),
		locker:    new(sync.Mutex),
		readers:   readers,
		doneChan:  make(chan error, len(readers)),
		writeCond: sync.NewCond(new(sync.Mutex)),
	}
	combinedOut.combine()
	return combinedOut
}

func (reader *combinedReader) Read(buffer []byte) (int, error) {
	numBytes, err := reader.read(buffer)
	if err != io.EOF {
		return numBytes, err
	}
	if reader.doneCount == len(reader.readers) && numBytes == 0 {
		return 0, io.EOF
	}
	for reader.doneCount < len(reader.readers) {
		// fmt.Println("loop")
		select {
		case err := <-reader.doneChan:
			reader.doneCount++
			if err != io.EOF {
				return numBytes, err
			}
		default:
			if numBytes > 0 {
				return numBytes, nil
			}
			reader.writeCond.L.Lock()
			reader.writeCond.Wait()
			reader.writeCond.L.Unlock()
			return reader.Read(buffer)
		}
	}
	return numBytes, nil
}

func (reader *combinedReader) read(buffer []byte) (int, error) {
	return reader.buffer.Read(buffer)
}

func (reader *combinedReader) combine() {
	for _, subReader := range reader.readers {
		go func(subReader io.Reader) {
			err := reader.syncCopy(subReader)
			reader.writeCond.L.Lock()
			reader.doneChan <- err
			reader.writeCond.Signal()
			reader.writeCond.L.Unlock()
		}(subReader)
	}
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

	reader.writeCond.L.Lock()
	_, err := reader.buffer.Write(buffer)
	reader.writeCond.Signal()
	reader.writeCond.L.Unlock()
	return err
}
