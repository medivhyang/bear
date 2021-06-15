package bear

import (
	"database/sql"
	"reflect"

	"github.com/medivhyang/duck/ice"
	"github.com/medivhyang/duck/naming"
)

type Rows struct {
	Raw *sql.Rows
}

func NewRows(rows *sql.Rows) *Rows {
	return &Rows{Raw: rows}
}

func (r *Rows) Scan(callback func(scan func(...interface{}) error, abort func()) error) error {
	if callback == nil {
		return nil
	}
	abort := false
	abortFunc := func() { abort = true }
	for r.Raw.Next() {
		if err := callback(r.Raw.Scan, abortFunc); err != nil {
			return err
		}
		if abort {
			break
		}
	}
	return r.Raw.Close()
}

func (r *Rows) Bind(i interface{}) error {
	rv := reflect.ValueOf(i)
	if rv.Kind() != reflect.Ptr {
		panic("bear: rows: bind: require pointer type")
	}
	rv2 := ice.DeepUnrefAndNewValue(rv)
	switch rv2.Interface().(type) {
	case map[string]interface{}:
		if !rv.CanSet() {
			panic("bear: query: can not set value")
		}
		m, err := r.Map()
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(m))
	default:
		switch rv2.Kind() {
		case reflect.Struct:
			if err := r.Struct(rv2); err != nil {
				return err
			}
		case reflect.Slice:
			switch rv2.Type().Elem().Kind() {
			case reflect.Map:
				ss, err := r.MapSlice()
				if err != nil {
					return err
				}
				rv2.Set(reflect.ValueOf(ss))
			case reflect.Struct:
				if err := r.StructSlice(rv2); err != nil {
					return err
				}
			default:
				if err := r.ScalarSlice(rv2); err != nil {
					return err
				}
			}
		default:
			if err := r.Scalar(rv2); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Rows) Scalar(value interface{}) error {
	if !r.Raw.Next() {
		return sql.ErrNoRows
	}
	if err := r.Raw.Scan(value); err != nil {
		return err
	}
	return r.Raw.Close()
}

func (r *Rows) ScalarSlice(slice interface{}) error {
	reflectValue := reflect.ValueOf(slice)
	if reflectValue.Kind() != reflect.Ptr {
		return errorf("scan rows to values", "require pointer type")
	}
	reflectValue = reflectValue.Elem()
	if reflectValue.Kind() != reflect.Slice {
		return errorf("scan rows to values", "require slice type")
	}

	elemType := reflectValue.Type().Elem()

	for r.Raw.Next() {
		item := reflect.New(elemType)
		if err := r.Raw.Scan(item.Interface()); err != nil {
			return err
		}
		reflectValue.Set(reflect.Append(reflectValue, item.Elem()))
	}

	return r.Raw.Close()
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

	var items []map[string]interface{}
	for r.Raw.Next() {
		if err := r.Raw.Scan(values...); err != nil {
			return nil, err
		}
		item := make(map[string]interface{})
		for i, v := range values {
			item[columns[i]] = *v.(*interface{})
		}
		items = append(items, item)
	}

	if err := r.Raw.Close(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Rows) Struct(i interface{}) error {
	rv := reflect.ValueOf(i)
	if rv.Kind() != reflect.Ptr {
		return errorf("rows to struct", "require pointer type")
	}
	cc, err := r.Raw.Columns()
	if err != nil {
		return err
	}
	vv := make([]interface{}, 0, len(cc))
	fieldMap := r.getFieldMap(i)
	for _, c := range cc {
		if f, ok := fieldMap[c]; ok {
			vv = append(vv, f)
		} else {
			var tmp interface{}
			vv = append(vv, &tmp)
		}
	}
	if !r.Raw.Next() {
		return sql.ErrNoRows
	}
	if err := r.Raw.Scan(vv...); err != nil {
		return err
	}
	if err := r.Raw.Close(); err != nil {
		return err
	}
	return nil
}

func (r *Rows) StructSlice(i interface{}) error {
	rv := reflect.ValueOf(i)
	if rv.Kind() != reflect.Ptr {
		return errorf("rows to struct slice", "require pointer type")
	}
	rv2 := ice.DeepUnrefAndNewValue(rv)
	if rv2.Kind() != reflect.Slice {
		return errorf("rows to struct slice", "require slice type")
	}
	cc, err := r.Raw.Columns()
	if err != nil {
		return err
	}
	vv := make([]interface{}, 0, len(cc))
	for r.Raw.Next() {
		item := reflect.New(rv2.Type().Elem())
		fieldMap := r.getFieldMap(item)
		for _, c := range cc {
			if v, ok := fieldMap[c]; ok {
				vv = append(vv, v)
			} else {
				var temp interface{}
				vv = append(vv, &temp)
			}
		}
		if err := r.Raw.Scan(vv...); err != nil {
			return err
		}
		rv.Set(reflect.Append(rv, item))
	}
	if err := r.Raw.Close(); err != nil {
		return err
	}
	return nil
}

func (*Rows) getFieldMap(i interface{}) map[string]interface{} {
	fieldMap := ice.ParseStructToMap(i)
	copyFieldMap := make(map[string]interface{}, len(fieldMap))
	for k, v := range fieldMap {
		copyFieldMap[k] = v
	}
	fieldTagMap := ice.ParseStructChildTags(i, TagKey, TagItemSep, TagKVSep, TagChildKeyName)
	for name, value := range fieldMap {
		tags := fieldTagMap[name]
		if tags != nil {
			s := tags[TagChildKeyName]
			if s != "" {
				fieldMap[s] = value
				delete(fieldMap, name)
				continue
			}
		}
		if naming.ToSnake(name) != name {
			fieldMap[naming.ToSnake(name)] = value
			delete(fieldMap, name)
		}
	}
	return fieldMap
}
