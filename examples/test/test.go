package main

import (
	"fmt"
)

func main() {
	slice := []string{"hello"}
	fmt.Println(slice[1:])
	fmt.Println(slice[len(slice):])
}
