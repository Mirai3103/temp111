package tool

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetProceduresInput is the input schema for the getDbProcedures tool.
// No input is required; the struct is intentionally empty.
type GetProceduresInput struct{}

// ProcedureInfo represents a single stored procedure/function entry.
type ProcedureInfo struct {
	SchemaName   string `json:"schemaName"`
	FunctionName string `json:"functionName"`
	ReturnType   string `json:"returnType"`
	Arguments    string `json:"arguments"`
}

const getProceduresSQL = `
SELECT
    n.nspname                          AS schema_name,
    p.proname                          AS function_name,
    pg_get_function_result(p.oid)      AS return_type,
    pg_get_function_arguments(p.oid)   AS arguments
FROM pg_proc p
JOIN pg_namespace n ON n.oid = p.pronamespace
WHERE n.nspname = 'public'
  AND p.prokind = 'f'
  AND p.proname LIKE 'get\_%' ESCAPE '\'
ORDER BY p.proname;
`

func registerGetProcedures(g *genkit.Genkit, pool *pgxpool.Pool) *ai.ToolDef[GetProceduresInput, []ProcedureInfo] {
	return genkit.DefineTool(g, "getDbProcedures",
		"List all stored functions in the public schema whose names start with 'get_'. "+
			"Returns function name, return type, and arguments. "+
			"Use these functions via executeQuery with SELECT function_name(args).",
		func(ctx *ai.ToolContext, _ GetProceduresInput) ([]ProcedureInfo, error) {
			rows, err := pool.Query(ctx, getProceduresSQL)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			var procs []ProcedureInfo
			for rows.Next() {
				var p ProcedureInfo
				if err := rows.Scan(&p.SchemaName, &p.FunctionName, &p.ReturnType, &p.Arguments); err != nil {
					return nil, err
				}
				procs = append(procs, p)
			}
			return procs, rows.Err()
		},
	)
}
