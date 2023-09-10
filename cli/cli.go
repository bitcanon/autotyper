/*
Copyright © 2023 Mikael Schultz <bitcanon@proton.me>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Define a type for the prompt
type ShellOption int

// Define constants for the prompt
const (
	PS ShellOption = iota
	Cmd
	Bash
)

// Define a type for the prompt
type Prompt struct {
	// The prompt username and hostname (e.g. "user@host")
	Username string
	Hostname string

	// The prompt path (e.g. "C:\", "/home/user")
	Path string

	// The prompt shell (e.g. "PS", "Cmd", "Bash")
	Shell ShellOption
}

// ClearScreen clears the terminal screen. If the screen is not cleared,
// an error is returned.
func ClearScreen() error {
	var cmdList []string
	if runtime.GOOS == "windows" {
		cmdList = []string{"cmd", "/c", "cls"}
	} else {
		cmdList = []string{"sh", "-c", "clear"}
	}
	cmd := exec.Command(cmdList[0], cmdList[1:]...)
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// PrintPrompt prints a prompt to the output. The prompt is printed
// based on the shell option.
func PrintPrompt(p Prompt, out io.Writer) {
	switch p.Shell {
	case PS:
		// PowerShell prompt: "PS C:\> "
		fmt.Fprintf(out, "PS %s> ", p.Path)
	case Cmd:
		// Command prompt: "C:\> "
		fmt.Fprintf(out, "%s> ", p.Path)
	case Bash:
		// Bash prompt: "user@host:~$ "
		green := "\033[38;5;82m"
		blue := "\033[38;5;32m"
		white := "\033[0m"
		fmt.Fprintf(out, "%s%s@%s%s:%s%s%s$ ", green, p.Username, p.Hostname, white, blue, p.Path, white)
	default:
		// Unknown shell
		fmt.Fprintf(out, "Unknown shell: %v\n", p.Shell)
	}
}

// ExecuteCommand executes a command in the terminal and returns
// the output of the command as a string. If the command fails,
// an error is returned.
func ExecuteCommand(command string, out io.Writer) error {
	cmdList := strings.Split(command, " ")
	cmd := exec.Command(cmdList[0], cmdList[1:]...)
	cmd.Stdout = out
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// TypeAsHuman types a string as a human would. The delayMs parameter
// is the delay in milliseconds between each character. If the delayMs
// parameter is set to 0, there is no delay between each character.
func TypeAsHuman(str string, out io.Writer, delayMs int) error {
	// If delayMs is 0, just write the entire string to the output
	if delayMs == 0 {
		out.Write([]byte(str))
		return nil
	}

	// Colorize the first word in the string (the executable name)
	// https://talyian.github.io/ansicolors/
	fmt.Printf("\033[38;5;229m")

	// Otherwise, write each character to the output with a delay
	// between each character
	for _, char := range str {
		// If the character is a space, reset the color
		if string(char) == " " {
			fmt.Printf("\033[0m")
		}

		// Write the character to the output
		out.Write([]byte(string(char)))

		// Delay between each character
		delay := time.Duration(delayMs) * time.Millisecond
		time.Sleep(delay)
	}

	// Reset the color
	fmt.Printf("\033[0m")

	// No error
	return nil
}

// ProcessStdin reads all data from standard input
// and returns the input as a string
func ProcessStdin() (string, error) {
	// Read all data from standard input
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}

	// Return the input string on success
	return string(input), nil
}

// ProcessInteractiveInput processes the interactive input
// and extracts MAC addresses from the input string
func ProcessInteractiveInput() (string, error) {
	// A string for the user input
	var input string

	// Create a scanner to read from standard input
	scanner := bufio.NewScanner(os.Stdin)

	// Tell the user how to finish the input
	// based on the operating system
	eofKeys := "CTRL+D"
	if runtime.GOOS == "windows" {
		eofKeys = "CTRL+Z"
	}
	fmt.Fprintf(os.Stderr, "Please enter the input text. Press %s to finish.\n", eofKeys)

	// Read each line from standard input as the user types
	for scanner.Scan() {
		input += fmt.Sprintf("%s\n", scanner.Text())
	}

	// Remove the trailing newline character
	input = strings.TrimRight(input, "\n")

	// Check for errors that may have occurred while reading
	if err := scanner.Err(); err != nil {
		return "", err
	}

	// Return the input string on success
	return input, nil
}

// ProcessFile reads all data from the specified file
// and returns the input as a string
func ProcessFile(filename string) (string, error) {
	// Open the input file
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a scanner to read from the file
	scanner := bufio.NewScanner(file)

	// A string for the file contents
	var input string

	// Read each line from the file
	for scanner.Scan() {
		input += fmt.Sprintf("%s\n", scanner.Text())
	}

	// Remove the trailing newline character
	input = strings.TrimRight(input, "\n")

	// Check for errors that may have occurred while reading
	if err := scanner.Err(); err != nil {
		return "", err
	}

	// Return the input string on success
	return input, nil
}
