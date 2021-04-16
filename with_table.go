package bear

type WithTable struct {
	dialect string
	table   string
}

func Dialect(name string) *WithTable {
	return &WithTable{dialect: name}
}

func Table(name string) *WithTable {
	return &WithTable{table: name}
}

func (t *WithTable) Table(name string) *WithTable {
	t.table = name
	return t
}

func (t *WithTable) Dialect(name string) *WithTable {
	t.dialect = name
	return t
}

func (t *WithTable) Select(columns ...interface{}) *QueryBuilder {
	return NewQueryBuilder().Dialect(t.dialect).Select(t.table, columns...)
}

func (t *WithTable) Insert(columns ...interface{}) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).Insert(t.table, columns...)
}

func (t *WithTable) Update(columns ...interface{}) *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).Update(t.table, columns...)
}

func (t *WithTable) Delete() *CommandBuilder {
	return NewCommandBuilder().Dialect(t.dialect).Delete(t.table)
}
