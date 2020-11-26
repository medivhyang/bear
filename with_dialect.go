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

func (t *WithDialect) Select(table string, columns ...interface{}) *QueryBuilder {
	return NewQueryBuilder().Dialect(t.dialect).Select(table, columns...)
}

func (t *WithDialect) Insert(table string, columns ...interface{}) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).Insert(table, columns...)
}

func (t *WithDialect) Update(table string, args ...interface{}) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).Update(table, args...)
}

func (t *WithDialect) UpdateStruct(table string, aStruct interface{}, includeZeroValues bool) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).Update(table, aStruct, includeZeroValues)
}

func (t *WithDialect) Delete(table string) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).Delete(table)
}
