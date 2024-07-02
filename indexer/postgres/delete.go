package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
)

func (tm *TableManager) Delete(ctx context.Context, tx *sql.Tx, key interface{}) error {
	buf := new(strings.Builder)
	var params []interface{}
	var err error
	if tm.options.RetainDeletions && tm.typ.RetainDeletions {
		params, err = tm.RetainDeleteSqlAndParams(buf, key)
	} else {
		params, err = tm.DeleteSqlAndParams(buf, key)
	}
	if err != nil {
		return err
	}

	sqlStr := buf.String()
	tm.options.Logger.Debug("Delete", "sql", sqlStr, "params", params)
	_, err = tx.ExecContext(ctx, sqlStr, params...)
	return err
}

func (tm *TableManager) DeleteSqlAndParams(w io.Writer, key interface{}) ([]interface{}, error) {
	_, err := fmt.Fprintf(w, "DELETE FROM %q", tm.TableName())
	if err != nil {
		return nil, err
	}

	_, keyParams, err := tm.WhereSqlAndParams(w, key, 1)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintf(w, ";")
	return keyParams, err
}

func (tm *TableManager) RetainDeleteSqlAndParams(w io.Writer, key interface{}) ([]interface{}, error) {
	_, err := fmt.Fprintf(w, "UPDATE %q SET _deleted = TRUE", tm.TableName())
	if err != nil {
		return nil, err
	}

	_, keyParams, err := tm.WhereSqlAndParams(w, key, 1)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprintf(w, ";")
	return keyParams, err
}
