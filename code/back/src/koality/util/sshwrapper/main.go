package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
)

func main() {
	file, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}

	filename := file.Name()
	if _, err = file.WriteString(os.Getenv("PRIVATE_KEY")); err != nil {
		panic(err)
	}

	if err = file.Chmod(0600); err != nil {
		panic(err)
	} else if err = file.Close(); err != nil {
		panic(err)
	}

	bashPath := path.Join("/bin", "bash")
	if err = exec.Command(bashPath, "-c", "sleep 120; rm "+filename).Start(); err != nil {
		panic(err)
	}

	sshPath := path.Join("/", "usr", "bin", "ssh")
	// sshPath, err := exec.LookPath("ssh")
	// if err != nil {
	// 	panic(err)
	// }

	if sshPath, err = filepath.Abs(sshPath); err != nil {
		panic(err)
	}

	argv := append(append([]string{sshPath}, "-oStrictHostKeyChecking=no", "-oUserKnownHostsFile=/dev/null", "-i", filename), os.Args[1:]...)
	panic(syscall.Exec(argv[0], argv, nil))
}
