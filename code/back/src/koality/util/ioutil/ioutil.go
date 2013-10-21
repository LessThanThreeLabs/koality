package ioutil

import (
	"bytes"
	"io"
	"sync"
)

func CombineReaders(readers ...io.Reader) io.Reader {
	combinedOut := new(bytes.Buffer)

	mutex := new(sync.Mutex)
	for _, reader := range readers {
		go syncCopy(reader, combinedOut, mutex)
	}
	return combinedOut
}

func syncCopy(reader io.Reader, writer io.Writer, locker sync.Locker) {
	buffer := make([]byte, 1024)
	for {
		numBytes, err := reader.Read(buffer)
		if numBytes == 0 || err == io.EOF {
			return
		} else if err != nil {
			panic(err)
		}
		syncWrite(writer, buffer[:numBytes], locker)
	}
}

func syncWrite(writer io.Writer, buffer []byte, locker sync.Locker) {
	locker.Lock()
	defer locker.Unlock()

	_, err := writer.Write(buffer)
	if err != nil {
		panic(err)
	}
}
