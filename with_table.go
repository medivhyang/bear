package bear

type WithTable struct {
	dialect string
	table   string
}

func Table(name string) *WithTable {
	return &WithTable{table: name}
}

func (t *WithTable) Table(name string) *WithTable {
	t.table = name
	return t
}

func (t *WithTable) Select(columns ...string) *QueryBuilder {
	return NewQueryBuilder().Dialect(t.dialect).Select(t.table, columns...)
}

func (t *WithTable) SelectTemplate(columns ...Template) *QueryBuilder {
	return NewQueryBuilder().Dialect(t.dialect).SelectTemplate(t.table, columns...)
}

func (t *WithTable) SelectStruct(aStruct interface{}) *QueryBuilder {
	return NewQueryBuilder().Dialect(t.dialect).SelectStruct(t.table, aStruct)
}

func (t *WithTable) Insert(pairs map[string]interface{}) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).Insert(t.table, pairs)
}

func (t *WithTable) InsertStruct(aStruct interface{}, includeZeroValues bool) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).InsertStruct(t.table, aStruct, includeZeroValues)
}

func (t *WithTable) Update(pairs map[string]interface{}) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).Update(t.table, pairs)
}

func (t *WithTable) UpdateStruct(aStruct interface{}, includeZeroValues bool) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).UpdateStruct(t.table, aStruct, includeZeroValues)
}

func (t *WithTable) Delete() *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).Delete(t.table)
}