package tool

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetTablesInput is the input schema for the getDbTables tool.
// No input is required; the struct is intentionally empty.
type GetTablesInput struct{}

// TableInfo represents a single table entry returned by getDbTables.
type TableInfo struct {
	Schema string `json:"schema"`
	Table  string `json:"table"`
}

const getTablesSQL = `
SELECT table_schema, table_name
FROM information_schema.tables
WHERE table_type = 'BASE TABLE'
  AND table_schema NOT IN ('pg_catalog', 'information_schema')
ORDER BY table_schema, table_name;
`

func registerGetTables(g *genkit.Genkit, pool *pgxpool.Pool) *ai.ToolDef[GetTablesInput, []TableInfo] {
	return genkit.DefineTool(g, "getDbTables",
		"List all base tables in the database. Returns schema and table names. "+
			"Use this tool first to discover available tables before querying them.",
		func(ctx *ai.ToolContext, _ GetTablesInput) ([]TableInfo, error) {
			rows, err := pool.Query(ctx, getTablesSQL)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			var tables []TableInfo
			for rows.Next() {
				var t TableInfo
				if err := rows.Scan(&t.Schema, &t.Table); err != nil {
					return nil, err
				}
				tables = append(tables, t)
			}
			return tables, rows.Err()
		},
	)
}
