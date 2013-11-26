package provision

import (
	"fmt"
	"koality/shell"
	"path/filepath"
	"strings"
)

const baseDirectory = "~"

func rcAppendCommand(contents string) shell.Command {
	return shell.Silent(shell.Append(shell.Command(fmt.Sprintf("echo %s", shell.Quote(contents))), shell.Command(rcPath()), true))
}

func rcPath() string {
	return filepath.Join(baseDirectory, ".koalityrc")
}

func foundExistingCommand(language string, location string) shell.Command {
	return shell.Command(fmt.Sprintf("printf \"%sFound existing %s install at: %s%s\\n\"",
		shell.AnsiFormat(shell.AnsiFgGreen, shell.AnsiBold),
		language,
		location,
		shell.AnsiFormat(shell.AnsiReset)),
	)
}

type PackageManager struct {
	name        string
	installArgs string
	updateArgs  string
}

func (pm *PackageManager) installCommand(packages []string) shell.Command {
	install := shell.Advertised(shell.Command(fmt.Sprintf("%s %s %s", pm.name, pm.installArgs, strings.Join(packages, " "))))
	return shell.And(
		shell.Silent(shell.Command(fmt.Sprintf("which %s", pm.name))),
		shell.Or(
			install,
			shell.Chain(
				shell.Advertised(shell.Command(fmt.Sprintf("%s %s", pm.name, pm.updateArgs))),
				install,
			),
		),
	)
}

func installPackages(packageStrings []string, platformSpecific map[string]([]string)) shell.Command {
	var installAttempts []shell.Command

	packageManagers := map[string]PackageManager{
		"apt-get": PackageManager{"apt-get", "install -y --force-yes", "update -y"},
		"yum":     PackageManager{"yum", "install -y", "check-update -y"},
		"Zypper":  PackageManager{"zypper", "install -y", "refresh -y"},
	}

	var (
		supportedManagers []string
		installAttempt    shell.Command
	)

	for name, packageManager := range packageManagers {
		if specificCommands, ok := platformSpecific[name]; ok {
			installAttempt = packageManager.installCommand(specificCommands)
		} else {
			installAttempt = packageManager.installCommand(packageStrings)
		}
		supportedManagers = append(supportedManagers, name)
		installAttempts = append(installAttempts, installAttempt)
	}

	errorMessage := fmt.Sprintf("%sCould not find a package manager to install: %s.\\nSupports: %s%s",
		shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold),
		packageStrings,
		supportedManagers,
		shell.AnsiFormat(shell.AnsiReset),
	)
	errorCommand := shell.And(
		shell.Command(fmt.Sprintf("echo -e %s", shell.Quote(errorMessage))),
		shell.Command("false"),
	)

	return shell.Or(append(installAttempts, errorCommand)...)
}

//TODO(akostov) debug
func ParseLanguages(languageConfig map[string]string) (provisionCommand shell.Command, err error) {
	languageDispatcher := map[string]func(string) (shell.Command, shell.Command){
		"python": parsePython,
		"ruby":   parseRuby,
		"nodejs": parseNodejs,
		"jvm":    parseJvm,
	}

	profile := filepath.Join("~", ".bash_profile")
	sourceCommand := fmt.Sprintf("source %s", rcPath())

	languageSteps := []shell.Command{
		shell.And(
			shell.Command(fmt.Sprintf("mkdir -p %s", baseDirectory)),
			// TODO(akostov) Ask brian about stderr in some shell commands
			shell.Redirect(
				shell.Command(fmt.Sprintf("echo %s", shell.Quote("# Automatically generated by Koality"))),
				shell.Command(rcPath()),
				true,
			),
			shell.Or(
				shell.Silent(shell.Command(fmt.Sprintf("grep -q %s %s", shell.Quote(sourceCommand), profile))),
				shell.Append(
					shell.Command(fmt.Sprintf("echo %s", shell.Quote(sourceCommand))),
					shell.Command(profile),
					true,
				),
			),
		),
	}

	var setupSteps []shell.Command

	// TODO(akostov) environment - do we still need it?

	for language, version := range languageConfig {
		if parser, ok := languageDispatcher[language]; ok {
			languageCommand, versionCommand := parser(version)
			languageSteps = append(languageSteps, languageCommand)
			setupSteps = append(setupSteps, versionCommand)
		} else {
			return provisionCommand, BadLanguageError{fmt.Sprintf("The language %s is not currently supported by our system. Please contact us.", language)}
		}
	}

	languageCommand := shell.And(languageSteps...)
	setupCommand := shell.And(setupSteps...)

	return shell.And(
		shell.Or(
			shell.Login(shell.Sudo(languageCommand)),
			shell.And(
				shell.Command("echo Language configuration failed with return code $?."),
				shell.Command("false"),
			),
		),
		shell.Or(
			shell.Login(shell.Sudo(setupCommand)),
			shell.And(
				shell.Command("echo Setup failed with return code $?."),
				shell.Command("false"),
			),
		),
	), err
}

