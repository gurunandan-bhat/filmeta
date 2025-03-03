package model

import (
	"database/sql"
	"fmt"
)

func (m *Model) GetIDByTitle(title string) ([]int64, error) {

	var rows []int64

	qry := `SELECT iTMDBID from film where vFCGTitle = ?`
	err := m.DbHandle.Select(&rows, qry, title)
	if err != nil {
		if err == sql.ErrNoRows {
			return []int64{}, nil
		}
		return []int64{}, fmt.Errorf("error executing match query: %w", err)
	}

	return rows, nil
}
