package tool

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetTableDefInput is the input schema for the getTableDefinition tool.
type GetTableDefInput struct {
	TableName  string `json:"tableName" jsonschema_description:"Name of the table to inspect"`
	SchemaName string `json:"schemaName" jsonschema_description:"Schema of the table (e.g. public)"`
}

// ColumnInfo represents a single column definition returned by getTableDefinition.
type ColumnInfo struct {
	ColumnName       string  `json:"columnName"`
	DataType         string  `json:"dataType"`
	IsNullable       string  `json:"isNullable"`
	ColumnDefault    *string `json:"columnDefault"`
	ConstraintType   *string `json:"constraintType"`
	ReferencedTable  *string `json:"referencedTable"`
	ReferencedColumn *string `json:"referencedColumn"`
}

const getTableDefSQL = `
SELECT
  c.column_name,
  c.data_type,
  c.is_nullable,
  c.column_default,
  tc.constraint_type,
  ccu.table_name AS referenced_table,
  ccu.column_name AS referenced_column
FROM
  information_schema.columns c
LEFT JOIN
  information_schema.key_column_usage kcu
  ON c.table_name = kcu.table_name
  AND c.column_name = kcu.column_name
LEFT JOIN
  information_schema.table_constraints tc
  ON kcu.constraint_name = tc.constraint_name
  AND tc.constraint_type = 'FOREIGN KEY'
LEFT JOIN
  information_schema.constraint_column_usage ccu
  ON tc.constraint_name = ccu.constraint_name
WHERE
  c.table_name = $1
  AND c.table_schema = $2
ORDER BY
  c.ordinal_position;
`

func registerGetTableDefinition(g *genkit.Genkit, pool *pgxpool.Pool) *ai.ToolDef[GetTableDefInput, []ColumnInfo] {
	return genkit.DefineTool(g, "getTableDefinition",
		"Get the column definitions of a specific table, including data types, "+
			"nullability, defaults, and foreign key references. "+
			"Use this to understand a table's structure before writing queries.",
		func(ctx *ai.ToolContext, input GetTableDefInput) ([]ColumnInfo, error) {
			rows, err := pool.Query(ctx, getTableDefSQL, input.TableName, input.SchemaName)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			var columns []ColumnInfo
			for rows.Next() {
				var c ColumnInfo
				if err := rows.Scan(
					&c.ColumnName, &c.DataType, &c.IsNullable, &c.ColumnDefault,
					&c.ConstraintType, &c.ReferencedTable, &c.ReferencedColumn,
				); err != nil {
					return nil, err
				}
				columns = append(columns, c)
			}
			return columns, rows.Err()
		},
	)
}