func parsePython(version string) (languageCommand shell.Command, versionCommand shell.Command) {
	versionMap := map[string]string{
		"2.4": "2.4.6",
		"2.5": "2.5.6",
		"2.6": "2.6.8",
		"2.7": "2.7.5",
		"3.0": "3.0.1",
		"3.1": "3.1.5",
		"3.2": "3.2.5",
		"3.3": "3.3.2",
	}

	fullVersion, ok := versionMap[version]
	if !ok {
		fullVersion = version
	}

	virtualEnvsPath := filepath.Join(baseDirectory, ".virtualenvs")
	virtualEnvPath := filepath.Join(virtualEnvsPath, version)
	virtualEnvActivatePath := filepath.Join(virtualEnvPath, "bin", "activate")

	installPythonzCommand := shell.Or(
		shell.Silent("which pythonz"),
		shell.Advertised(
			shell.Pipe(
				shell.Command("curl -kL https://raw.github.com/saghul/pythonz/master/pythonz-install"),
				shell.Command("bash"),
			),
		),
	)

	installPythonCommand := shell.Or(
		shell.And(
			shell.Silent(shell.Command(fmt.Sprintf("koality_python_bin=%s", shell.Capture(shell.Command(fmt.Sprintf("which python%s", version)))))),
			shell.Or(
				shell.Silent(shell.Command(fmt.Sprintf("python -c %s", shell.Quote("from distutils import sysconfig as s; open(s.get_config_vars()[\"INCLUDEPY\"] + \"/Python.h\")")))),
				installPackages([]string{"python-dev"}, map[string]([]string){"yum": []string{"python-devel"}}),
			),
			foundExistingCommand("python", "$koality_python_bin"),
		),
		shell.And(
			installPythonzCommand,
			shell.Advertised("source /etc/profile"),
			shell.Or(
				shell.And(
					shell.Pipe(
						shell.Command("pythonz list"),
						shell.Command(fmt.Sprintf("grep %s", shell.Quote(version))),
					),
					shell.And(
						shell.Silent(shell.Command(fmt.Sprintf("koality_python_bin=\"$PYTHONZ_ROOT/pythons/%s/bin/python\"", shell.Capture(
							shell.Pipe(
								shell.Command("ls $PYTHONZ_ROOT/pythons/"),
								shell.Command(fmt.Sprintf("grep %s", shell.Quote(version))),
								shell.Command("head -1"),
							))))),
						foundExistingCommand("python", "$koality_python_bin"),
					),
				),
				shell.And(
					installPackages([]string{"gcc", "zlib1g-dev"}, map[string]([]string){"yum": []string{"gcc", "zlib-devel", "openssl-devel"}}),
					shell.Advertised(shell.Command(fmt.Sprintf("sudo-pythonz install %s", fullVersion))),
				),
			),
		),
	)

	installEasyInstallCommand := shell.Or(
		shell.Silent("which easy_install"),
		shell.And(
			shell.Advertised("curl http://python-distribute.org/distribute_setup.py | python"),
			shell.Or(
				shell.Silent("rm distribute-*.tar.gz"),
				shell.Command("true"),
			),
		),
	)

	setupPythonCommand := shell.And(
		shell.Or(
			shell.Silent("deactivate"),
			shell.Command("true"),
		),
		installPythonCommand,
		shell.Or(
			shell.Test(shell.Command(fmt.Sprintf("-e %s", virtualEnvPath))),
			shell.And(
				shell.Or(
					shell.Silent("which virtualenv"),
					shell.And(
						installEasyInstallCommand,
						shell.Advertised("easy_install virtualenv"),
					),
				),
				shell.Command(fmt.Sprintf("mkdir -p %s", virtualEnvsPath)),
				shell.Or(
					shell.Test("\"$koality_python_bin\""),
					shell.Silent(shell.Command(fmt.Sprintf("koality_python_bin=\"$PYTHONZ_ROOT/pythons/%s/bin/python\"", shell.Capture(
						shell.Pipe(
							shell.Command("ls $PYTHONZ_ROOT/pythons/"),
							shell.Command(fmt.Sprintf("grep %s", shell.Quote(version))),
							shell.Command("head -1"),
						),
					)))),
				),
				shell.Advertised(shell.Command(fmt.Sprintf("virtualenv -p $koality_python_bin --no-site-packages %s", virtualEnvPath))),
			),
		),
	)

	sourceCommand := rcAppendCommand(fmt.Sprintf("source %s", virtualEnvActivatePath))

	languageCommand = shell.And(
		setupPythonCommand,
		sourceCommand,
	)

	return languageCommand, shell.AdvertisedWithActual(
		"python --version",
		shell.Command(fmt.Sprintf("printf \"%s%s%s\\n\"",
			shell.AnsiFormat(shell.AnsiFgGreen, shell.AnsiBold),
			shell.Capture(shell.Redirect("python --version", "/dev/stdout", true)),
			shell.AnsiFormat(shell.AnsiReset),
		)))
}

