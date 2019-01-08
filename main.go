package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/phayes/permbits"
)

type pathOrCommand int

const (
	path    = 1
	command = 2
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Please choose wether you want to choose a path(1) or a command(2) as the 'Exec' parameter")
	pathOrCommand := askIfPathOrCommand(reader)

	if pathOrCommand == path {
		fmt.Println("Please input path to the applications executable.")
	} else {
		fmt.Println("Please input the command to be executed.")
	}
	executablePath := askExecParameter(pathOrCommand, reader)

	fmt.Println("Please input the applications name.")
	name := askName(reader)

	fmt.Println("Please input the path to the applications icon. (Leave empty for none)")
	icon := askIcon(reader)

	var outputPath string
	if pathOrCommand == path {
		outputPath = generateOutputPath(filepath.Base(executablePath))
	} else {
		reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
		sanitizedName := reg.ReplaceAllString(name, "")
		outputPath = generateOutputPath(sanitizedName)
	}

	writeError := writeDesktopFile(outputPath, name, executablePath, icon)
	if writeError != nil {
		fmt.Printf("Error writing desktop file at '%s': %s", outputPath, writeError.Error())
		os.Exit(0)
	}
}

func askIfPathOrCommand(reader *bufio.Reader) pathOrCommand {
	answer, answerError := reader.ReadString('\n')
	if answerError != nil {
		fmt.Printf("Error reading answer: %s", answerError.Error())
		os.Exit(0)
	}

	answer = strings.TrimSuffix(answer, "\n")

	if answer == "1" {
		return path
	} else if answer == "2" {
		return command
	}

	fmt.Println("Invalid answer, choose either '1' or '2'.")
	return askIfPathOrCommand(reader)
}

func generateOutputPath(filename string) string {
	currentUser, userError := user.Current()
	if userError != nil {
		fmt.Printf("Unable to retrieve current user: %s", userError.Error())
		os.Exit(0)
	}

	return filepath.Join(currentUser.HomeDir, ".local", "share", "applications", filename+".desktop")
}

func writeDesktopFile(outputPath, name, executablePath string, iconPath *string) error {
	file, fileError := os.Create(outputPath)
	if fileError != nil {
		fmt.Printf("Unable to create file '%s': %s", outputPath, fileError.Error())
		os.Exit(0)
	}

	outputWriter := bufio.NewWriter(file)
	outputWriter.WriteString("[Desktop Entry]\n")
	outputWriter.WriteString("Type=Application\n")
	outputWriter.WriteString("Terminal=False\n")
	if iconPath != nil {
		outputWriter.WriteString(fmt.Sprintf("Icon=%s\n", *iconPath))
	}

	outputWriter.WriteString(fmt.Sprintf("Name=%s\n", name))
	outputWriter.WriteString(fmt.Sprintf("Exec=%s\n", executablePath))

	return outputWriter.Flush()
}

func askIcon(reader *bufio.Reader) *string {
	icon, iconError := reader.ReadString('\n')
	if iconError != nil {
		fmt.Printf("There was an error during execution: %s", iconError.Error())
		os.Exit(0)
	}

	icon = strings.TrimSuffix(icon, "\n")

	if len(icon) < 1 {
		return nil
	}

	absoluteIconPath, pathError := filepath.Abs(icon)
	if pathError != nil {
		fmt.Printf("The absolute path for '%s' couldn't be found: %s", icon, pathError.Error())
		os.Exit(0)
	}

	return &absoluteIconPath
}

func askName(reader *bufio.Reader) string {

	name, nameError := reader.ReadString('\n')
	if nameError != nil {
		fmt.Printf("There was an error during execution: %s", nameError.Error())
		os.Exit(0)
	}

	name = strings.TrimSuffix(name, "\n")

	if len(name) < 1 {
		fmt.Println("Thename has to be non-empty, try again.")
		return askName(reader)
	}

	return name
}

func askExecParameter(choice pathOrCommand, reader *bufio.Reader) string {
	executablePath, readError := reader.ReadString('\n')

	if readError != nil {
		fmt.Printf("There was an error during execution: %s", readError.Error())
		os.Exit(0)
	}

	executablePath = strings.TrimSuffix(executablePath, "\n")

	if len(executablePath) < 1 {
		fmt.Println("Your input has to be non-empty, try again.")
		return askExecParameter(choice, reader)
	}

	if choice == command {
		return executablePath
	}

	absoluteExecutablePath, pathError := filepath.Abs(executablePath)
	if pathError != nil {
		fmt.Printf("The absolute path for '%s' couldn't be found: %s", executablePath, readError.Error())
		os.Exit(0)
	}

	permissions, statError := permbits.Stat(absoluteExecutablePath)
	if os.IsNotExist(statError) {
		fmt.Printf("The file '%s' doesn't exist.", absoluteExecutablePath)
		os.Exit(0)
	}

	if statError != nil {
		fmt.Printf("Error querying file '%s': %s", absoluteExecutablePath, statError.Error())
		os.Exit(0)
	}

	if !permissions.UserExecute() {
		fmt.Printf("The file '%s' isn't executable, make it executable? (y/n).", absoluteExecutablePath)

		decision, _, readError := reader.ReadRune()
		if readError != nil {
			fmt.Printf("There was an error during execution: %s", readError.Error())
			os.Exit(0)
		}
		if decision == 'y' {
			permissions.SetUserExecute(true)
			permSetError := permbits.Chmod(absoluteExecutablePath, permissions)
			if permSetError != nil {
				fmt.Printf("Couldn't set file permissions for '%s'.", permSetError.Error())
			}
		}
	}

	return absoluteExecutablePath
}
