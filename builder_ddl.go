package bear

import (
	"fmt"
	"strings"
)

type ddlAction string

const (
	ddlActionCreateTable ddlAction = "create_table"
	ddlActionDropTable   ddlAction = "drop_table"
)

func (a ddlAction) Valid() bool {
	switch a {
	case ddlActionCreateTable, ddlActionDropTable:
		return true
	default:
		return false
	}
}

type DDLTable struct {
	Name    string
	Columns []DDLColumn
}

type DDLColumn struct {
	Name   string
	Type   string
	Suffix string
}

type DDLBuilder struct {
	action  ddlAction
	dialect string
	table   string
	columns []DDLColumn
	//include          map[string]bool
	//exclude          map[string]bool
	//prepends         []Template
	//appends          []Template
	//innerPrepends    []Template
	//innerAppends     []Template
	//newline bool
	pretty bool
	prefix string
	indent string
	//onlyDropIfExists bool
	//onCreateIfExists bool
	checkExists bool
}

func NewDDLBuilder(dialect ...string) *DDLBuilder {
	b := &DDLBuilder{}
	if len(dialect) > 0 {
		b.Dialect(dialect[0])
	}
	return b
}

func (b *DDLBuilder) Dialect(name string) *DDLBuilder {
	b.dialect = name
	return b
}

func (b *DDLBuilder) Pretty(prefix string, indent string) *DDLBuilder {
	b.pretty = true
	b.prefix = prefix
	b.indent = indent
	return b
}

func (b *DDLBuilder) Compact() *DDLBuilder {
	b.pretty = false
	return b
}

func (b *DDLBuilder) CreateTable(table DDLTable, checkExists bool) *DDLBuilder {
	b.action = ddlActionCreateTable
	b.table = table.Name
	b.columns = table.Columns
	b.checkExists = checkExists
	return b
}

func (b *DDLBuilder) DropTable(table string, checkExists bool) *DDLBuilder {
	b.action = ddlActionDropTable
	b.table = table
	b.checkExists = checkExists
	return b
}

func (b *DDLBuilder) Build() Template {
	result := Template{}
	buffer := strings.Builder{}
	switch b.action {
	case ddlActionCreateTable:
		if len(b.columns) == 0 {
			break
		}
		if b.checkExists {
			buffer.WriteString(fmt.Sprintf("create table if not exists %s (", b.table))
		} else {
			buffer.WriteString(fmt.Sprintf("create table %s (", b.table))
		}
		if b.pretty {
			buffer.WriteString("\n")
		}
		for _, column := range b.columns {
			buffer.WriteString(b.indent)
			buffer.WriteString(b.prefix)
			buffer.WriteString(fmt.Sprintf("%s %s", column.Name, column.Type))
			suffix := strings.TrimSpace(column.Suffix)
			if suffix != "" {
				buffer.WriteString(" ")
				buffer.WriteString(suffix)
			}
			buffer.WriteString(",")
			if b.pretty {
				buffer.WriteString("\n")
			}
		}
		content := strings.TrimRight(buffer.String(), ", \n")
		buffer.Reset()
		buffer.WriteString(content)
		if b.pretty {
			buffer.WriteString("\n")
		}
		buffer.WriteString(");")
		if b.pretty {
			buffer.WriteString("\n")
		}
	case ddlActionDropTable:
		if b.checkExists {
			buffer.WriteString(fmt.Sprintf("drop table if exists %s;", b.table))
		} else {
			buffer.WriteString(fmt.Sprintf("drop table %s;", b.table))
		}
	}
	result.Format = strings.TrimSpace(buffer.String())
	if b.pretty && !strings.HasSuffix(result.Format, "\n") {
		result.Format += "\n"
	}
	return result
}
