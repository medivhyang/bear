package bear

type Builder interface {
	Build() (*BuildResult, error)
}

type BuildResult struct {
	SQL    string
	Names  []string
	Values []interface{}
}

type QueryBuilder struct {
	table      string
	names      []string
	values     []interface{}
	conditions []string

	fields  []string
	groupBy []string
	orderBy []string
}

func (b *QueryBuilder) Where(condition string) *QueryBuilder {
	b.conditions = append(b.conditions, condition)
	return b
}

func (b *QueryBuilder) GroupBy(items ...string) *QueryBuilder{
	b.groupBy = append(b.groupBy, items...)
	return b
}

func (b *QueryBuilder) OrderBy(items ...string) *QueryBuilder{
	b.orderBy = append(b.orderBy, items...)
	return b
}

func (b *QueryBuilder) Build() (BuildResult, error) {
	panic("implement me")
}

type CommandBuilder struct {
	action string
	table      string
	names      []string
	values     []interface{}
	conditions []string
}

func (b *CommandBuilder) Where(condition string) *CommandBuilder {
	b.conditions = append(b.conditions, condition)
	return b
}

func (b *CommandBuilder) Build() (*BuildResult, error) {
	names := make([]string, len(b.names))
	copy(names, b.names)

	values := make([]interface{}, len(b.values))
	copy(values, b.values)

	result := BuildResult{Names: names, Values: values}

	_ = result

	switch b.action {
	case "insert":
		//result.SQL = fmt.Sprintf("insert into %s(%s) values(%s) %s",
		//	b.table,
		//	strings.Join(b.names, ","),
		//)
	case "update":
	case "delete":
	}

	panic("not implemented")
}

func Select(table string, fields ...string) *QueryBuilder {
	return &QueryBuilder{table: table, fields: fields}
}

func Insert(table string, pairs map[string]interface{}) *CommandBuilder {
	b := &CommandBuilder{table: table, action: "insert"}
	for k, v := range pairs {
		b.names = append(b.names, k)
		b.values = append(b.values, v)
	}
	return b
}

func Update(table string, pairs map[string]interface{}) *CommandBuilder {
	b := &CommandBuilder{table: table, action: "update"}
	for k, v := range pairs {
		b.names = append(b.names, k)
		b.values = append(b.values, v)
	}
	return b
}

func Delete(table string) *CommandBuilder {
	b := &CommandBuilder{table: table, action: "delete"}
	return b
}
