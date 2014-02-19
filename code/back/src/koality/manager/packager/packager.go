package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/pwaller/goupx/hemfix"
	"io/ioutil"
	"koality/shell"
	"koality/util/pathtranslator"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var compressFlag = flag.Bool("compress", true, "compress the Koality package")
var upxFlag = flag.Bool("upx", false, "compress Koality go binaries with upx (not suggested)")

var requiredLibraries = []string{"curl", "libpcre3-dev", "upx"}

var expectedArtifacts = []string{"koality", "koalityRunner", "getXunitResults", "exportPaths", "sshwrapper",
	"installer", "authorizedKeys", "restrictedShell"}

var expectedFiles = []string{"code", "postgres", "nginx", "dependencies", ".metadata", "misc"}

func main() {
	flag.Parse()
	installDependenciesCommand := exec.Command("bash", "-c", string(shell.Sudo(
		shell.Commandf("apt-get install -y %s", strings.Join(requiredLibraries, " ")),
	)))
	installDependenciesStdout, installDependenciesStderr := new(bytes.Buffer), new(bytes.Buffer)
	installDependenciesCommand.Stdout, installDependenciesCommand.Stderr = installDependenciesStdout, installDependenciesStderr
	if err := installDependenciesCommand.Run(); err != nil {
		panic(err)
	}

	outputDirectory := filepath.Join("/", "tmp", "out")
	if *compressFlag {
		defer os.RemoveAll(outputDirectory)
	}
	if err := packageKoality(outputDirectory); err != nil {
		panic(err)
	}
	fmt.Printf("Koality packaged successfully to %s\n", outputDirectory)

	tgzPath := filepath.Join("/", "tmp", "out.tgz")
	compressCommand := exec.Command("tar", "-czf", tgzPath, filepath.Base(outputDirectory))
	compressCommand.Dir = filepath.Dir(tgzPath)
	compressStdout, compressStderr := new(bytes.Buffer), new(bytes.Buffer)
	compressCommand.Stdout, compressCommand.Stderr = compressStdout, compressStderr

	if *compressFlag {
		fmt.Printf("Compressing Koality from %s to %s...\n", outputDirectory, tgzPath)
		if err := compressCommand.Run(); err != nil {
			panic(err)
		}
		fmt.Printf("Koality compressed to %s\n", tgzPath)
	}
}

