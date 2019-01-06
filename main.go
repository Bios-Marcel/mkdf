package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/phayes/permbits"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Please input path to the applications executable.")
	executablePath := askExecutablePath(reader)

	fmt.Println("Please input the applications name.")
	name := askName(reader)

	fmt.Println("Please input the path to the applications icon. (Leave empty for none)")
	icon := askIcon(reader)
	outputPath := generateOutputPath(executablePath)

	writeError := writeDesktopFile(outputPath, name, executablePath, icon)
	if writeError != nil {
		fmt.Printf("Error writing desktop file at '%s': %s", outputPath, writeError.Error())
		os.Exit(0)
	}
}

func generateOutputPath(executablePath string) string {
	currentUser, userError := user.Current()
	if userError != nil {
		fmt.Printf("Unable to retrieve current user: %s", userError.Error())
		os.Exit(0)
	}

	return filepath.Join(currentUser.HomeDir, ".local", "share", "applications", filepath.Base(executablePath)+".desktop")
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

func askExecutablePath(reader *bufio.Reader) string {
	executablePath, readError := reader.ReadString('\n')

	if readError != nil {
		fmt.Printf("There was an error during execution: %s", readError.Error())
		os.Exit(0)
	}

	executablePath = strings.TrimSuffix(executablePath, "\n")

	if len(executablePath) < 1 {
		fmt.Println("The executable path has to be non-empty, try again.")
		return askExecutablePath(reader)
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
