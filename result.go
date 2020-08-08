package bear

import (
	"database/sql"
	"errors"
	"reflect"
)

type QueryResult struct {
	Rows *sql.Rows
	Err  error
}

func WrapQueryResult(rows *sql.Rows, err error) *QueryResult {
	return &QueryResult{Rows: rows, Err: err}
}

func (r *QueryResult) Scan(callback func(scan func(...interface{}) error, stop func()) error) error {
	if r.Err != nil {
		return r.Err
	}
	stop := false
	stopFunc := func() { stop = true }
	for r.Rows.Next() {
		if err := callback(r.Rows.Scan, stopFunc); err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return r.Rows.Close()
}

func (r *QueryResult) MapSlice() ([]map[string]interface{}, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	columns, err := r.Rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	for i := range values {
		var v interface{}
		values[i] = &v
	}

	var list []map[string]interface{}
	for r.Rows.Next() {
		if err := r.Rows.Scan(values...); err != nil {
			return nil, err
		}
		item := make(map[string]interface{})
		for i, v := range values {
			item[columns[i]] = *v.(*interface{})
		}
		list = append(list, item)
	}

	if err := r.Rows.Close(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *QueryResult) Map() (map[string]interface{}, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	columns, err := r.Rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	for i := range values {
		var v interface{}
		values[i] = &v
	}

	if !r.Rows.Next() {
		return nil, sql.ErrNoRows
	}
	if err := r.Rows.Scan(values...); err != nil {
		return nil, err
	}

	item := make(map[string]interface{})
	for i, v := range values {
		item[columns[i]] = *v.(*interface{})
	}
	if err := r.Rows.Close(); err != nil {
		return nil, err
	}

	return item, nil
}

func (r *QueryResult) StructSlice(structSlicePtr interface{}) error {
	structSliceValue := reflect.ValueOf(structSlicePtr)
	if structSliceValue.Kind() != reflect.Ptr {
		return errors.New("bear: scan rows to struct slice: require pointer type")
	}
	structSliceValue = structSliceValue.Elem()
	if structSliceValue.Kind() != reflect.Slice {
		return errors.New("bear: scan rows to struct slice: require slice type")
	}

	columns, err := r.Rows.Columns()
	if err != nil {
		return err
	}
	values := make([]interface{}, len(columns))

	for r.Rows.Next() {
		structValue := reflect.New(structSliceValue.Type().Elem())
		for i, name := range columns {
			values[i] = structValue.FieldByName(name).Addr().Interface()
		}
		if err := r.Rows.Scan(values...); err != nil {
			return err
		}
		reflect.Append(structSliceValue, structValue)
	}
	if err := r.Rows.Close(); err != nil {
		return err
	}

	return nil
}

func (r *QueryResult) Struct(structPtr interface{}) error {
	structValue := reflect.ValueOf(structPtr)
	if structValue.Kind() != reflect.Ptr {
		return errors.New("bear: scan first row to struct: require pointer type")
	}
	structValue = structValue.Elem()

	columns, err := r.Rows.Columns()
	if err != nil {
		return err
	}
	values := make([]interface{}, len(columns))
	for i, v := range columns {
		values[i] = structValue.FieldByName(v).Addr().Interface()
	}

	if !r.Rows.Next() {
		return sql.ErrNoRows
	}
	if err := r.Rows.Scan(values...); err != nil {
		return err
	}
	if err := r.Rows.Close(); err != nil {
		return err
	}

	return nil
}

type CommandResult struct {
	Raw sql.Result
	Err error
}

func WrapCommandResult(result sql.Result, err error) *CommandResult {
	return &CommandResult{Raw: result, Err: err}
}

func (r *CommandResult) LastInsertId() (int64, error) {
	if r.Err != nil {
		return 0, r.Err
	}
	return r.Raw.LastInsertId()
}

func (r *CommandResult) RowsAffected() (int64, error) {
	if r.Err != nil {
		return 0, r.Err
	}
	return r.Raw.RowsAffected()
}