func packageKoality(outputDirectory string) error {
	var err error

	installCommand := exec.Command("go", "install", "koality/...")
	installStdout, installStderr := new(bytes.Buffer), new(bytes.Buffer)
	installCommand.Stdout, installCommand.Stderr = installStdout, installStderr

	fmt.Println("Installing koality binaries...")
	if err = installCommand.Run(); err != nil {
		return fmt.Errorf("Failed to install koality binaries\nStdout: %s\nStderr: %s\nError: %v", installStdout.String(), installStderr.String(), err)
	}
	fmt.Println("Koality binaries installed")

	ownDirectory, err := pathtranslator.TranslatePath(".")
	if err != nil {
		return err
	}

	if _, err = os.Stat(outputDirectory); err == nil {
		fmt.Printf("Removing old output directory at %s...\n", outputDirectory)
		if err = os.RemoveAll(outputDirectory); err != nil {
			return err
		}
		fmt.Printf("Removed old output directory at %s\n", outputDirectory)
	}

	copyCommand := exec.Command("cp", "-r", ownDirectory, outputDirectory)
	copyStdout, copyStderr := new(bytes.Buffer), new(bytes.Buffer)
	copyCommand.Stdout, copyCommand.Stderr = copyStdout, copyStderr

	fmt.Printf("Copying repository from %s to %s...\n", ownDirectory, outputDirectory)
	if err = copyCommand.Run(); err != nil {
		return fmt.Errorf("Failed to copy the repository from %s to %s\nStdout: %s\nStderr: %s\nError: %v", ownDirectory, outputDirectory, copyStdout.String(), copyStderr.String(), err)
	}
	fmt.Printf("Copied repository from %s to %s\n", ownDirectory, outputDirectory)

	for _, artifact := range expectedArtifacts {
		artifactPath := filepath.Join(outputDirectory, pathtranslator.BinaryPath(artifact))
		if err = pathtranslator.CheckExecutable(artifactPath); err != nil {
			return fmt.Errorf("Could not find expected artifact: %s\nError: %v", artifact, err)
		}

		if *upxFlag {
			if err = runUpx(artifactPath); err != nil {
				return err
			}
		}
	}

	expectedArtifactsSet := make(map[string]bool, len(expectedArtifacts))
	for _, artifact := range expectedArtifacts {
		expectedArtifactsSet[artifact] = true
	}

	artifactsDir := filepath.Clean(filepath.Join(outputDirectory, pathtranslator.BinaryPath(".")))
	artifacts, err := ioutil.ReadDir(artifactsDir)
	if err != nil {
		return fmt.Errorf("Unable to list artifacts in directory: %s\nError: %v", artifactsDir, err)
	}

	for _, artifact := range artifacts {
		if _, expected := expectedArtifactsSet[artifact.Name()]; !expected {
			unexpectedArtifact := filepath.Join(artifactsDir, artifact.Name())
			fmt.Printf("Removing unexpected artifact at %s...\n", unexpectedArtifact)
			if err = os.Remove(unexpectedArtifact); err != nil {
				return fmt.Errorf("Unable to remove unexpected artifact at %s\nError: %v", unexpectedArtifact, err)
			}
			fmt.Printf("Removed unexpected artifact at %s\n", unexpectedArtifact)
		}
	}

	srcDir := filepath.Join(outputDirectory, "code", "back", "src")
	fmt.Printf("Removing src directory at %s...\n", srcDir)
	if err = os.RemoveAll(srcDir); err != nil {
		return fmt.Errorf("Failed to remove src directory at %s\nError: %v", srcDir, err)
	}
	fmt.Printf("Removed src directory at %s\n", srcDir)

	pkgDir := filepath.Join(outputDirectory, "code", "back", "pkg")
	fmt.Printf("Removing pkg directory at %s...\n", pkgDir)
	if err = os.RemoveAll(pkgDir); err != nil {
		return fmt.Errorf("Failed to remove pkg directory at %s\nError: %v", pkgDir, err)
	}
	fmt.Printf("Removed pkg directory at %s\n", pkgDir)

	dependenciesDirectory := filepath.Join(outputDirectory, "dependencies")
	if err = os.Mkdir(dependenciesDirectory, 0755); err != nil {
		return fmt.Errorf("Failed to create dependencies directory at %s\nError: %v", dependenciesDirectory, err)
	}

	curlCommand := exec.Command("bash", "-c", string(shell.Pipe(
		"curl http://nginx.org/download/nginx-1.4.4.tar.gz",
		"tar -xzf -",
	)))
	curlCommand.Dir = dependenciesDirectory
	curlStdout, curlStderr := new(bytes.Buffer), new(bytes.Buffer)
	curlCommand.Stdout, curlCommand.Stderr = curlStdout, curlStderr
	fmt.Println("Downloading nginx...")
	if err = curlCommand.Run(); err != nil {
		return fmt.Errorf("Failed to download nginx\nStdout: %s\nStderr: %s\nError: %v", curlStdout.String(), curlStderr.String(), err)
	}
	fmt.Println("Nginx downloaded")

	configureCommand := exec.Command(filepath.Join(dependenciesDirectory, "nginx-1.4.4", "configure"), "--prefix=/etc/koality/current/nginx", "--with-http_ssl_module")
	configureCommand.Dir = filepath.Join(dependenciesDirectory, "nginx-1.4.4")
	configureStdout, configureStderr := new(bytes.Buffer), new(bytes.Buffer)
	configureCommand.Stdout, configureCommand.Stderr = configureStdout, configureStderr
	fmt.Println("Configuring nginx for compilation...")
	if err = configureCommand.Run(); err != nil {
		return fmt.Errorf("Failed to configure nginx\nStdout: %s\nStderr: %s\nError: %v", configureStdout.String(), configureStderr.String(), err)
	}
	fmt.Println("Nginx configured")

	makeCommand := exec.Command("make")
	makeCommand.Dir = filepath.Join(dependenciesDirectory, "nginx-1.4.4")
	makeStdout, makeStderr := new(bytes.Buffer), new(bytes.Buffer)
	makeCommand.Stdout, makeCommand.Stderr = makeStdout, makeStderr
	fmt.Println("Compiling nginx...")
	if err = makeCommand.Run(); err != nil {
		return fmt.Errorf("Failed to make nginx\nStdout: %s\nStderr: %s\nError: %v", makeStdout.String(), makeStderr.String(), err)
	}
	fmt.Println("Compiled nginx")

	nginxOutputPath := filepath.Join(outputDirectory, "nginx")

	nginxOriginalPath := filepath.Join(dependenciesDirectory, "nginx-1.4.4", "objs", "nginx")
	fmt.Println("Moving nginx...")
	if err = os.Rename(nginxOriginalPath, filepath.Join(nginxOutputPath, "nginx")); err != nil {
		return fmt.Errorf("Failed to move nginx binary from %s to %s\nError: %v", nginxOriginalPath, nginxOutputPath, err)
	}
	fmt.Printf("Moved nginx to %s\n", nginxOutputPath)

	nginxMimeTypesOriginalPath := filepath.Join(dependenciesDirectory, "nginx-1.4.4", "conf", "mime.types")
	fmt.Println("Moving nginx mime.types...")
	if err = os.Rename(nginxMimeTypesOriginalPath, filepath.Join(nginxOutputPath, "mime.types")); err != nil {
		return fmt.Errorf("Failed to move nginx mime.types from %s to %s\nError: %v", nginxMimeTypesOriginalPath, nginxOutputPath, err)
	}
	fmt.Printf("Moved nginx mime.types to %s\n", nginxOutputPath)

	if err = os.RemoveAll(filepath.Join(dependenciesDirectory, "nginx-1.4.4")); err != nil {
		return fmt.Errorf("Failed to clean up nginx directory\nError: %v", err)
	}

	nginxConfPath := filepath.Join(nginxOutputPath, "nginx.conf")
	if err = unlinkFile(nginxConfPath); err != nil {
		return err
	}

	fmt.Println("Modifying nginx.conf...")
	nginxConfBytes, err := ioutil.ReadFile(nginxConfPath)
	if err != nil {
		return fmt.Errorf("Failed to read %s\nError: %v", nginxConfPath, err)
	}

	nginxConfContents := strings.Replace(string(nginxConfBytes), "/etc/nginx", "/etc/koality/current/nginx", -1) + "\ndaemon off;"

	if err = ioutil.WriteFile(nginxConfPath, []byte(nginxConfContents), 0644); err != nil {
		return fmt.Errorf("Failed to modify nginx.conf\nError: %v", err)
	}

	curlCommand = exec.Command("bash", "-c", string(shell.Pipe(
		"curl http://openbsd.mirrors.pair.com/OpenSSH/portable/openssh-6.0p1.tar.gz",
		"tar -xzf -",
	)))
	curlCommand.Dir = dependenciesDirectory
	curlStdout, curlStderr = new(bytes.Buffer), new(bytes.Buffer)
	curlCommand.Stdout, curlCommand.Stderr = curlStdout, curlStderr
	fmt.Println("Downloading OpenSSH...")
	if err = curlCommand.Run(); err != nil {
		return fmt.Errorf("Failed to download OpenSSH\nStdout: %s\nStderr: %s\nError: %v", curlStdout.String(), curlStderr.String(), err)
	}
	fmt.Println("OpenSSH downloaded")

	cloneCommand := exec.Command("git", "clone", "--depth", "1", "git://github.com/LessThanThreeLabs/openssh-for-git.git")
	cloneCommand.Dir = dependenciesDirectory
	cloneStdout, cloneStderr := new(bytes.Buffer), new(bytes.Buffer)
	cloneCommand.Stdout, cloneCommand.Stderr = cloneStdout, cloneStderr
	fmt.Println("Downloading the OpenSSH patch...")
	if err = cloneCommand.Run(); err != nil {
		return fmt.Errorf("Failed to download the OpenSSH patch\nStdout: %s\nStderr: %s\nError: %v", cloneStdout.String(), cloneStderr.String(), err)
	}
	fmt.Println("OpenSSH patch downloaded")

	patchCommand := exec.Command("patch", "-p1")
	patchCommand.Dir = filepath.Join(dependenciesDirectory, "openssh-6.0p1")
	patchPath := filepath.Join(dependenciesDirectory, "openssh-for-git", "openssh-6.0p1-authorized-keys-script.diff")
	patchFile, err := os.Open(patchPath)
	if err != nil {
		return fmt.Errorf("Could not find the OpenSSH patch at %s\nError: %v", patchPath, err)
	}
	defer patchFile.Close()
	patchCommand.Stdin = patchFile

	patchStdout, patchStderr := new(bytes.Buffer), new(bytes.Buffer)
	patchCommand.Stdout, patchCommand.Stderr = patchStdout, patchStderr
	fmt.Println("Applying the OpenSSH patch...")
	if err = patchCommand.Run(); err != nil {
		return fmt.Errorf("Failed to apply the OpenSSH patch\nStdout: %s\nStderr: %s\nError: %v", patchStdout.String(), patchStderr.String(), err)
	}
	fmt.Println("OpenSSH patched")

	if err = os.RemoveAll(filepath.Join(dependenciesDirectory, "openssh-for-git")); err != nil {
		return fmt.Errorf("Failed to remove the OpenSSH patch repository at %s\nError: %v", filepath.Join(dependenciesDirectory, "openssh-for-git"), err)
	}

	configureCommand = exec.Command(filepath.Join(dependenciesDirectory, "openssh-6.0p1", "configure"))
	configureCommand.Dir = filepath.Join(dependenciesDirectory, "openssh-6.0p1")
	configureStdout, configureStderr = new(bytes.Buffer), new(bytes.Buffer)
	configureCommand.Stdout, configureCommand.Stderr = configureStdout, configureStderr
	fmt.Println("Configuring OpenSSH for compilation...")
	if err = configureCommand.Run(); err != nil {
		return fmt.Errorf("Failed to configure OpenSSH\nStdout: %s\nStderr: %s\nError: %v", configureStdout.String(), configureStderr.String(), err)
	}
	fmt.Println("OpenSSH configured")

	makeCommand = exec.Command("make", "-j", "4")
	makeCommand.Dir = filepath.Join(dependenciesDirectory, "openssh-6.0p1")
	makeStdout, makeStderr = new(bytes.Buffer), new(bytes.Buffer)
	makeCommand.Stdout, makeCommand.Stderr = makeStdout, makeStderr
	fmt.Println("Compiling OpenSSH...")
	if err = makeCommand.Run(); err != nil {
		return fmt.Errorf("Failed to make OpenSSH\nStdout: %s\nStderr: %s\nError: %v", makeStdout.String(), makeStderr.String(), err)
	}
	fmt.Println("Compiled OpenSSH")

	fmt.Println("Modifying OpenSSH configuration files...")
	sshdConfigFile, err := os.OpenFile(filepath.Join(dependenciesDirectory, "openssh-6.0p1", "sshd_config"), os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("Failed to read the sshd_config file\nError: %v", err)
	}
	defer sshdConfigFile.Close()

	if _, err = fmt.Fprintf(sshdConfigFile, "AuthorizedKeysScript %s", filepath.Join("/", "etc", "koality", "current", pathtranslator.BinaryPath("authorizedKeys"))); err != nil {
		return fmt.Errorf("Failed to add AuthorizedKeysScript option to the ssh_config file\nError: %v", err)
	}

	for _, file := range expectedFiles {
		expectedFilePath := filepath.Join(outputDirectory, file)
		if _, err = os.Stat(expectedFilePath); err != nil {
			return fmt.Errorf("Could not find expected file: %s\nError: %v", file, err)
		}
	}

	expectedFilesSet := make(map[string]bool, len(expectedFiles))
	for _, directory := range expectedFiles {
		expectedFilesSet[directory] = true
	}

	files, err := ioutil.ReadDir(outputDirectory)
	if err != nil {
		return fmt.Errorf("Unable to list the contents of the output directory: %s\nError: %v", outputDirectory, err)
	}

	for _, file := range files {
		if _, expected := expectedFilesSet[file.Name()]; !expected {
			unexpectedFile := filepath.Join(outputDirectory, file.Name())
			fmt.Printf("Removing unexpected file at %s...\n", unexpectedFile)
			if err = os.RemoveAll(unexpectedFile); err != nil {
				return fmt.Errorf("Unable to remove unexpected file at %s\nError: %v", unexpectedFile, err)
			}
			fmt.Printf("Removed unexpected file at %s\n", unexpectedFile)
		}
	}

	chefDir := filepath.Join(outputDirectory, "chef")
	fmt.Printf("Removing chef directory at %s...\n", chefDir)
	if err = os.RemoveAll(chefDir); err != nil {
		return fmt.Errorf("Failed to remove chef directory at %s\nError: %v", chefDir, err)
	}
	fmt.Printf("Removed chef directory at %s\n", chefDir)

	return nil
}

