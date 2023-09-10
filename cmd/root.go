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
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bitcanon/autotyper/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "autotyper [flags] <command>",
	Short: "A CLI tool to simulate user input",
	Long: `A CLI tool to simulate user input

This tool can be used to simulate user input in a terminal. It can be used to
test command line applications or to create demos of command line applications.`,
	Example: `  autotyper -i commands.txt
  autotyper -i commands.txt --char-delay 25
  autotyper -i commands.txt --shell cmd
  autotyper -i commands.txt --no-cls
  autotyper -i commands.txt --pre-delay 250 --post-delay 2000
  autotyper -i commands.txt -u bitcanon -H code -p C:\Users\bitcanon\Documents -s bash
  autotyper ping one.one.one.one
  cat commands.txt | autotyper`,
	Version:      "1.0.0",
	SilenceUsage: true,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		// Input string to hold the processed input
		var input string
		var err error

		// Check if data is being piped, read from file or redirected to stdin
		if viper.GetString("input-file") != "" {
			// Read input from file
			input, err = cli.ProcessFile(viper.GetString("input-file"))
			if err != nil {
				return err
			}
		} else if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
			// Process data from pipe or redirection (stdin)
			input, err = cli.ProcessStdin()
			if err != nil {
				return err
			}
		} else {
			if len(args) == 0 {
				// If there are no command line arguments, print the help and exit
				cmd.Help()
				return nil
			} else {
				// If there are command line arguments, join them
				// into a single string and use that as user input
				input = strings.Join(args, " ")
			}
		}

		// Clear the screen before printing the prompt
		if err := cli.ClearScreen(); err != nil {
			fmt.Println(err)
		}

		// Prepare the prompt
		var shellOption cli.ShellOption
		switch viper.GetString("shell") {
		case "cmd":
			shellOption = cli.Cmd
		case "bash":
			shellOption = cli.Bash
		default:
			shellOption = cli.PS
		}

		// Setup the path
		path := viper.GetString("prompt-path")
		if path == "" {
			switch shellOption {
			case cli.Cmd:
				path = "C:\\"
			case cli.Bash:
				path = "~"
			default:
				path = "C:\\"
			}
		}

		// Setup the prompt
		p := cli.Prompt{
			Username: viper.GetString("prompt-username"),
			Hostname: viper.GetString("prompt-hostname"),
			Path:     path,
			Shell:    shellOption,
		}

		// Replace "\r\n" with "\n" to ensure consistent line endings
		input = strings.ReplaceAll(input, "\r\n", "\n")

		// Split the input string into a slice of strings
		// based on the newline character
		commands := strings.Split(input, "\n")

		// Print the prompt
		cli.PrintPrompt(p, os.Stdout)

		// Delay before typing the first character of each command
		typeDelay := viper.GetInt("pre-delay")

		// Iterate over the slice of strings
		for _, command := range commands {

			// Delay before starting to type the command
			if typeDelay > 0 {
				time.Sleep(time.Duration(typeDelay) * time.Millisecond)
			}

			// Type command as human, with a delay between each character
			charDelay := viper.GetInt("char-delay")
			if err := cli.TypeAsHuman(command, os.Stdout, charDelay); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			fmt.Println()

			// Execute the command and print the output
			if err := cli.ExecuteCommand(command, os.Stdout); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

			// Print the prompt after the command output
			cli.PrintPrompt(p, os.Stdout)

			// Delay between each command
			delay := viper.GetInt("post-delay")
			if delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}

			// Clear the screen between commands (not the last command)
			lastLine := commands[len(commands)-1]
			if !viper.GetBool("no-cls") && command != lastLine {
				if err := cli.ClearScreen(); err != nil {
					fmt.Println(err)
				}
				cli.PrintPrompt(p, os.Stdout)
			}
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.autotyper.yaml)")

	// Add flags for input file
	rootCmd.Flags().StringP("input-file", "i", "", "input file")
	viper.BindPFlag("input-file", rootCmd.Flags().Lookup("input-file"))

	// Add flags for the delay between each character
	rootCmd.Flags().IntP("char-delay", "c", 75, "delay between each character in milliseconds")
	viper.BindPFlag("char-delay", rootCmd.Flags().Lookup("char-delay"))

	// Add flags for the delay between each character
	rootCmd.Flags().IntP("pre-delay", "d", 500, "delay before each command in milliseconds")
	viper.BindPFlag("pre-delay", rootCmd.Flags().Lookup("pre-delay"))

	// Add flags for the shell prompt
	rootCmd.Flags().StringP("shell", "s", "ps", "shell prompt to simulate: bash, cmd or ps")
	viper.BindPFlag("shell", rootCmd.Flags().Lookup("shell"))

	// Add flags for the username
	rootCmd.Flags().StringP("username", "u", "bitcanon", "username to print in the bash prompt")
	viper.BindPFlag("prompt-username", rootCmd.Flags().Lookup("username"))

	// Add flags for the hostname
	rootCmd.Flags().StringP("hostname", "H", "code", "hostname to print in the bash prompt")
	viper.BindPFlag("prompt-hostname", rootCmd.Flags().Lookup("hostname"))

	// Add flags for the path
	rootCmd.Flags().StringP("path", "p", "", "path to use in the prompt")
	viper.BindPFlag("prompt-path", rootCmd.Flags().Lookup("path"))

	// Add flags for the delay between each command if multiple commands are entered
	rootCmd.Flags().IntP("post-delay", "D", 3500, "delay after each command in milliseconds")
	viper.BindPFlag("post-delay", rootCmd.Flags().Lookup("post-delay"))

	// Add flags for the option to clear the screen between commands
	rootCmd.Flags().BoolP("no-cls", "n", false, "disable the clear screen between commands")
	viper.BindPFlag("no-cls", rootCmd.Flags().Lookup("no-cls"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".autotyper" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".autotyper")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