func parseRuby(version string) (languageCommand shell.Command, versionCommand shell.Command) {
	useSystemRuby := shell.And(
		shell.Or(
			shell.Silent(shell.Command(fmt.Sprintf("source %s/../scripts/rvm", shell.Capture(shell.Command(fmt.Sprintf("dirname %s", shell.Capture("which rvm"))))))),
			shell.Command("true"),
		),
		shell.Silent(shell.Command(fmt.Sprintf("koality_ruby_bin=%s", shell.Capture(shell.Command(fmt.Sprintf("which ruby-%s", version)))))),
		foundExistingCommand("ruby", "$koality_ruby_bin"),
	)

	installRubyCommand := shell.Or(
		shell.And(
			shell.Silent(shell.Command(fmt.Sprintf("rvm use %s", version))),
			foundExistingCommand("ruby", string(shell.Capture(shell.Command("which ruby")))),
		),
		shell.IfElse(
			shell.Silent("which rvm"),
			shell.Advertised(shell.Command(fmt.Sprintf("rvm install %s --verify-downloads 1", version))),
			shell.And(
				shell.Advertised(
					shell.Pipe(
						shell.Command("curl -L https://get.rvm.io"),
						shell.Command("bash"),
					),
				),
				shell.Advertised("source /usr/local/rvm/scripts/rvm"),
				shell.Advertised(shell.Command(fmt.Sprintf("rvm install %s --verify-downloads 1", version))),
			),
		),
	)

	languageCommand = shell.And(
		shell.Or(
			useSystemRuby,
			installRubyCommand,
		),
		shell.If(
			shell.Silent("which rvm"),
			shell.And(
				rcAppendCommand(string(shell.Silent(shell.Command(fmt.Sprintf("rvm use %s", version))))),
				rcAppendCommand("alias sudo=rvmsudo"),
			),
		),
	)
	return languageCommand, shell.AdvertisedWithActual("ruby --version",
		shell.Command(fmt.Sprintf("printf \"%s%s%s\\n\"",
			shell.AnsiFormat(shell.AnsiFgGreen, shell.AnsiBold),
			shell.Capture(shell.Command("ruby --version")),
			shell.AnsiFormat(shell.AnsiReset),
		)))
}

