# Copier

  I am a copier, I copy everything from one to another

[![test status](https://github.com/uutw/copier/workflows/tests/badge.svg "test status")](https://github.com/uutw/copier/actions)
[![Coverage Status](https://coveralls.io/repos/github/uutw/copier/badge.svg?branch=master)](https://coveralls.io/github/uutw/copier?branch=master)

## This fork's purpose

- Added the `Valuer` interface.
- Fixed bug: with the `DeepCopy` option to `true`, when the destination slice or map had a longer length than the source slice or map, the destination was not resized accordingly.
- Fixed bug: if the destination field implement `sql/driver.Scanner` and the source field implements `sql/driver.Valuer`, `Value()` was not called on the source field.
- Add more test cases

## Features

* Copy from field to field with same name
* Copy from method to field with same name
* Copy from field to method with same name
* Copy from slice to slice
* Copy from struct to slice
* Copy from map to map
* Enforce copying a field with a tag
* Ignore a field with a tag
* Deep Copy

## Usage

```go
package main

import (
	"fmt"
	"github.com/uutw/copier"
)

type Address struct{
	City string
	Street string
} 

// copier.Valuer interface lets custom types implement a function returning the actual value to copy.
// For example if your type is a wrapper, or if it doesn't have to implement `sql/driver.Valuer`,
// you can implement this interface so the returned value will be used instead. It can also be used
// to format your type or convert it to another one before being copied.
// This also enables conversion for types using generics, as you cannot use them with `TypeConverter`.
func (a Address) CopyValue() interface{} {
	return fmt.Sprintf("%s, %s", a.Street, a.City)
}

type User struct {
	Name        string
	Role        string
	Age         int32
	EmployeeCode int64 `copier:"EmployeeNum"` // specify field name

	// Explicitly ignored in the destination struct.
	Salary   int

	Address Address
}

func (user *User) DoubleAge() int32 {
	return 2 * user.Age
}

// Tags in the destination Struct provide instructions to copier.Copy to ignore
// or enforce copying and to panic or return an error if a field was not copied.
type Employee struct {
	// Tell copier.Copy to panic if this field is not copied.
	Name      string `copier:"must"`

	// Tell copier.Copy to return an error if this field is not copied.
	Age       int32  `copier:"must,nopanic"`

	// Tell copier.Copy to explicitly ignore copying this field.
	Salary    int    `copier:"-"`

	DoubleAge int32
	EmployeeId int64 `copier:"EmployeeNum"` // specify field name
	SuperRole string

	Address string
}

func (employee *Employee) Role(role string) {
	employee.SuperRole = "Super " + role
}

func main() {
	var (
		user  = User{Name: "Jinzhu", Age: 18, Role: "Admin", Salary: 200000, Address: Address{Street: "123 Main Street", City: "Somewhere"}}
		users = []User{
			{Name: "Jinzhu", Age: 18, Role: "Admin", Salary: 100000, Address: Address{Street: "124 Secondary Street", City: "SomewhereElse"}},
			{Name: "jinzhu 2", Age: 30, Role: "Dev", Salary: 60000, Address: Address{Street: "125 Secondary Street", City: "SomewhereElse"}}}
		employee  = Employee{Salary: 150000}
		employees = []Employee{}
	)

	copier.Copy(&employee, &user)

	fmt.Printf("%#v \n", employee)
	// Employee{
	//    Name: "Jinzhu",                          // Copy from field
	//    Age: 18,                                 // Copy from field
	//    Salary:150000,                           // Copying explicitly ignored
	//    DoubleAge: 36,                           // Copy from method
	//    EmployeeId: 0,                           // Ignored
	//    SuperRole: "Super Admin",                // Copy to method
	//    Address: "123 Main Street, Somewhere",   // Copy from value returned by CopyValue()
	// }

	// Copy struct to slice
	copier.Copy(&employees, &user)

	fmt.Printf("%#v \n", employees)
	// []Employee{
	//   {Name: "Jinzhu", Age: 18, Salary:0, DoubleAge: 36, EmployeeId: 0, SuperRole: "Super Admin", Address: "123 Main Street, Somewhere"}
	// }

	// Copy slice to slice
	employees = []Employee{}
	copier.Copy(&employees, &users)

	fmt.Printf("%#v \n", employees)
	// []Employee{
	//   {Name: "Jinzhu", Age: 18, Salary:0, DoubleAge: 36, EmployeeId: 0, SuperRole: "Super Admin", Address: "124 Secondary Street, SomewhereElse"},
	//   {Name: "jinzhu 2", Age: 30, Salary:0, DoubleAge: 60, EmployeeId: 0, SuperRole: "Super Dev", Address: "125 Secondary Street, SomewhereElse"},
	// }

 	// Copy map to map
	map1 := map[int]int{3: 6, 4: 8}
	map2 := map[int32]int8{}
	copier.Copy(&map2, map1)

	fmt.Printf("%#v \n", map2)
	// map[int32]int8{3:6, 4:8}
}
```

### Copy with Option

```go
copier.CopyWithOption(&to, &from, copier.Option{IgnoreEmpty: true, DeepCopy: true})
```

## Contributing

You can help to make the project better, check out [http://gorm.io/contribute.html](http://gorm.io/contribute.html) for things you can do.

## License

Released under the [MIT License](https://github.com/uutw/copier/blob/master/License).
