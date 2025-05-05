package tool

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type dbTool struct {
	db *sql.DB
}

func (d dbTool) Query(ctx context.Context, params map[string]string) (ret string, err error) {
	iniStr := params["ini"]
	endStr := params["end"]
	toolName := params["tool"]

	iniDt, err := time.Parse(time.RFC3339, iniStr)
	if err != nil {
		iniDt, err = time.Parse(time.RFC3339Nano, iniStr)
		if err != nil {
			return "", err
		}
	}

	endDt, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		endDt, err = time.Parse(time.RFC3339Nano, endStr)
		if err != nil {
			return "", err
		}
	}

	rows, err := d.db.QueryContext(ctx,
		"SELECT dt,value FROM kpis WHERE kpi = $1 and dt >= $2 and dt <= $3 order by dt",
		toolName, iniDt, endDt)
	if err != nil {
		return "", err
	}

	if rows.Err() != nil {
		return "", err
	}

	defer rows.Close()

	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("Valores para %s por data: ", toolName))

	for rows.Next() {
		var val float64

		var dt time.Time

		if err = rows.Scan(&dt, &val); err != nil {
			return "", err
		}

		sb.WriteString(fmt.Sprintf("%s: %v, ", dt.Format("02-01-2006"), val))
	}

	return strings.TrimSuffix(sb.String(), ", "), nil
}