func parseNodejs(version string) (languageCommand shell.Command, versionCommand shell.Command) {
	nvmPath := filepath.Join(baseDirectory, ".nvm", "nvm.sh")

	installNodeCommand := shell.And(
		shell.Or(
			shell.Test(shell.Command(fmt.Sprintf("-e %s", nvmPath))),
			shell.And(
				shell.Or(
					shell.Silent("which git"),
					installPackages([]string{"git"}, nil),
				),
				shell.Advertised(
					shell.Pipe(
						shell.Command("curl https://raw.github.com/creationix/nvm/master/install.sh"),
						shell.Command("sh"),
					),
				),
				shell.Advertised(shell.Append(shell.Command("grep nvm ~/.bash_profile"), shell.Command(rcPath()), true)),
			),
		),
		shell.Command(fmt.Sprintf("source %s", nvmPath)),
		shell.Or(
			shell.And(
				shell.Silent(shell.Command(fmt.Sprintf("nvm use %s", version))),
				foundExistingCommand("node", string(shell.Capture("which node"))),
			),
			shell.Advertised(shell.Command(fmt.Sprintf("nvm install %s", version))),
		),
	)

	languageCommand = shell.And(
		installNodeCommand,
		rcAppendCommand(string(shell.Silent(shell.Command(fmt.Sprintf("source %s", nvmPath))))),
		rcAppendCommand(string(shell.Silent(shell.Command(fmt.Sprintf("nvm use %s", version))))),
	)

	versionCommand = shell.And(
		shell.AdvertisedWithActual("node --version",
			shell.Command(fmt.Sprintf("printf \"%s%s%s\\n\"",
				shell.AnsiFormat(shell.AnsiFgGreen, shell.AnsiBold),
				shell.Capture(shell.Command("node --version")),
				shell.AnsiFormat(shell.AnsiReset),
			))),
		shell.Chain(
			shell.Or(
				shell.Silent("npm set color always"),
				shell.Command("true"),
			),
			shell.Or(
				shell.Silent("npm set unsafe-perm true"),
				shell.Command("true"),
			),
		),
	)
	return languageCommand, versionCommand
}

func parseJvm(version string) (languageCommand shell.Command, versionCommand shell.Command) {
	version = strings.ToLower(version)
	versionAliases := map[string]([]string){
		"5":        []string{"1.5", "1.5.0"},
		"6":        []string{"1.6", "1.6.0"},
		"openjdk6": []string{"openjdk-6"},
	}

	versionMap := map[string]string{
		"5":        "/usr/lib/jvm/java-1.5.0-sun",
		"6":        "/usr/lib/jvm/java-6-sun",
		"openjdk6": "/usr/lib/jvm/java-6-openjdk",
	}

	for versionName, aliases := range versionAliases {
		for _, alias := range aliases {
			versionMap[alias] = versionMap[versionName]
		}
	}

	javaHome, ok := versionMap[version]

	if !ok {
		//TODO(akostov) BadLanguageError{fmt.Sprintf("Java version %s not supported", version)}
		return
	}

	javaPath := filepath.Join(javaHome, "bin")

	languageCommand = shell.And(
		rcAppendCommand(fmt.Sprintf("export JAVA_HOME=%s", javaHome)),
		rcAppendCommand(fmt.Sprintf("export PATH=%s:$PATH", javaPath)),
	)
	return languageCommand, shell.AdvertisedWithActual("java -version",
		shell.Command(fmt.Sprintf("printf \"%s%s%s\\n\"",
			shell.AnsiFormat(shell.AnsiFgGreen, shell.AnsiBold),
			shell.Capture(shell.Command("java -version")),
			shell.AnsiFormat(shell.AnsiReset),
		)))

}

type BadLanguageError struct {
	msg string
}

func (e BadLanguageError) Error() string { return e.msg }