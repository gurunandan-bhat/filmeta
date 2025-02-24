package model

import (
	"filmeta/tmdb"
	"strings"
)

func (m *Model) Save(f tmdb.FilmMeta) (int64, error) {

	iFilmID := f.Id

	qry := `REPLACE INTO film 
				(iFilmID, vTitle, vOriginalTitle, vOverView, vLanguage, vBackdropPath, vPosterPath, dtReleaseDate)
			VALUES (
				?, ?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?
			)`

	_, err := m.DbHandle.Exec(qry,
		iFilmID, f.Title, f.OriginalTitle, f.Overview, f.OriginalLanguage, f.BackdropPath, f.PosterPath, f.ReleaseDate,
	)
	if err != nil {
		return 0, err
	}

	genreCount := len(f.Genres)
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
	_, err = m.DbHandle.Exec(gQry, bindValues...)
	if err != nil {
		return 0, err
	}

	delQry := `DELETE FROM film_credit WHERE iFilmID = ?`
	_, err = m.DbHandle.Exec(delQry, iFilmID)
	if err != nil {
		return 0, err
	}

	cast := f.Cast
	castCount := len(cast)
	bindStr = `(?, ?, ?, ?, ?)`
	bindArray = make([]string, castCount)
	bindValues = []any{}
	for i := range castCount {
		bindArray[i] = bindStr
		bindValues = append(bindValues, iFilmID, "cast", cast[i].Name, cast[i].Role, i)
	}
	castQry := `INSERT INTO film_credit 
					(iFilmID, eCreditType, vName, vRole, iOrderID)
				VALUES
				` + strings.Join(bindArray, ", ")
	_, err = m.DbHandle.Exec(castQry, bindValues...)
	if err != nil {
		return 0, nil
	}

	crew := f.Crew
	bindStr = `(?, ?, ?, ?)`
	bindArray = []string{}
	bindValues = []any{}
	for role, name := range crew {
		bindArray = append(bindArray, bindStr)
		bindValues = append(bindValues, iFilmID, "crew", name, role)
	}
	crewQry := `INSERT INTO film_credit 
					(iFilmID, eCreditType, vName, vRole)
				VALUES
				` + strings.Join(bindArray, ", ")
	_, err = m.DbHandle.Exec(crewQry, bindValues...)
	if err != nil {
		return 0, nil
	}

	return 0, err
}
