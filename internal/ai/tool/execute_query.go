package tool

import (
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ExecuteQueryInput is the input schema for the executeQuery tool.
type ExecuteQueryInput struct {
	Query string `json:"query" jsonschema_description:"SQL SELECT query to execute against the database"`
}

// maxRows is the maximum number of rows returned by executeQuery to prevent
// excessive output in the AI context window.
const maxRows = 100

// forbiddenKeywords are SQL keywords that indicate a write/DDL operation.
// The tool will reject any query containing these.
var forbiddenKeywords = []string{
	"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "TRUNCATE",
	"CREATE", "GRANT", "REVOKE", "EXEC ", "EXECUTE ",
}

func registerExecuteQuery(g *genkit.Genkit, pool *pgxpool.Pool) *ai.ToolDef[ExecuteQueryInput, []map[string]interface{}] {
	return genkit.DefineTool(g, "executeQuery",
		"Execute a read-only SQL SELECT query against the database and return the results as rows. "+
			"Only SELECT statements are allowed; INSERT, UPDATE, DELETE, DROP, ALTER, etc. are rejected. "+
			"Results are capped at 100 rows. Use getDbTables and getTableDefinition first to understand the schema.",
		func(ctx *ai.ToolContext, input ExecuteQueryInput) ([]map[string]interface{}, error) {
			// Guard: reject write/DDL operations.
			upper := strings.ToUpper(strings.TrimSpace(input.Query))
			for _, kw := range forbiddenKeywords {
				if strings.Contains(upper, kw) {
					return nil, fmt.Errorf("forbidden: only SELECT queries are allowed, found '%s'", kw)
				}
			}

			rows, err := pool.Query(ctx, input.Query)
			if err != nil {
				return nil, fmt.Errorf("query execution failed: %w", err)
			}
			defer rows.Close()

			fieldDescs := rows.FieldDescriptions()
			var results []map[string]interface{}

			for rows.Next() {
				if len(results) >= maxRows {
					break
				}

				values, err := rows.Values()
				if err != nil {
					return nil, fmt.Errorf("failed to read row values: %w", err)
				}

				row := make(map[string]interface{}, len(fieldDescs))
				for i, fd := range fieldDescs {
					row[string(fd.Name)] = values[i]
				}
				results = append(results, row)
			}

			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("row iteration error: %w", err)
			}

			return results, nil
		},
	)
}
