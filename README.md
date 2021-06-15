# Bear

A go language sql builder.

## Why Bear

Lightweight, lower dependence and efficient.

## Quick Start

```go
package main

import (
	"fmt"
	"github.com/medivhyang/bear"

	_ "github.com/medivhyang/bear/dialect/sqlite3"
)

func main() {
	s := bear.NewBuilder().Select("user", "id", "name", "age").Build()
	fmt.Println(s)
}
```

> More examples refer to `/examples`
