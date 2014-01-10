package stagerunner

import (
	"bytes"
	"io"
	"koality/resources"
	"strings"
	"sync"
	"time"
)

type syncWriter struct {
	writer io.Writer
	locker sync.Mutex
}

func (writer *syncWriter) Write(bytes []byte) (int, error) {
	writer.locker.Lock()
	defer writer.locker.Unlock()
	return writer.writer.Write(bytes)
}

type consoleTextWriter struct {
	stagesUpdateHandler resources.StagesUpdateHandler
	stageRunId          uint64
	buffer              bytes.Buffer
	locker              sync.Mutex
	closeChan           chan bool
	lastLine            string
	lastLineNumber      uint64
}

func newConsoleTextWriter(stagesUpdateHandler resources.StagesUpdateHandler, stageRunId uint64) *consoleTextWriter {
	var buffer bytes.Buffer
	var locker sync.Mutex
	writer := &consoleTextWriter{stagesUpdateHandler, stageRunId, buffer, locker, make(chan bool, 1), "", 1}
	go writer.flushOnTick()
	return writer
}

func (writer *consoleTextWriter) flushOnTick() {
	ticker := time.NewTicker(250 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			writer.flush()
		case <-writer.closeChan:
			ticker.Stop()
			writer.flush()
			return
		}
	}
}

func (writer *consoleTextWriter) Write(bytes []byte) (int, error) {
	writer.locker.Lock()
	defer writer.locker.Unlock()

	numBytes, err := writer.buffer.Write(bytes)

	return numBytes, err
}

func (writer *consoleTextWriter) Close() error {
	writer.closeChan <- true
	close(writer.closeChan)
	return nil
}

func (writer *consoleTextWriter) flush() error {
	writer.locker.Lock()
	defer writer.locker.Unlock()

	if writer.buffer.Len() == 0 {
		return nil
	}

	lines := strings.Split(writer.buffer.String(), "\n")
	linesMap := make(map[uint64]string, len(lines))

	firstLineEmpty := lines[0] == ""

	lines[0] = writer.lastLine + lines[0]

	for index, line := range lines {
		linesMap[writer.lastLineNumber+uint64(index)] = line
	}

	if firstLineEmpty {
		delete(linesMap, writer.lastLineNumber)
	}

	if strings.TrimSpace(lines[len(lines)-1]) == "" {
		delete(linesMap, writer.lastLineNumber+uint64(len(lines)-1))
	}

	writer.buffer.Reset()

	writer.lastLine = lines[len(lines)-1]
	writer.lastLineNumber = writer.lastLineNumber + uint64(len(lines)) - 1

	err := writer.stagesUpdateHandler.AddConsoleLines(writer.stageRunId, linesMap)
	return err
}
