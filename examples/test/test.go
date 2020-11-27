package main

import (
	"fmt"
	"log"
	"reflect"
)

func main() {
	testPtrReflect()
}

func testSlice() {
	slice := []string{"hello"}
	fmt.Println(slice[1:])
	fmt.Println(slice[len(slice):])
}

func testPtrReflect() {
	var s string

	changeValue(&s)

	log.Println(s)
}

func changeValue(s *string) {
	rv := reflect.ValueOf(s).Elem()

	//if rv.IsValid() {
	//	fmt.Println("invalid")
	//	return
	//}

	//if rv.IsZero() {
	//	fmt.Println("zero")
	//	return
	//}

	//if rv.IsNil() {
	//	fmt.Println("nil")
	//	return
	//}

	if !rv.CanSet() {
		fmt.Println("can not set")
		return
	}

	//for rv.Kind() == reflect.Ptr {
	//	rv = rv.Elem()
	//}

	rv.Set(reflect.ValueOf("hello"))
}
