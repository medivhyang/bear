package bear

type BuilderOptionFunc func(b *Builder)

func WithDialect(d string) BuilderOptionFunc {
	return func(b *Builder) {
		b.Dialect(d)
	}
}

func Select(table string, columns ...string) BuilderOptionFunc {
	return func(b *Builder) {
		b.Select(table, columns...)
	}
}

func SelectStruct(table string, i interface{}, ignoreColumns ...string) BuilderOptionFunc {
	return func(b *Builder) {
		b.SelectStruct(table, i, ignoreColumns...)
	}
}

func Insert(table string, columns map[string]interface{}) BuilderOptionFunc {
	return func(b *Builder) {
		b.Insert(table, columns)
	}
}

func Update(table string, columns map[string]interface{}) BuilderOptionFunc {
	return func(b *Builder) {
		b.Update(table, columns)
	}
}

func Delete(table string) BuilderOptionFunc {
	return func(b *Builder) {
		b.Delete(table)
	}
}

func Where(format string, values ...interface{}) BuilderOptionFunc {
	return func(b *Builder) {
		b.Where(format, values...)
	}
}

func WhereIn(column string, values ...interface{}) BuilderOptionFunc {
	return func(b *Builder) {
		b.WhereIn(column, values...)
	}
}

func OrderBy(fields ...string) BuilderOptionFunc {
	return func(b *Builder) {
		b.OrderBy(fields...)
	}
}

func Paging(page, size int) BuilderOptionFunc {
	return func(b *Builder) {
		b.Paging(page, size)
	}
}

func GroupBy(fields ...string) BuilderOptionFunc {
	return func(b *Builder) {
		b.GroupBy(fields...)
	}
}

func Having(format string, values ...interface{}) BuilderOptionFunc {
	return func(b *Builder) {
		b.Having(format, values...)
	}
}
