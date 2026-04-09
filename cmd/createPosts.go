/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"filmeta/tmdb"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type PostFormat struct {
	Title    string    `toml:"title"`
	Date     time.Time `toml:"date"`
	Draft    bool      `toml:"draft"`
	Cast     []string  `toml:"cast"`
	Genres   []string  `toml:"genres"`
	Director []string  `toml:"director"`
	Language []string  `toml:"language"`
}

// Hardcoding the output directory so that this will
// will not affect the guild system for which this was designed.
var rootFolder = "/Users/nandan/test"
var sectionFolder = "content/films"

// createPostsCmd represents the createPosts command
var createPostsCmd = &cobra.Command{
	Use:     "createPosts",
	Short:   "Create test posts from IMDB IDs",
	Aliases: []string{"create-posts"},
	RunE: func(cmd *cobra.Command, args []string) error {

		id, err := cmd.Flags().GetInt("id")
		if err != nil {
			return fmt.Errorf("error reading TMDB id: %w", err)
		}
		ids := []int{id}
		if id == 0 {
			idsFile, err := cmd.Flags().GetString("ids-file")
			if err != nil {
				return fmt.Errorf("error fetching file with TMDB IDs: %w", err)
			}
			ids, err = getIDsFromFile(idsFile)
			if err != nil {
				return err
			}
		}

		client := tmdb.NewClient(metaCfg.TMDB.APIKey)

		tmdbFilm, err := client.Film(context.Background(), "movie", ids[0])
		if err != nil {
			return err
		}

		fm, err := toOutputString(tmdbFilm)
		if err != nil {
			return fmt.Errorf("error generating output: %w", err)
		}

		folder := filepath.Join(rootFolder, sectionFolder, sanitizeFilename(tmdbFilm.Title))
		if err := os.MkdirAll(folder, 0755); err != nil {
			return fmt.Errorf("error making folder %s: %w", folder, err)
		}

		pFileName := filepath.Join(folder, "index.md")
		if err := os.WriteFile(pFileName, []byte(fm), 0644); err != nil {
			return fmt.Errorf("error writing post to %s: %w", pFileName, err)
		}

		if tmdbFilm.PosterPath != "" {
			posterExt := filepath.Ext(tmdbFilm.PosterPath)
			posterOutPath := filepath.Join(folder, fmt.Sprintf("poster%s", posterExt))
			if err := client.TMDBImage(context.Background(), metaCfg.TMDB.PosterBase, tmdbFilm.PosterPath, posterOutPath); err != nil {
				return fmt.Errorf("error fetching poster for %s: %w", tmdbFilm.Title, err)
			}
		}

		if tmdbFilm.BackdropPath != "" {
			backdropExt := filepath.Ext(tmdbFilm.BackdropPath)
			backdropOutPath := filepath.Join(folder, fmt.Sprintf("backdrop%s", backdropExt))
			if err := client.TMDBImage(context.Background(), metaCfg.TMDB.BackdropBase, tmdbFilm.BackdropPath, backdropOutPath); err != nil {
				return fmt.Errorf("error fetching backdrop for %s: %w", tmdbFilm.Title, err)
			}
		}

		filmBytes, err := json.MarshalIndent(tmdbFilm, "", "\t")
		if err != nil {
			return fmt.Errorf("error marshaling data for %s: %w", tmdbFilm.Title, err)
		}

		hash := md5.Sum([]byte(tmdbFilm.Title))
		metaFName := fmt.Sprintf("%s.json", hex.EncodeToString(hash[:]))

		metaFile, err := os.Create(filepath.Join(rootFolder, "assets/metadata", metaFName))
		if err != nil {
			return fmt.Errorf("error writing to file %s: %w", metaFName, err)
		}
		defer func() {
			if err := metaFile.Close(); err != nil {
				log.Fatalf("error closing file %s: %v", metaFName, err)
			}
		}()

		if _, err := metaFile.Write(filmBytes); err != nil {
			return fmt.Errorf("error writing netadata for %s: %w", metaFName, err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(createPostsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createPostsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	createPostsCmd.Flags().StringP("ids-file", "f", "", "File listing TMDB IDs one per line")
	createPostsCmd.Flags().IntP("id", "i", 0, "single tmdb id to fetch")

	createPostsCmd.MarkFlagsOneRequired("ids-file", "id")
	createPostsCmd.MarkFlagsMutuallyExclusive("ids-file", "id")
}

func getIDsFromFile(fName string) ([]int, error) {

	file, err := os.Open(fName)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", fName, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("error closing file %s: %v", fName, err)
		}
	}()

	var numbers []int
	scanner := bufio.NewScanner(file)

	// 2. Iterate through each line
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		num, err := strconv.Atoi(line)
		if err != nil {
			log.Printf("Skipping invalid line %q: %v", line, err)
			continue
		}
		numbers = append(numbers, num)
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error in file scan: %w", err)
	}

	return numbers, nil
}

func toOutputString(film tmdb.FilmWithCredits) (string, error) {

	cast := film.Cast()
	min := int(math.Min(10, float64(len(cast))))
	data := PostFormat{
		Title:    film.Title,
		Date:     time.Now(),
		Draft:    false,
		Cast:     cast[:min],
		Director: film.Director(),
		Genres:   film.Genres(),
		Language: []string{film.OriginalLanguage},
	}

	tomlBytes, err := toml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error marshaling film %s data to toml: %w", film.Title, err)
	}

	return fmt.Sprintf("+++\n%s+++\n\n%s\n", string(tomlBytes), film.Overview), nil
}

func sanitizeFilename(s string) string {
	// 1. Transliterate Unicode to ASCII (e.g., 'ñ' -> 'n')
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ = transform.String(t, s)

	// 2. Replace any non-ASCII or illegal characters with an underscore
	reg := regexp.MustCompile(`[^a-zA-Z0-9.\-]+`)
	safe := reg.ReplaceAllString(s, "-")

	// 3. Replace multiple underscores (e.g., "hello___world") with a single one ("hello_world")
	multiUnderscore := regexp.MustCompile(`-+`)
	safe = multiUnderscore.ReplaceAllString(safe, "-")

	// 4. Trim underscores and dots from the start and end (prevents hidden/invalid files)
	safe = strings.Trim(safe, ".-")

	// 5. Enforce common OS length limits
	if len(safe) > 255 {
		safe = safe[:255]
	}

	safe = strings.ToLower(safe)
	return safe
}
