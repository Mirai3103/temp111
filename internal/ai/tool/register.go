package tool

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterTools defines all database query tools and returns them as a slice
// that can be passed to ai.WithTools(...) in the flow. Must be called after
// Genkit initialization.
func RegisterTools(g *genkit.Genkit, pool *pgxpool.Pool) []ai.Tool {
	getTablesTool := registerGetTables(g, pool)
	getTableDefTool := registerGetTableDefinition(g, pool)
	getProceduresTool := registerGetProcedures(g, pool)
	executeQueryTool := registerExecuteQuery(g, pool)

	return []ai.Tool{
		getTablesTool,
		getTableDefTool,
		getProceduresTool,
		executeQueryTool,
	}
}
