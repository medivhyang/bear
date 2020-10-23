package bear

import "fmt"

func ExampleTableBuilderIncludeAndExclude() {
	s := CreateTable("user", []Column{
		{Name: "id", Type: "varchar(64)", Suffix: "primary key"},
		{Name: "name", Type: "varchar(512)", Suffix: "not null"},
		{Name: "created", Type: "bigint", Suffix: "not null"},
	}).
		Include("id", "name").
		Exclude("name").
		Build()
	fmt.Println(s.Format)

	// Output:
	// create table user (
	//   id varchar(64) primary key
	// );
}

func ExampleTableBuilderAppendAndPrepend() {
	s := CreateTable("user", []Column{
		{Name: "id", Type: "varchar(64)", Suffix: "primary key"},
		{Name: "name", Type: "varchar(512)", Suffix: "not null"},
		{Name: "created", Type: "bigint", Suffix: "not null"},
	}).
		Indent("", "  ").
		Build()
	fmt.Println(s.Format)

	// Output:
	// create table user (
	//   id varchar(64) primary key,
	//   name varchar(512) not null,
	//   created bigint not null
	// );
}
