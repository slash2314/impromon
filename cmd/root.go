/*
Copyright © 2024 Dex Wood
*/
package cmd

import (
	"bufio"
	"fmt"
	"impromon/tui"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "impromon",
	Short: "Used to quickly monitor web services on the fly.",
	Long: `Impromon is a tool to monitor web services on the fly
	For example:
		impromon -u http://example.com -u http://example2.com
		impromon -s serverlist.lst`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	Run: func(cmd *cobra.Command, args []string) {
		//servers, err := readServerList(serverListFileName)
		//if err != nil {
		//	fmt.Printf("There was an error reading the server list file: %v\n", err)
		//	os.Exit(1)
		//}
		urlsArg, err := cmd.Flags().GetStringSlice("url")
		if err != nil {
			fmt.Printf("There was an error reading the url list: %v\n", err)
			os.Exit(1)
		}
		serverList := cmd.Flag("server-list").Value.String()
		if serverList == "" {
			if len(urlsArg) == 0 {
				fmt.Println("No urls to monitor")
				cmd.Help()
				os.Exit(1)
			}
		}

		if len(urlsArg) == 0 && serverList == "" {
			fmt.Println("No urls to monitor")
			cmd.Help()
			os.Exit(1)
		}
		var urls []string
		if serverList != "" {
			urls, err = readServerList(serverList)
			if err != nil {
				fmt.Printf("There was an error reading the server list file: %v\n", err)
				os.Exit(1)
			}
		} else {
			urls = urlsArg
		}
		m := tui.InitModel(urls)
		if _, err := tea.NewProgram(&m).Run(); err != nil {
			fmt.Printf("There was an error: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringSliceP("url", "u", []string{}, "URL to monitor")
	rootCmd.Flags().StringP("server-list", "s", "", "File containing list of servers to monitor")
}

func readServerList(filename string) ([]string, error) {
	// Open the file.
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var serverList []string
	// Create a new scanner for the file.
	scanner := bufio.NewScanner(f)
	// Loop over the file.
	for scanner.Scan() {
		// Add the server to the list.
		serverList = append(serverList, scanner.Text())
	}
	// Return the server list and any errors that happened during scanning.
	return serverList, scanner.Err()
}
