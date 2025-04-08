/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// algoinCmd represents the algoin command
var algoinCmd = &cobra.Command{
	Use:   "algoin",
	Short: "Collects searchable attributes for input into Algolia Search",
	RunE: func(cmd *cobra.Command, args []string) error {

		base, err := cmd.Flags().GetString("base")
		if err != nil {
			return fmt.Errorf("error reading base directory of Hugo site: %w", err)
		}
		fmt.Println("algoin called with ", base)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(algoinCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// algoinCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// algoinCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	algoinCmd.Flags().StringP("base", "b", "/Users/nandan/repos/guild/public", "public directory of the hugo site")
}
