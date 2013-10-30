package shell

type FileCopier interface {
	FileCopy(sourcePath, destPath string) Executable
}
