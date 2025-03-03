package model

import (
	"filmeta/tmdb"
	"fmt"
	"strings"
)

func (m *Model) Save(f tmdb.FilmWithCredits, showType string) error {

	iTMDBID := f.Id

	tx, err := m.DbHandle.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	qry := `REPLACE INTO film 
				(iTMDBID, vTitle, vFCGTitle, vOriginalTitle, vType, vOverView, vLanguage, vBackdropPath, vPosterPath, dtReleaseDate)
			VALUES (
				?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, NULLIF(?, '')
			)`

	result, err := tx.Exec(qry,
		iTMDBID, f.Title, f.FCGTitle, f.OriginalTitle, showType, f.Overview, f.OriginalLanguage, f.BackdropPath, f.PosterPath, f.ReleaseDate,
	)
	if err != nil {
		return fmt.Errorf("error inserting film: %w", err)
	}

	iFilmID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error fetching inserted id")
	}

	genreCount := len(f.Genres)
	if genreCount > 0 {
		bindStr := `(?, ?, ?)`
		bindArray := make([]string, genreCount)
		bindValues := []any{}
		for i := range genreCount {
			bindArray[i] = bindStr
			bindValues = append(bindValues, iFilmID, f.Genres[i].Id, f.Genres[i].Name)
		}

		gQry := `INSERT INTO film_genre 
				(iFilmID, iGenreID, vGenreName)
			VALUES 
			` + strings.Join(bindArray, ", ") + `
			AS new
			ON DUPLICATE KEY UPDATE 
			vGenreName  = new.vGenreName`
		_, err = tx.Exec(gQry, bindValues...)
		if err != nil {
			return fmt.Errorf("error inserting genres: %w", err)
		}
	}
	delQry := `DELETE FROM film_credit WHERE iFilmID = ?`
	_, err = tx.Exec(delQry, iFilmID)
	if err != nil {
		return fmt.Errorf("error deleting credits: %w", err)
	}

	cast := f.Credits.Cast
	castCount := len(cast)
	if castCount > 0 {
		bindStr := `(?, ?, ?, ?, ?, ?)`
		bindArray := make([]string, castCount)
		bindValues := []any{}
		for i := range castCount {
			bindArray[i] = bindStr
			bindValues = append(bindValues, iFilmID, "cast", cast[i].Name, cast[i].Character, cast[i].ProfilePath, cast[i].Order)
		}
		castQry := `INSERT INTO film_credit 
					(iFilmID, eCreditType, vName, vRole, vProfilePath, iOrderID)
				VALUES
				` + strings.Join(bindArray, ", ")
		_, err = tx.Exec(castQry, bindValues...)
		if err != nil {
			return fmt.Errorf("error inserting cast row: %w", err)
		}
	}

	crew := f.Credits.Crew
	crewCount := len(crew)
	if crewCount > 0 {
		bindStr := `(?, ?, ?, ?, ?)`
		bindArray := []string{}
		bindValues := []any{}
		for i := range crewCount {
			bindArray = append(bindArray, bindStr)
			bindValues = append(bindValues, iFilmID, "crew", crew[i].Name, crew[i].Job, crew[i].ProfilePath)
		}
		crewQry := `INSERT INTO film_credit 
					(iFilmID, eCreditType, vName, vRole, vProfilePath)
				VALUES
				` + strings.Join(bindArray, ", ")
		_, err = tx.Exec(crewQry, bindValues...)
		if err != nil {
			return fmt.Errorf("error inserting crew row: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
