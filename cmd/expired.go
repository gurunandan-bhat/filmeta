/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Guild struct {
	Name string `json:"LinkTitle,omitempty"`
}

type Critic struct {
	Name        string `json:"LinkTitle,omitempty"`
	ReviewCount int    `json:"ReviewCount,omitempty"`
	Lastmod     string `json:"Lastmod,omitempty"`
}

type CriticMap map[string]Critic

// expiredCmd represents the expired command
var expiredCmd = &cobra.Command{
	Use:   "expired",
	Short: "List all critics without a review after a date",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("expired called")

		// First read all members
		guildFName := fmt.Sprintf("%s/guild/index.json", metaCfg.HugoRoot)
		jsonBytes, err := os.ReadFile(guildFName)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", guildFName, err)
		}
		var guildmembers = []Guild{}
		if err := json.Unmarshal(jsonBytes, &guildmembers); err != nil {
			return fmt.Errorf("error un-marshaling guild members: %w", err)
		}

		// Read critics with reviews
		criticsFName := fmt.Sprintf("%s/critics/index.json", metaCfg.HugoRoot)
		jsonBytes, err = os.ReadFile(criticsFName)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", criticsFName, err)
		}
		var critics = []Critic{}
		if err := json.Unmarshal(jsonBytes, &critics); err != nil {
			return fmt.Errorf("error un-marshaling critics: %w", err)
		}
		criticMap := make(CriticMap)
		for _, critic := range critics {
			criticMap[critic.Name] = critic
		}

		// Loop over all guild members and read off the data
		csvSlice := [][]string{}
		for _, member := range guildmembers {
			critic, ok := criticMap[member.Name]
			if !ok {
				csvSlice = append(csvSlice, []string{member.Name, "0", "NEVER"})
				continue
			}
			csvSlice = append(csvSlice, []string{critic.Name, fmt.Sprintf("%d", critic.ReviewCount), critic.Lastmod})
		}

		w := csv.NewWriter(os.Stdout)
		if err := w.WriteAll(csvSlice); err != nil {
			return fmt.Errorf("error writing csv: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(expiredCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// expiredCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// expiredCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
