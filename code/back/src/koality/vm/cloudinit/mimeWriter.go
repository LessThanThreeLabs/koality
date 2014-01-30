package cloudinit

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"
)

type mimeMultipartWriter struct {
	writer          io.Writer
	multipartWriter *multipart.Writer
	wroteHeader     bool
}

func NewMimeMultipartWriter(writer io.Writer) *mimeMultipartWriter {
	return &mimeMultipartWriter{
		writer:          writer,
		multipartWriter: multipart.NewWriter(writer),
	}
}

func (writer *mimeMultipartWriter) WriteHeader() error {
	if writer.wroteHeader {
		return fmt.Errorf("Already wrote header")
	}

	writer.wroteHeader = true
	_, err := fmt.Fprintf(writer.writer, "Content-Type: multipart/mixed; boundary=\"%s\"\nMIME-Version: 1.0\n\n", writer.multipartWriter.Boundary())
	return err
}

func (writer *mimeMultipartWriter) WriteMimePart(name, contents string) error {
	if !writer.wroteHeader {
		err := writer.WriteHeader()
		if err != nil {
			return err
		}
	}
	mimeHeader := make(textproto.MIMEHeader)
	mimeHeader.Set("Content-Type", cloudInitMimeType(contents))
	mimeHeader.Set("Content-Disposition", fmt.Sprintf("form-data; name=\"%s\"", name))
	partWriter, err := writer.multipartWriter.CreatePart(mimeHeader)
	if err != nil {
		return err
	}
	_, err = io.WriteString(partWriter, contents)
	return err
}

func (writer *mimeMultipartWriter) Close() error {
	return writer.multipartWriter.Close()
}

func cloudInitMimeType(contents string) string {
	startsWithMapping := map[string]string{
		"#include":              "text/x-include-url",
		"#!":                    "text/x-shellscript",
		"#cloud-boothook":       "text/cloud-boothook",
		"#cloud-config":         "text/cloud-config",
		"#cloud-config-archive": "text/cloud-config-archive",
		"#upstart-job":          "text/upstart-job",
		"#part-handler":         "text/part-handler",
	}
	for prefix, contentType := range startsWithMapping {
		if strings.HasPrefix(contents, prefix) {
			return contentType
		}
	}
	return "text/plain"
}
