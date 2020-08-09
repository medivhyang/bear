package bear

import (
	"database/sql"
	"errors"
	"reflect"
)

type Rows struct {
	Raw *sql.Rows
}

func WrapRows(rows *sql.Rows) *Rows {
	return &Rows{Raw: rows}
}

func (r *Rows) Scan(callback func(scan func(...interface{}) error, stop func()) error) error {
	if callback == nil {
		return nil
	}
	stop := false
	stopFunc := func() { stop = true }
	for r.Raw.Next() {
		if err := callback(r.Raw.Scan, stopFunc); err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return r.Raw.Close()
}

func (r *Rows) MapSlice() ([]map[string]interface{}, error) {
	columns, err := r.Raw.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	for i := range values {
		var v interface{}
		values[i] = &v
	}

	var list []map[string]interface{}
	for r.Raw.Next() {
		if err := r.Raw.Scan(values...); err != nil {
			return nil, err
		}
		item := make(map[string]interface{})
		for i, v := range values {
			item[columns[i]] = *v.(*interface{})
		}
		list = append(list, item)
	}

	if err := r.Raw.Close(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *Rows) Map() (map[string]interface{}, error) {
	columns, err := r.Raw.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	for i := range values {
		var v interface{}
		values[i] = &v
	}

	if !r.Raw.Next() {
		return nil, sql.ErrNoRows
	}
	if err := r.Raw.Scan(values...); err != nil {
		return nil, err
	}

	item := make(map[string]interface{})
	for i, v := range values {
		item[columns[i]] = *v.(*interface{})
	}
	if err := r.Raw.Close(); err != nil {
		return nil, err
	}

	return item, nil
}

func (r *Rows) StructSlice(structSlicePtr interface{}) error {
	structSliceValue := reflect.ValueOf(structSlicePtr)
	if structSliceValue.Kind() != reflect.Ptr {
		return errors.New("bear: scan rows to struct slice: require pointer type")
	}
	structSliceValue = structSliceValue.Elem()
	if structSliceValue.Kind() != reflect.Slice {
		return errors.New("bear: scan rows to struct slice: require slice type")
	}

	columns, err := r.Raw.Columns()
	if err != nil {
		return err
	}
	values := make([]interface{}, len(columns))

	for r.Raw.Next() {
		structValue := reflect.New(structSliceValue.Type().Elem())
		for i, name := range columns {
			values[i] = structValue.FieldByName(name).Addr().Interface()
		}
		if err := r.Raw.Scan(values...); err != nil {
			return err
		}
		reflect.Append(structSliceValue, structValue)
	}
	if err := r.Raw.Close(); err != nil {
		return err
	}

	return nil
}

func (r *Rows) Struct(structPtr interface{}) error {
	structValue := reflect.ValueOf(structPtr)
	if structValue.Kind() != reflect.Ptr {
		return errors.New("bear: scan first row to struct: require pointer type")
	}
	structValue = structValue.Elem()

	columns, err := r.Raw.Columns()
	if err != nil {
		return err
	}
	values := make([]interface{}, len(columns))
	for i, v := range columns {
		values[i] = structValue.FieldByName(v).Addr().Interface()
	}

	if !r.Raw.Next() {
		return sql.ErrNoRows
	}
	if err := r.Raw.Scan(values...); err != nil {
		return err
	}
	if err := r.Raw.Close(); err != nil {
		return err
	}

	return nil
}

