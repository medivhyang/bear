package bear

type WithDialect struct {
	dialect string
}

func Dialect(name string) *WithDialect {
	return &WithDialect{dialect: name}
}

func (t *WithDialect) Table(name string) *WithTable {
	return &WithTable{dialect: t.dialect, table: name}
}

func (t *WithDialect) Select(table string, columns ...string) *QueryBuilder {
	return NewQueryBuilder().Dialect(t.dialect).Select(table, columns...)
}

func (t *WithDialect) SelectTemplate(table string, columns ...Template) *QueryBuilder {
	return NewQueryBuilder().Dialect(t.dialect).SelectTemplate(table, columns...)
}

func (t *WithDialect) SelectStruct(table string, aStruct interface{}) *QueryBuilder {
	return NewQueryBuilder().Dialect(t.dialect).SelectStruct(table, aStruct)
}

func (t *WithDialect) Insert(table string, pairs map[string]interface{}) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).Insert(table, pairs)
}

func (t *WithDialect) InsertStruct(table string, aStruct interface{}, includeZeroValues bool) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).InsertStruct(table, aStruct, includeZeroValues)
}

func (t *WithDialect) Update(table string, pairs map[string]interface{}) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).Update(table, pairs)
}

func (t *WithDialect) UpdateStruct(table string, aStruct interface{}, includeZeroValues bool) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).UpdateStruct(table, aStruct, includeZeroValues)
}

func (t *WithDialect) Delete(table string) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).Delete(table)
}
