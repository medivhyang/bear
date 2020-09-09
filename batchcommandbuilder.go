package bear

type batchCommandBuilder struct {
	table   string
	rows    [][]Template
	include map[string]bool
	exclude map[string]bool
	where   Condition
}

func BatchInsert(table string, rows []map[string]interface{}) *batchCommandBuilder {
	result := &batchCommandBuilder{table: table}
	for _, row := range rows {
		var columns []Template
		for k, v := range row {
			columns = append(columns, NewTemplate(k, v))
		}
		if len(columns) > 0 {
			result.rows = append(result.rows, columns)
		}
	}
	return result
}
