/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Use Go templates and generate html",
	RunE: func(cmd *cobra.Command, args []string) error {

		tRoot, err := cmd.Flags().GetString("troot")
		if err != nil {
			return fmt.Errorf("invalid template root: %w", err)
		}
		tplRoot := filepath.Join(metaCfg.AppRoot, tRoot)
		fmt.Printf("template called with %s\n", tplRoot)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// templateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// templateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	templateCmd.Flags().StringP("troot", "t", "template", "folder with templates relative to aplication root")
}
