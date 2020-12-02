package bear

type WithDialect struct {
	dialect string
}

func Dialect(name string) *WithDialect {
	return &WithDialect{dialect: name}
}

func (t *WithDialect) Table(name string) *WithTable {
	return Table(name).Dialect(t.dialect)
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

func (t *WithDialect) Delete(table string) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).Delete(table)
}
