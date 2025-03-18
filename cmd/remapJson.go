/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Language struct {
	Name        string `json:"name,omitempty"`
	EnglishName string `json:"english_name,omitempty"`
	ISO639_1    string `json:"iso_639_1,omitempty"`
}

// remapJsonCmd represents the remapJson command
var remapJsonCmd = &cobra.Command{
	Use:     "remapJson",
	Aliases: []string{"remap-json"},
	Short:   "Remap JSON languages",
	RunE: func(cmd *cobra.Command, args []string) error {

		inFile, err := cmd.Flags().GetString("input-file")
		if err != nil {
			return err
		}
		fmt.Println("remapJson called with ", inFile)

		jsonBytes, err := os.ReadFile(inFile)
		if err != nil {
			return err
		}

		languages := []Language{}
		if err := json.Unmarshal(jsonBytes, &languages); err != nil {
			return err
		}

		langMap := make(map[string]Language, 0)
		for _, lang := range languages {
			langMap[lang.ISO639_1] = lang
		}

		mapBytes, err := json.MarshalIndent(&langMap, "", "\t")
		if err != nil {
			return err
		}

		fmt.Println(string(mapBytes))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(remapJsonCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// remapJsonCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// remapJsonCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	remapJsonCmd.Flags().StringP("input-file", "i", "", "Input file with languages")
	cobra.MarkFlagRequired(remapJsonCmd.Flags(), "input-file")
}
