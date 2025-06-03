/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/spf13/cobra"
)

type Secret map[string]string

// mkcfgCmd represents the mkcfg command
var mkcfgCmd = &cobra.Command{
	Use:   "mkcfg",
	Short: "Generate the configuration",
	RunE: func(cmd *cobra.Command, args []string) error {

		secretName, err := cmd.Flags().GetString("secretName")
		if err != nil {
			return nil
		}
		secretKey, err := cmd.Flags().GetString("secretKey")
		if err != nil {
			return nil
		}

		config, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return err
		}

		// Create Secrets Manager client
		svc := secretsmanager.NewFromConfig(config)
		input := &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(secretName),
			VersionStage: aws.String("AWSCURRENT"),
		}

		result, err := svc.GetSecretValue(context.TODO(), input)
		if err != nil {
			return err
		}

		// Decrypts secret using the associated KMS key.
		var secretString string = *result.SecretString
		isPlainText, err := cmd.Flags().GetBool("plain-text")
		if err != nil {
			return err
		}
		if isPlainText {
			fmt.Println(secretString)
			return nil
		}

		secret := Secret{}
		if err := json.Unmarshal([]byte(secretString), &secret); err != nil {
			return err
		}

		secretStr, ok := secret[secretKey]
		if !ok {
			return fmt.Errorf("key %s not found in secret", secretKey)
		}

		decBytes := make([]byte, base64.StdEncoding.DecodedLen(len(secretStr)))
		n, err := base64.StdEncoding.Decode(decBytes, []byte(secretStr))
		if err != nil {
			return err
		}
		decBytes = decBytes[:n]
		fmt.Println(string(decBytes))

		// Your code goes here.
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mkcfgCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mkcfgCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	mkcfgCmd.Flags().BoolP("plain-text", "t", false, "Secret is plain text")
	mkcfgCmd.Flags().StringP("secretName", "s", "filmeta", "Name of the secret")
	mkcfgCmd.Flags().StringP("secretKey", "k", "filmeta.cfg", "Name of the key")
}
