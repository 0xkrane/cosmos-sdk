package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
)

func (tm *TableManager) CreateTable(ctx context.Context, tx *sql.Tx) error {
	buf := new(strings.Builder)
	err := tm.CreateTableSql(buf)
	if err != nil {
		return err
	}

	sqlStr := buf.String()
	tm.options.Logger.Debug("Creating table", "table", tm.TableName(), "sql", sqlStr)
	_, err = tx.ExecContext(ctx, sqlStr)
	return err
}

// CreateTableSql generates a CREATE TABLE statement for the object type.
func (tm *TableManager) CreateTableSql(writer io.Writer) error {
	_, err := fmt.Fprintf(writer, "CREATE TABLE IF NOT EXISTS %q (", tm.TableName())
	if err != nil {
		return err
	}
	isSingleton := false
	if len(tm.typ.KeyFields) == 0 {
		isSingleton = true
		_, err = fmt.Fprintf(writer, "_id INTEGER NOT NULL CHECK (_id = 1),\n\t")
	} else {
		for _, field := range tm.typ.KeyFields {
			err = tm.createColumnDef(writer, field)
			if err != nil {
				return err
			}
		}
	}

	for _, field := range tm.typ.ValueFields {
		err = tm.createColumnDef(writer, field)
		if err != nil {
			return err
		}
	}

	// add _deleted column when we have RetainDeletions set and enabled
	// NOTE: needs more design
	if tm.options.RetainDeletions && tm.typ.RetainDeletions {
		_, err = fmt.Fprintf(writer, "_deleted BOOLEAN NOT NULL DEFAULT FALSE,\n\t")
		if err != nil {
			return err
		}
	}

	var pKeys []string
	if !isSingleton {
		for _, field := range tm.typ.KeyFields {
			name, err := tm.updatableColumnName(field)
			if err != nil {
				return err
			}

			pKeys = append(pKeys, name)
		}
	} else {
		pKeys = []string{"_id"}
	}

	_, err = fmt.Fprintf(writer, "PRIMARY KEY (%s)", strings.Join(pKeys, ", "))
	if err != nil {
		return err
	}

	// TODO: we need test data to not generate constraint failures to safely enable this
	//for _, uniq := range tm.typ.UniqueConstraints {
	//	cols := make([]string, len(uniq.FieldNames))
	//	for i, name := range uniq.FieldNames {
	//		field, ok := tm.allFields[name]
	//		if !ok {
	//			return fmt.Errorf("unknown field %q in unique constraint", name)
	//		}
	//
	//		cols[i], err = tm.updatableColumnName(field)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//
	//	_, err = fmt.Fprintf(writer, ",\n\tUNIQUE NULLS NOT DISTINCT (%s)", strings.Join(cols, ", "))
	//}

	_, err = fmt.Fprintf(writer, `
);

GRANT SELECT ON TABLE %q TO PUBLIC;
`, tm.TableName())
	if err != nil {
		return err
	}

	return nil
}