func unlinkFile(filePath string) error {
	linkInfo, err := os.Lstat(filePath)
	if err != nil {
		return fmt.Errorf("Unable to check link status of %s\nError: %v", filePath, err)
	}

	if linkInfo.Mode()&os.ModeSymlink != 0 {
		linkPath, err := os.Readlink(filePath)
		if err != nil {
			return fmt.Errorf("Unable to read link at %s\nError: %v", filePath, err)
		}

		if err = os.Remove(filePath); err != nil {
			return fmt.Errorf("Failed to remove link at %s\nError: %v", filePath, err)
		}
		if !filepath.IsAbs(linkPath) {
			linkPath = filepath.Join(filepath.Dir(filePath), linkPath)
		}
		copyCommand := exec.Command("cp", linkPath, filePath)
		copyStdout, copyStderr := new(bytes.Buffer), new(bytes.Buffer)
		copyCommand.Stdout, copyCommand.Stderr = copyStdout, copyStderr
		if err = copyCommand.Run(); err != nil {
			return fmt.Errorf("Failed to copy file from %s to %s\nStdout: %s\nStderr: %s\nError: %v", linkPath, filePath, copyStdout.String(), copyStderr.String(), err)
		}
	}

	return nil
}

func runUpx(binaryPath string) error {
	var err error

	fmt.Printf("Compressing binary %s...\n", binaryPath)

	if err = hemfix.FixFile(binaryPath); err != nil {
		return fmt.Errorf("Failed to fix file %s for upx\nError: %v", binaryPath, err)
	}

	stripCommand := exec.Command("strip", "-s", binaryPath)
	if err = stripCommand.Run(); err != nil {
		return fmt.Errorf("Failed to strip file %s\nError: %v", binaryPath, err)
	}

	upxCommand := exec.Command("bash", "-c", string(shell.And("upx", "upx -d", "upx")), binaryPath)
	if err = upxCommand.Run(); err != nil {
		return fmt.Errorf("Failed to run upx on file %s\nError: %v", upxCommand, err)
	}

	fmt.Printf("Compressed binary %s...\n", binaryPath)
	return nil
}
