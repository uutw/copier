package copier_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/uutw/copier"
)

type User struct {
	Name     string
	Birthday *time.Time
	Nickname string
	Role     string
	Age      int32
	FakeAge  *int32
	Notes    []string
	flags    []byte
}

func (user User) DoubleAge() int32 {
	return 2 * user.Age
}

type Employee struct {
	Name      string
	Birthday  *time.Time
	NickName  *string
	Age       int64
	FakeAge   int
	EmployeID int64
	DoubleAge int32
	SuperRule string
	Notes     []*string
}

func (employee *Employee) Role(role string) {
	employee.SuperRule = "Super " + role
}

func checkEmployee(employee Employee, user User, t *testing.T, testCase string) {
	if employee.Name != user.Name {
		t.Errorf("%v: Name haven't been copied correctly.", testCase)
	}
	if employee.NickName == nil || *employee.NickName != user.Nickname {
		t.Errorf("%v: NickName haven't been copied correctly.", testCase)
	}
	if employee.Birthday == nil && user.Birthday != nil {
		t.Errorf("%v: Birthday haven't been copied correctly.", testCase)
	}
	if employee.Birthday != nil && user.Birthday == nil {
		t.Errorf("%v: Birthday haven't been copied correctly.", testCase)
	}
	if employee.Birthday != nil && user.Birthday != nil &&
		!employee.Birthday.Equal(*(user.Birthday)) {
		t.Errorf("%v: Birthday haven't been copied correctly.", testCase)
	}
	if employee.Age != int64(user.Age) {
		t.Errorf("%v: Age haven't been copied correctly.", testCase)
	}
	if user.FakeAge != nil && employee.FakeAge != int(*user.FakeAge) {
		t.Errorf("%v: FakeAge haven't been copied correctly.", testCase)
	}
	if employee.DoubleAge != user.DoubleAge() {
		t.Errorf("%v: Copy from method doesn't work", testCase)
	}
	if employee.SuperRule != "Super "+user.Role {
		t.Errorf("%v: Copy to method doesn't work", testCase)
	}

	if len(employee.Notes) != len(user.Notes) {
		t.Fatalf("%v: Copy from slice doesn't work, employee notes len: %v, user: %v", testCase, len(employee.Notes), len(user.Notes))
	}

	for idx, note := range user.Notes {
		if note != *employee.Notes[idx] {
			t.Fatalf("%v: Copy from slice doesn't work, notes idx: %v employee: %v user: %v", testCase, idx, *employee.Notes[idx], note)
		}
	}
}

func TestCopySameStructWithPointerField(t *testing.T) {
	var fakeAge int32 = 12
	var currentTime = time.Now()
	user := &User{Birthday: &currentTime, Name: "Jinzhu", Nickname: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, flags: []byte{'x'}}
	newUser := &User{}
	err := copier.Copy(newUser, user)
	if err != nil {
		t.Error("should not error")
	}
	if user.Birthday == newUser.Birthday {
		t.Errorf("TestCopySameStructWithPointerField: copy Birthday failed since they need to have different address")
	}

	if user.FakeAge == newUser.FakeAge {
		t.Errorf("TestCopySameStructWithPointerField: copy FakeAge failed since they need to have different address")
	}
}

func checkEmployee2(employee Employee, user *User, t *testing.T, testCase string) {
	if user == nil {
		if employee.Name != "" || employee.NickName != nil || employee.Birthday != nil || employee.Age != 0 ||
			employee.DoubleAge != 0 || employee.FakeAge != 0 || employee.SuperRule != "" || employee.Notes != nil {
			t.Errorf("%v : employee should be empty", testCase)
		}
		return
	}

	checkEmployee(employee, *user, t, testCase)
}

func TestCopySliceOfDifferentTypes(t *testing.T) {
	var ss []string
	var is []int
	if err := copier.Copy(&ss, is); err != nil {
		t.Error(err)
	}
	var anotherSs []string
	if !reflect.DeepEqual(ss, anotherSs) {
		t.Errorf("Copy nil slice to nil slice should get nil slice")
	}
}

func TestCopyStruct(t *testing.T) {
	var fakeAge int32 = 12
	user := User{Name: "Jinzhu", Nickname: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, flags: []byte{'x'}}
	employee := Employee{}

	if err := copier.Copy(employee, &user); err == nil {
		t.Errorf("Copy to unaddressable value should get error")
	}

	err := copier.Copy(&employee, &user)
	if err != nil {
		t.Error("should not error")
	}
	checkEmployee(employee, user, t, "Copy From Ptr To Ptr")

	employee2 := Employee{}
	err = copier.Copy(&employee2, user)
	if err != nil {
		t.Error("should not error")
	}
	checkEmployee(employee2, user, t, "Copy From Struct To Ptr")

	employee3 := Employee{}
	ptrToUser := &user
	err = copier.Copy(&employee3, &ptrToUser)
	if err != nil {
		t.Error("should not error")
	}
	checkEmployee(employee3, user, t, "Copy From Double Ptr To Ptr")

	employee4 := &Employee{}
	err = copier.Copy(&employee4, user)
	if err != nil {
		t.Error("should not error")
	}
	checkEmployee(*employee4, user, t, "Copy From Ptr To Double Ptr")

	employee5 := &Employee{}
	err = copier.Copy(&employee5, &employee)
	if err != nil {
		t.Error("should not error")
	}
	checkEmployee(*employee5, user, t, "Copy From Employee To Employee")
}

func TestCopyFromStructToSlice(t *testing.T) {
	user := User{Name: "Jinzhu", Age: 18, Role: "Admin", Notes: []string{"hello world"}}
	employees := []Employee{}

	if err := copier.Copy(employees, &user); err != nil && len(employees) != 0 {
		t.Errorf("Copy to unaddressable value should get error")
	}

	err := copier.Copy(&employees, &user)
	if err != nil {
		t.Error("should not error")
	}
	if len(employees) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee(employees[0], user, t, "Copy From Struct To Slice Ptr")
	}

	employees2 := &[]Employee{}
	err = copier.Copy(&employees2, user)
	if err != nil {
		t.Error("should not error")
	}
	if len(*employees2) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee((*employees2)[0], user, t, "Copy From Struct To Double Slice Ptr")
	}

	employees3 := []*Employee{}
	err = copier.Copy(&employees3, user)
	if err != nil {
		t.Error("should not error")
	}
	if len(employees3) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee(*(employees3[0]), user, t, "Copy From Struct To Ptr Slice Ptr")
	}

	employees4 := &[]*Employee{}
	err = copier.Copy(&employees4, user)
	if err != nil {
		t.Error("should not error")
	}
	if len(*employees4) != 1 {
		t.Errorf("Should only have one elem when copy struct to slice")
	} else {
		checkEmployee(*((*employees4)[0]), user, t, "Copy From Struct To Double Ptr Slice Ptr")
	}
}

func TestCopyFromSliceToSlice(t *testing.T) {
	users := []User{
		{Name: "Jinzhu", Age: 18, Role: "Admin", Notes: []string{"hello world"}},
		{Name: "Jinzhu2", Age: 22, Role: "Dev", Notes: []string{"hello world", "hello"}}}
	employees := []Employee{}

	err := copier.Copy(&employees, users)
	if err != nil {
		t.Error("should not error")
	}
	if len(employees) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee(employees[0], users[0], t, "Copy From Slice To Slice Ptr @ 1")
		checkEmployee(employees[1], users[1], t, "Copy From Slice To Slice Ptr @ 2")
	}

	employees2 := &[]Employee{}
	err = copier.Copy(&employees2, &users)
	if err != nil {
		t.Error("should not error")
	}
	if len(*employees2) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee((*employees2)[0], users[0], t, "Copy From Slice Ptr To Double Slice Ptr @ 1")
		checkEmployee((*employees2)[1], users[1], t, "Copy From Slice Ptr To Double Slice Ptr @ 2")
	}

	employees3 := []*Employee{}
	err = copier.Copy(&employees3, users)
	if err != nil {
		t.Error("should not error")
	}
	if len(employees3) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee(*(employees3[0]), users[0], t, "Copy From Slice To Ptr Slice Ptr @ 1")
		checkEmployee(*(employees3[1]), users[1], t, "Copy From Slice To Ptr Slice Ptr @ 2")
	}

	employees4 := &[]*Employee{}
	err = copier.Copy(&employees4, users)
	if err != nil {
		t.Error("should not error")
	}
	if len(*employees4) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee(*((*employees4)[0]), users[0], t, "Copy From Slice Ptr To Double Ptr Slice Ptr @ 1")
		checkEmployee(*((*employees4)[1]), users[1], t, "Copy From Slice Ptr To Double Ptr Slice Ptr @ 2")
	}
}

func TestCopyFromSliceToSlice2(t *testing.T) {
	users := []*User{{Name: "Jinzhu", Age: 18, Role: "Admin", Notes: []string{"hello world"}}, nil}
	employees := []Employee{}

	err := copier.Copy(&employees, users)
	if err != nil {
		t.Error("should not error")
	}
	if len(employees) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee2(employees[0], users[0], t, "Copy From Slice To Slice Ptr @ 1")
		checkEmployee2(employees[1], users[1], t, "Copy From Slice To Slice Ptr @ 2")
	}

	employees2 := &[]Employee{}
	err = copier.Copy(&employees2, &users)
	if err != nil {
		t.Error("should not error")
	}
	if len(*employees2) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee2((*employees2)[0], users[0], t, "Copy From Slice Ptr To Double Slice Ptr @ 1")
		checkEmployee2((*employees2)[1], users[1], t, "Copy From Slice Ptr To Double Slice Ptr @ 2")
	}

	employees3 := []*Employee{}
	err = copier.Copy(&employees3, users)
	if err != nil {
		t.Error("should not error")
	}
	if len(employees3) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee2(*(employees3[0]), users[0], t, "Copy From Slice To Ptr Slice Ptr @ 1")
		checkEmployee2(*(employees3[1]), users[1], t, "Copy From Slice To Ptr Slice Ptr @ 2")
	}

	employees4 := &[]*Employee{}
	err = copier.Copy(&employees4, users)
	if err != nil {
		t.Error("should not error")
	}
	if len(*employees4) != 2 {
		t.Errorf("Should have two elems when copy slice to slice")
	} else {
		checkEmployee2(*((*employees4)[0]), users[0], t, "Copy From Slice Ptr To Double Ptr Slice Ptr @ 1")
		checkEmployee2(*((*employees4)[1]), users[1], t, "Copy From Slice Ptr To Double Ptr Slice Ptr @ 2")
	}
}

func TestCopyFromSliceToSlice3(t *testing.T) {
	type CollectionAlias struct {
		CollectionName string `json:"collection_name"`
		Name           string `json:"name"`
	}

	expectedResult := []*CollectionAlias{
		{"collection", "collection_alias1"},
		{"collection", "collection_alias2"},
		{"collection", "collection_alias3"},
	}

	mockedResult := []*CollectionAlias{}
	err := copier.Copy(&mockedResult, &expectedResult)
	if err != nil {
		t.Error("should not error")
	}

	if len(mockedResult) != len(expectedResult) {
		t.Fatalf("failed to copy results")
	}

	for idx := range mockedResult {
		if mockedResult[idx].Name != expectedResult[idx].Name || mockedResult[idx].CollectionName != expectedResult[idx].CollectionName {
			t.Fatalf("failed to copy results")
		}
	}
}

func TestEmbeddedAndBase(t *testing.T) {
	type Base struct {
		BaseField1 int
		BaseField2 int
		User       *User
	}

	type Embed struct {
		EmbedField1 int
		EmbedField2 int
		Base
	}

	base := Base{}
	embedded := Embed{}
	embedded.BaseField1 = 1
	embedded.BaseField2 = 2
	embedded.EmbedField1 = 3
	embedded.EmbedField2 = 4

	user := User{
		Name: "testName",
	}
	embedded.User = &user

	err := copier.Copy(&base, &embedded)
	if err != nil {
		t.Error("should not error")
	}

	if base.BaseField1 != 1 || base.User.Name != "testName" {
		t.Error("Embedded fields not copied")
	}

	base.BaseField1 = 11
	base.BaseField2 = 12
	user1 := User{
		Name: "testName1",
	}
	base.User = &user1

	err = copier.Copy(&embedded, &base)
	if err != nil {
		t.Error("should not error")
	}
	if embedded.BaseField1 != 11 || embedded.User.Name != "testName1" {
		t.Error("base fields not copied")
	}
}

func TestStructField(t *testing.T) {
	type Detail struct {
		Info1 string
		Info2 *string
	}

	type UserWithDetailsPtr struct {
		Details []*Detail
		Detail  *Detail
		Notes   *[]string
		Notes2  *[]string
	}
	type UserWithDetails struct {
		Details []Detail
		Detail  Detail
		Notes   []string
		Notes2  []string
	}
	type EmployeeWithDetails struct {
		Detail Detail
	}
	type EmployeeWithDetailsPtr struct {
		Detail *Detail
	}

	optionsDeepCopy := copier.Option{
		DeepCopy: true,
	}

	checkDetail := func(t *testing.T, source Detail, target Detail) {
		if source.Info1 != target.Info1 {
			t.Errorf("info1 is diff: source: %v, target: %v", source.Info1, target.Info1)
		}

		if (source.Info2 != nil || target.Info2 != nil) && (*source.Info2 != *target.Info2) {
			t.Errorf("info2 is diff: source: %v, target: %v", *source.Info2, *target.Info2)
		}
	}

	t.Run("Should work without deepCopy", func(t *testing.T) {
		t.Run("Should work with same type and both ptr field", func(t *testing.T) {
			info2 := "world"
			from := UserWithDetailsPtr{
				Detail:  &Detail{Info1: "hello", Info2: &info2},
				Details: []*Detail{{Info1: "hello", Info2: &info2}},
			}
			to := UserWithDetailsPtr{}
			err := copier.Copy(&to, from)
			if err != nil {
				t.Error("should not error")
			}

			checkDetail(t, *from.Detail, *to.Detail)

			*to.Detail.Info2 = "new value"
			if *from.Detail.Info2 != *to.Detail.Info2 {
				t.Fatalf("DeepCopy not enabled")
			}

			if len(from.Details) != len(to.Details) {
				t.Fatalf("slice should be copied")
			}

			for idx, detail := range from.Details {
				checkDetail(t, *detail, *to.Details[idx])
			}
		})

		t.Run("Should work with same type and both not ptr field", func(t *testing.T) {
			info2 := "world"
			from := UserWithDetails{
				Detail:  Detail{Info1: "hello", Info2: &info2},
				Details: []Detail{{Info1: "hello", Info2: &info2}},
			}
			to := UserWithDetails{}
			err := copier.Copy(&to, from)
			if err != nil {
				t.Error("should not error")
			}

			checkDetail(t, from.Detail, to.Detail)

			*to.Detail.Info2 = "new value"
			if *from.Detail.Info2 != *to.Detail.Info2 {
				t.Fatalf("DeepCopy not enabled")
			}

			if len(from.Details) != len(to.Details) {
				t.Fatalf("slice should be copied")
			}

			for idx, detail := range from.Details {
				checkDetail(t, detail, to.Details[idx])
			}
		})

		t.Run("Should work with different type and both ptr field", func(t *testing.T) {
			info2 := "world"
			from := UserWithDetailsPtr{Detail: &Detail{Info1: "hello", Info2: &info2}}
			to := EmployeeWithDetailsPtr{}
			err := copier.Copy(&to, from)
			if err != nil {
				t.Error("should not error")
			}

			newValue := "new value"
			to.Detail.Info2 = &newValue

			if to.Detail.Info1 == "" {
				t.Errorf("should not be empty")
			}
			if to.Detail.Info1 != from.Detail.Info1 {
				t.Errorf("should be the same")
			}
			if to.Detail.Info2 == from.Detail.Info2 {
				t.Errorf("should be different")
			}
		})

		t.Run("Should work with different type and both not ptr field", func(t *testing.T) {
			info2 := "world"
			from := UserWithDetails{Detail: Detail{Info1: "hello", Info2: &info2}}
			to := EmployeeWithDetails{}
			err := copier.Copy(&to, from)
			if err != nil {
				t.Error("should not error")
			}

			newValue := "new value"
			to.Detail.Info2 = &newValue

			if to.Detail.Info1 == "" {
				t.Errorf("should not be empty")
			}
			if to.Detail.Info1 != from.Detail.Info1 {
				t.Errorf("should be the same")
			}
			if to.Detail.Info2 == from.Detail.Info2 {
				t.Errorf("should be different")
			}
		})

		t.Run("Should work with from ptr field and to not ptr field", func(t *testing.T) {
			info2 := "world"
			from := UserWithDetailsPtr{Detail: &Detail{Info1: "hello", Info2: &info2}}
			to := EmployeeWithDetails{}
			err := copier.Copy(&to, from)
			if err != nil {
				t.Error("should not error")
			}

			newValue := "new value"
			to.Detail.Info2 = &newValue

			if to.Detail.Info1 == "" {
				t.Errorf("should not be empty")
			}
			if to.Detail.Info1 != from.Detail.Info1 {
				t.Errorf("should be the same")
			}
			if to.Detail.Info2 == from.Detail.Info2 {
				t.Errorf("should be different")
			}
		})

		t.Run("Should work with from not ptr field and to ptr field", func(t *testing.T) {
			info2 := "world"
			from := UserWithDetails{Detail: Detail{Info1: "hello", Info2: &info2}}
			to := EmployeeWithDetailsPtr{}
			err := copier.Copy(&to, from)
			if err != nil {
				t.Error("should not error")
			}

			newValue := "new value"
			to.Detail.Info2 = &newValue

			if to.Detail.Info1 == "" {
				t.Errorf("should not be empty")
			}
			if to.Detail.Info1 != from.Detail.Info1 {
				t.Errorf("should be the same")
			}
			if to.Detail.Info2 == from.Detail.Info2 {
				t.Errorf("should be different")
			}
		})

		t.Run("Should work with from a nil ptr slice field to a slice field", func(t *testing.T) {
			notes := []string{"hello", "world"}
			from := UserWithDetailsPtr{Notes: &notes, Notes2: nil}
			to := UserWithDetails{}
			err := copier.Copy(&to, from)
			if err != nil {
				t.Errorf("should not return an error")
				return
			}

			if len(to.Notes) != len(*from.Notes) {
				t.Errorf("should be the same length")
			}
			if to.Notes[0] != (*from.Notes)[0] {
				t.Errorf("should be the same")
			}
			if to.Notes[1] != (*from.Notes)[1] {
				t.Errorf("should be the same")
			}
		})
	})

	t.Run("Should work with deepCopy", func(t *testing.T) {
		t.Run("Should work with same type and both ptr field", func(t *testing.T) {
			info2 := "world"
			from := UserWithDetailsPtr{
				Detail:  &Detail{Info1: "hello", Info2: &info2},
				Details: []*Detail{{Info1: "hello", Info2: &info2}},
			}
			to := UserWithDetailsPtr{}
			err := copier.CopyWithOption(&to, from, optionsDeepCopy)
			if err != nil {
				t.Error("should not error")
			}

			checkDetail(t, *from.Detail, *to.Detail)

			*to.Detail.Info2 = "new value"
			if *from.Detail.Info2 == *to.Detail.Info2 {
				t.Fatalf("DeepCopy enabled")
			}

			if len(from.Details) != len(to.Details) {
				t.Fatalf("slice should be copied")
			}

			for idx, detail := range from.Details {
				checkDetail(t, *detail, *to.Details[idx])
			}
		})
		t.Run("Should work with same type and both not ptr field", func(t *testing.T) {
			info2 := "world"
			from := UserWithDetails{
				Detail:  Detail{Info1: "hello", Info2: &info2},
				Details: []Detail{{Info1: "hello", Info2: &info2}},
			}
			to := UserWithDetails{}
			err := copier.CopyWithOption(&to, from, optionsDeepCopy)
			if err != nil {
				t.Error("should not error")
			}

			checkDetail(t, from.Detail, to.Detail)

			*to.Detail.Info2 = "new value"
			if *from.Detail.Info2 == *to.Detail.Info2 {
				t.Fatalf("DeepCopy enabled")
			}

			if len(from.Details) != len(to.Details) {
				t.Fatalf("slice should be copied")
			}

			for idx, detail := range from.Details {
				checkDetail(t, detail, to.Details[idx])
			}
		})

		t.Run("Should work with different type and both ptr field", func(t *testing.T) {
			info2 := "world"
			from := UserWithDetailsPtr{Detail: &Detail{Info1: "hello", Info2: &info2}}
			to := EmployeeWithDetailsPtr{}
			err := copier.CopyWithOption(&to, from, optionsDeepCopy)
			if err != nil {
				t.Error("should not error")
			}

			newValue := "new value"
			to.Detail.Info2 = &newValue

			if to.Detail.Info1 == "" {
				t.Errorf("should not be empty")
			}
			if to.Detail.Info1 != from.Detail.Info1 {
				t.Errorf("should be the same")
			}
			if to.Detail.Info2 == from.Detail.Info2 {
				t.Errorf("should be different")
			}
		})

		t.Run("Should work with different type and both not ptr field", func(t *testing.T) {
			info2 := "world"
			from := UserWithDetails{Detail: Detail{Info1: "hello", Info2: &info2}}
			to := EmployeeWithDetails{}
			err := copier.CopyWithOption(&to, from, optionsDeepCopy)
			if err != nil {
				t.Error("should not error")
			}

			newValue := "new value"
			to.Detail.Info2 = &newValue

			if to.Detail.Info1 == "" {
				t.Errorf("should not be empty")
			}
			if to.Detail.Info1 != from.Detail.Info1 {
				t.Errorf("should be the same")
			}
			if to.Detail.Info2 == from.Detail.Info2 {
				t.Errorf("should be different")
			}
		})

		t.Run("Should work with from ptr field and to not ptr field", func(t *testing.T) {
			info2 := "world"
			from := UserWithDetailsPtr{Detail: &Detail{Info1: "hello", Info2: &info2}}
			to := EmployeeWithDetails{}
			err := copier.CopyWithOption(&to, from, optionsDeepCopy)
			if err != nil {
				t.Error("should not error")
			}

			newValue := "new value"
			to.Detail.Info2 = &newValue

			if to.Detail.Info1 == "" {
				t.Errorf("should not be empty")
			}
			if to.Detail.Info1 != from.Detail.Info1 {
				t.Errorf("should be the same")
			}
			if to.Detail.Info2 == from.Detail.Info2 {
				t.Errorf("should be different")
			}
		})

		t.Run("Should work with from not ptr field and to ptr field", func(t *testing.T) {
			info2 := "world"
			from := UserWithDetails{Detail: Detail{Info1: "hello", Info2: &info2}}
			to := EmployeeWithDetailsPtr{}
			err := copier.CopyWithOption(&to, from, optionsDeepCopy)
			if err != nil {
				t.Error("should not error")
			}

			newValue := "new value"
			to.Detail.Info2 = &newValue

			if to.Detail.Info1 == "" {
				t.Errorf("should not be empty")
			}
			if to.Detail.Info1 != from.Detail.Info1 {
				t.Errorf("should be the same")
			}
			if to.Detail.Info2 == from.Detail.Info2 {
				t.Errorf("should be different")
			}
		})

		t.Run("Should work with from a nil ptr slice field to a slice field", func(t *testing.T) {
			notes := []string{"hello", "world"}
			from := UserWithDetailsPtr{Notes: &notes, Notes2: nil}
			to := UserWithDetails{}
			err := copier.CopyWithOption(&to, from, optionsDeepCopy)
			if err != nil {
				t.Errorf("should not return an error")
				return
			}

			if len(to.Notes) != len(*from.Notes) {
				t.Errorf("should be the same length")
			}
			if to.Notes[0] != (*from.Notes)[0] {
				t.Errorf("should be the same")
			}
			if to.Notes[1] != (*from.Notes)[1] {
				t.Errorf("should be the same")
			}

			newValue := []string{"new", "value"}
			to.Notes = newValue

			if to.Notes[0] == (*from.Notes)[0] {
				t.Errorf("should be different")
			}
			if to.Notes[1] == (*from.Notes)[1] {
				t.Errorf("should be different")
			}
		})
	})
}

func TestMapInterface(t *testing.T) {
	type Inner struct {
		IntPtr          *int
		unexportedField string
	}

	type Outer struct {
		Inner Inner
	}

	type DriverOptions struct {
		GenOptions map[string]interface{}
	}

	t.Run("Should work without deepCopy", func(t *testing.T) {
		intVal := 5
		outer := Outer{
			Inner: Inner{
				IntPtr:          &intVal,
				unexportedField: "hello",
			},
		}
		from := DriverOptions{
			GenOptions: map[string]interface{}{
				"key": outer,
			},
		}
		to := DriverOptions{}
		if err := copier.Copy(&to, &from); nil != err {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		*to.GenOptions["key"].(Outer).Inner.IntPtr = 6

		if to.GenOptions["key"].(Outer).Inner.IntPtr != from.GenOptions["key"].(Outer).Inner.IntPtr {
			t.Errorf("should be the same")
		}
	})

	t.Run("Should work with deepCopy", func(t *testing.T) {
		intVal := 5
		outer := Outer{
			Inner: Inner{
				IntPtr:          &intVal,
				unexportedField: "Hello",
			},
		}
		from := DriverOptions{
			GenOptions: map[string]interface{}{
				"key": outer,
			},
		}
		to := DriverOptions{}
		if err := copier.CopyWithOption(&to, &from, copier.Option{
			DeepCopy: true,
		}); nil != err {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		*to.GenOptions["key"].(Outer).Inner.IntPtr = 6

		if to.GenOptions["key"].(Outer).Inner.IntPtr == from.GenOptions["key"].(Outer).Inner.IntPtr {
			t.Errorf("should be different")
		}
	})

	t.Run("Test copy map with nil interface", func(t *testing.T) {
		from := map[string]interface{}{"eventId": nil}
		to := map[string]interface{}{"eventId": nil}
		err := copier.CopyWithOption(&to, &from, copier.Option{IgnoreEmpty: true, DeepCopy: true})
		if err != nil {
			t.Error("should not error")
		}
		if v, ok := to["eventId"]; !ok || v != nil {
			t.Errorf("failed to deep copy map with nil, got %v", v)
		}

		from["eventId"] = 1
		if v, ok := to["eventId"]; !ok || v != nil {
			t.Errorf("failed to deep copy map with nil, got %v", v)
		}

		err = copier.CopyWithOption(&to, &from, copier.Option{IgnoreEmpty: true, DeepCopy: true})
		if err != nil {
			t.Error("should not error")
		}
		if v, ok := to["eventId"]; !ok || v != 1 {
			t.Errorf("failed to deep copy map with nil")
		}

		from["eventId"] = 2
		if v, ok := to["eventId"]; !ok || v != 1 {
			t.Errorf("failed to deep copy map with nil")
		}
	})

	t.Run("Test copy map with nested slice map", func(t *testing.T) {
		var out map[string]interface{}
		value := map[string]interface{}{
			"list": []map[string]interface{}{
				{
					"shop_id": 123,
				},
			},
			"list2": []interface{}{
				map[string]interface{}{
					"shop_id": 123,
				},
			},
		}
		err := copier.CopyWithOption(&out, &value, copier.Option{IgnoreEmpty: false, DeepCopy: true})
		if err != nil {
			t.Fatalf("failed to deep copy nested map")
		}
		if fmt.Sprintf("%v", out) != fmt.Sprintf("%v", value) {
			t.Fatalf("failed to deep copy nested map")
		}
	})
}

func TestInterface(t *testing.T) {
	type Inner struct {
		IntPtr *int
	}

	type Outer struct {
		Inner Inner
	}

	type DriverOptions struct {
		GenOptions interface{}
	}

	t.Run("Should work without deepCopy", func(t *testing.T) {
		intVal := 5
		outer := Outer{
			Inner: Inner{
				IntPtr: &intVal,
			},
		}
		from := DriverOptions{
			GenOptions: outer,
		}
		to := DriverOptions{}
		if err := copier.Copy(&to, from); nil != err {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		*to.GenOptions.(Outer).Inner.IntPtr = 6

		if to.GenOptions.(Outer).Inner.IntPtr != from.GenOptions.(Outer).Inner.IntPtr {
			t.Errorf("should be the same")
		}
	})

	t.Run("Should work with deepCopy", func(t *testing.T) {
		intVal := 5
		outer := Outer{
			Inner: Inner{
				IntPtr: &intVal,
			},
		}
		from := DriverOptions{
			GenOptions: outer,
		}
		to := DriverOptions{}
		if err := copier.CopyWithOption(&to, &from, copier.Option{
			DeepCopy: true,
		}); nil != err {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		*to.GenOptions.(Outer).Inner.IntPtr = 6

		if to.GenOptions.(Outer).Inner.IntPtr == from.GenOptions.(Outer).Inner.IntPtr {
			t.Errorf("should be different")
		}
	})
}

func TestSlice(t *testing.T) {
	type ElemOption struct {
		Value int
	}

	type A struct {
		X       []int
		Options []ElemOption
	}

	type B struct {
		X       []int
		Options []ElemOption
	}

	t.Run("Should work with simple slice", func(t *testing.T) {
		from := []int{1, 2}
		var to []int

		if err := copier.Copy(&to, from); nil != err {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		from[0] = 3
		from[1] = 4

		if to[0] == from[0] {
			t.Errorf("should be different")
		}

		if len(to) != len(from) {
			t.Errorf("should be the same length, got len(from): %v, len(to): %v", len(from), len(to))
		}
	})

	t.Run("Should work with empty slice", func(t *testing.T) {
		from := []int{}
		to := []int{}

		if err := copier.Copy(&to, from); nil != err {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		if to == nil {
			t.Errorf("should be not nil")
		}
	})

	t.Run("Should work without deepCopy", func(t *testing.T) {
		x := []int{1, 2}
		options := []ElemOption{
			{Value: 10},
			{Value: 20},
		}
		from := A{X: x, Options: options}
		to := B{}

		if err := copier.Copy(&to, from); nil != err {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		from.X[0] = 3
		from.X[1] = 4
		from.Options[0].Value = 30
		from.Options[1].Value = 40

		if to.X[0] != from.X[0] {
			t.Errorf("should be the same")
		}

		if len(to.X) != len(from.X) {
			t.Errorf("should be the same length, got len(from.X): %v, len(to.X): %v", len(from.X), len(to.X))
		}

		if to.Options[0].Value != from.Options[0].Value {
			t.Errorf("should be the same")
		}

		if to.Options[0].Value != from.Options[0].Value {
			t.Errorf("should be the same")
		}

		if len(to.Options) != len(from.Options) {
			t.Errorf("should be the same")
		}
	})

	t.Run("Should work with deepCopy", func(t *testing.T) {
		x := []int{1, 2}
		options := []ElemOption{
			{Value: 10},
			{Value: 20},
		}
		from := A{X: x, Options: options}
		to := B{}

		if err := copier.CopyWithOption(&to, from, copier.Option{
			DeepCopy: true,
		}); nil != err {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		from.X[0] = 3
		from.X[1] = 4
		from.Options[0].Value = 30
		from.Options[1].Value = 40

		if to.X[0] == from.X[0] {
			t.Errorf("should be different")
		}

		if len(to.X) != len(from.X) {
			t.Errorf("should be the same length, got len(from.X): %v, len(to.X): %v", len(from.X), len(to.X))
		}

		if to.Options[0].Value == from.Options[0].Value {
			t.Errorf("should be different")
		}

		if len(to.Options) != len(from.Options) {
			t.Errorf("should be the same")
		}
	})
}

func TestAnonymousFields(t *testing.T) {
	t.Run("Should work with unexported ptr fields", func(t *testing.T) {
		type nested struct {
			A string
		}
		type parentA struct {
			*nested
		}
		type parentB struct {
			*nested
		}

		from := parentA{nested: &nested{A: "a"}}
		to := parentB{}

		err := copier.CopyWithOption(&to, &from, copier.Option{
			DeepCopy: true,
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		from.A = "b"

		if to.nested != nil {
			t.Errorf("should be nil")
		}
	})
	t.Run("Should work with unexported fields", func(t *testing.T) {
		type nested struct {
			A string
		}
		type parentA struct {
			nested
		}
		type parentB struct {
			nested
		}

		from := parentA{nested: nested{A: "a"}}
		to := parentB{}

		err := copier.CopyWithOption(&to, &from, copier.Option{
			DeepCopy: true,
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		from.A = "b"

		if to.A == from.A {
			t.Errorf("should be different")
		}
	})

	t.Run("Should work with exported ptr fields", func(t *testing.T) {
		type Nested struct {
			A string
		}
		type parentA struct {
			*Nested
		}
		type parentB struct {
			*Nested
		}

		fieldValue := "a"
		from := parentA{Nested: &Nested{A: fieldValue}}
		to := parentB{}

		err := copier.CopyWithOption(&to, &from, copier.Option{
			DeepCopy: true,
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		from.A = "b"

		if to.A != fieldValue {
			t.Errorf("should not change")
		}
	})

	t.Run("Should work with exported ptr fields with same name src field", func(t *testing.T) {
		type Nested struct {
			A string
		}
		type parentA struct {
			A string
		}
		type parentB struct {
			*Nested
		}

		fieldValue := "a"
		from := parentA{A: fieldValue}
		to := parentB{}

		err := copier.CopyWithOption(&to, &from, copier.Option{
			DeepCopy: true,
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		from.A = "b"

		if to.A != fieldValue {
			t.Errorf("should not change")
		}
	})

	t.Run("Should work with exported fields", func(t *testing.T) {
		type Nested struct {
			A string
		}
		type parentA struct {
			Nested
		}
		type parentB struct {
			Nested
		}

		fieldValue := "a"
		from := parentA{Nested: Nested{A: fieldValue}}
		to := parentB{}

		err := copier.CopyWithOption(&to, &from, copier.Option{
			DeepCopy: true,
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}

		from.A = "b"

		if to.A != fieldValue {
			t.Errorf("should not change")
		}
	})
}

type someStruct struct {
	IntField  int
	UIntField uint64
}

type structSameName1 struct {
	A string
	B int64
	C time.Time
	D string
	E *someStruct
}

type structSameName2 struct {
	A string
	B time.Time
	C int64
	D string
	E *someStruct
}

func TestCopyFieldsWithSameNameButDifferentTypes(t *testing.T) {
	obj1 := structSameName1{A: "123", B: 2, C: time.Now()}
	obj2 := &structSameName2{}
	err := copier.Copy(obj2, &obj1)
	if err != nil {
		t.Error("Should not raise error")
	}

	if obj2.A != obj1.A {
		t.Errorf("Field A should be copied")
	}
}

type Foo1 struct {
	Name string
	Age  int32
}

type Foo2 struct {
	Name string
}

type StructWithMap1 struct {
	Map map[int]Foo1
}

type StructWithMap2 struct {
	Map map[int32]Foo2
}

func TestCopyMapOfStruct(t *testing.T) {
	obj1 := StructWithMap1{Map: map[int]Foo1{2: {Name: "A pure foo"}}}
	obj2 := &StructWithMap2{}
	err := copier.Copy(obj2, obj1)
	if err != nil {
		t.Error("Should not raise error")
	}
	for k, v1 := range obj1.Map {
		v2, ok := obj2.Map[int32(k)]
		if !ok || v1.Name != v2.Name {
			t.Errorf("Map should be copied")
		}
	}
}

func TestCopyMapOfInt(t *testing.T) {
	map1 := map[int]int{3: 6, 4: 8}
	map2 := map[int32]int8{}
	err := copier.Copy(&map2, map1)
	if err != nil {
		t.Error("Should not raise error")
	}

	for k, v1 := range map1 {
		v2, ok := map2[int32(k)]
		if !ok || v1 != int(v2) {
			t.Errorf("Map should be copied")
		}
	}
}

func TestCopyMapOfSliceValue(t *testing.T) {
	// case1: map's value is a simple slice
	key, value := 2, 3
	src := map[int][]int{key: {value}}

	dst1 := map[int][]int{}
	var dst2 map[int][]int
	err := copier.Copy(&dst1, src)
	if err != nil {
		t.Error("Should not raise error")
	}
	err = copier.Copy(&dst2, src)
	if err != nil {
		t.Error("Should not raise error")
	}

	for k, v1 := range src {
		v2, ok := dst1[k]
		if !ok || len(v1) != len(v2) || k != key {
			t.Errorf("Map should be copied")
		}
		for i := range v1 {
			if v2[i] != value {
				t.Errorf("Map's slice value shoud be copied")
			}
		}

		v3, ok := dst2[k]
		if !ok || len(v1) != len(v3) {
			t.Errorf("Map should be copied")
		}
		for i := range v1 {
			if v3[i] != value {
				t.Errorf("Map's slice value shoud be copied")
			}
		}
	}

	// case2: map's value is a slice whose element is map
	key1, key2 := 2, 3
	value = 4
	s := map[int][]map[int]int{key1: {{key2: value}}}
	d1 := map[int][]map[int]int{key1: {{key1: key2}}}
	d2 := map[int][]map[int]int{key1: {}}
	d3 := map[int][]map[int]int{key1: nil}
	d4 := map[int][]map[int]int{}
	d5 := map[int][]map[int]int(nil)
	ms := []map[int][]map[int]int{d1, d2, d3, d4, d5}
	for i := range ms {
		err := copier.CopyWithOption(&ms[i], s, copier.Option{IgnoreEmpty: false, DeepCopy: true})
		if err != nil {
			t.Error("should not error")
		}

		if len(ms[i]) != len(s) {
			t.Errorf("Number of map's keys should be equal")
		}
		for k, sliceMap := range ms[i] {
			if k != key1 {
				t.Errorf("Map's key should be copied")
			}
			if len(sliceMap) != len(s[key1]) || len(sliceMap) != 1 {
				t.Errorf("Map's slice value should be copied")
			}
			m := sliceMap[0]
			if len(m) != len(s[key1][0]) || len(m) != 1 {
				t.Errorf("Map's slice value should be copied recursively")
			}
			for k, v := range m {
				if k != key2 || v != value {
					t.Errorf("Map's slice value should be copied recursively")
				}
			}
		}
	}
}

func TestCopyMapOfPtrValue(t *testing.T) {
	intV := 3
	intv := intV
	src := map[int]*int{2: &intv}
	dst1 := map[int]*int{}
	var dst2 map[int]*int
	err := copier.Copy(&dst1, src)
	if err != nil {
		t.Error("Should not raise error")
	}
	err = copier.Copy(&dst2, src)
	if err != nil {
		t.Error("Should not raise error")
	}

	for k, v1 := range src {
		v2, ok := dst1[k]
		if !ok || v2 == nil || v1 == nil || *v2 != *v1 || *v2 != intV {
			t.Errorf("Map should be copied")
		}

		v3, ok := dst2[k]
		if !ok || v3 == nil || *v3 != *v1 || *v3 != intV {
			t.Errorf("Map should be copied")
		}
	}
}

func TestCopyWithOption(t *testing.T) {
	from := structSameName2{D: "456", E: &someStruct{IntField: 100, UIntField: 1000}}
	to := &structSameName1{A: "123", B: 2, C: time.Now(), D: "123", E: &someStruct{UIntField: 5000}}
	if err := copier.CopyWithOption(to, &from, copier.Option{IgnoreEmpty: true}); err != nil {
		t.Error("Should not raise error")
	}

	if to.A == from.A {
		t.Errorf("Field A should not be copied")
	} else if to.D != from.D {
		t.Errorf("Field D should be copied")
	}
}

type ScannerValue struct {
	V int
}

func (s *ScannerValue) Scan(_ interface{}) error {
	return errors.New("I failed")
}

type ScannerStruct struct {
	V *ScannerValue
}

type ScannerStructTo struct {
	V *ScannerValue
}

func TestScanner(t *testing.T) {
	s := &ScannerStruct{
		V: &ScannerValue{
			V: 12,
		},
	}

	s2 := &ScannerStructTo{}

	err := copier.Copy(s2, s)
	if err != nil {
		t.Error("Should not raise error")
	}

	if s.V.V != s2.V.V {
		t.Errorf("Field V should be copied")
	}
}

type Valuer struct {
	A string
	B int
}

func (v Valuer) Value() (driver.Value, error) {
	return json.Marshal(v)
}

type ScannerString string

func (s *ScannerString) Scan(v any) error {
	switch val := v.(type) {
	case string:
		*s = ScannerString(val)
		return nil
	case []byte:
		*s = ScannerString(val)
		return nil
	default:
		return fmt.Errorf("ScannerString: cannot scan %#v", v)
	}
}

func TestScannerWithValuerSrc(t *testing.T) {
	from := &struct {
		Field *Valuer
	}{
		Field: &Valuer{
			A: "a",
			B: 123,
		},
	}

	to := &struct {
		Field ScannerString
	}{}

	err := copier.Copy(to, from)
	if err != nil {
		t.Error("Should not raise error")
	}

	expected := ScannerString(`{"A":"a","B":123}`)
	if to.Field != expected {
		t.Errorf("to.Field expected %q but had %q", expected, to.Field)
	}
}

func TestScanFromPtrToSqlNullable(t *testing.T) {
	var (
		from struct {
			S     string
			Sptr  *string
			Snull sql.NullString
			T1    sql.NullTime
			T2    sql.NullTime
			T3    *time.Time
		}

		to struct {
			S     sql.NullString
			Sptr  sql.NullString
			Snull *string
			T1    time.Time
			T2    *time.Time
			T3    sql.NullTime
		}

		s string

		err error
	)

	s = "test"
	from.S = s
	from.Sptr = &s

	if from.T1.Valid || from.T2.Valid {
		t.Errorf("Must be not valid")
	}

	err = copier.Copy(&to, from)
	if err != nil {
		t.Error("Should not raise error")
	}

	if !to.T1.IsZero() {
		t.Errorf("to.T1 should be Zero but %v", to.T1)
	}

	if to.T2 != nil {
		t.Errorf("to.T2 should be nil but %v", to.T2)
	}

	if to.Snull != nil {
		t.Errorf("to.Snull should be nil but %v", to.Snull)
	}

	now := time.Now()

	err = from.T1.Scan(now)
	if err != nil {
		t.Error("should not error")
	}
	err = from.T2.Scan(now)
	if err != nil {
		t.Error("should not error")
	}

	err = copier.Copy(&to, from)
	if err != nil {
		t.Error("Should not raise error")
	}

	if to.S.String != from.S {
		t.Errorf("Field S should be copied")
	}

	if to.Sptr.String != *from.Sptr {
		t.Errorf("Field Sptr should be copied")
	}

	if from.T1.Time != to.T1 {
		t.Errorf("Fields T1 fields should be equal")
	}

	if from.T2.Time != *to.T2 {
		t.Errorf("Fields T2 fields should be equal")
	}
}

func TestDeepCopyInterface(t *testing.T) {
	m := make(map[string]string)
	m["a"] = "ccc"

	from := []interface{}{[]int{7, 8, 9}, 2, 3, m, errors.New("aaaa")}
	var to []interface{}

	err := copier.CopyWithOption(&to, &from, copier.Option{
		IgnoreEmpty: false,
		DeepCopy:    true,
	})
	if err != nil {
		t.Error("should not error")
	}

	from[0].([]int)[0] = 10
	from[1] = "3"
	from[3].(map[string]string)["a"] = "bbb"

	if fmt.Sprint(to[0]) != fmt.Sprint([]int{7, 8, 9}) {
		t.Errorf("to value failed to be deep copied")
	}

	if fmt.Sprint(to[1]) != "2" {
		t.Errorf("to value failed to be deep copied")
	}

	if to[3].(map[string]string)["a"] != "ccc" {
		t.Errorf("to value failed to be deep copied")
	}
}

func TestDeepCopyTime(t *testing.T) {
	type embedT1 struct {
		T5 time.Time
	}

	type embedT2 struct {
		T6 *time.Time
	}

	var (
		from struct {
			T1 time.Time
			T2 *time.Time

			T3 *time.Time
			T4 time.Time
			T5 time.Time
			T6 time.Time
		}

		to struct {
			T1 time.Time
			T2 *time.Time

			T3 time.Time
			T4 *time.Time
			embedT1
			embedT2
		}
	)

	t1 := time.Now()
	from.T1 = t1
	t2 := t1.Add(time.Second)
	from.T2 = &t2
	t3 := t2.Add(time.Second)
	from.T3 = &t3
	t4 := t3.Add(time.Second)
	from.T4 = t4
	t5 := t4.Add(time.Second)
	from.T5 = t5
	t6 := t5.Add(time.Second)
	from.T6 = t6

	err := copier.CopyWithOption(&to, from, copier.Option{DeepCopy: true})
	if err != nil {
		t.Error("Should not raise error")
	}

	if !to.T1.Equal(from.T1) {
		t.Errorf("Field T1 should be copied")
	}
	if !to.T2.Equal(*from.T2) {
		t.Errorf("Field T2 should be copied")
	}
	if !to.T3.Equal(*from.T3) {
		t.Errorf("Field T3 should be copied")
	}
	if !to.T4.Equal(from.T4) {
		t.Errorf("Field T4 should be copied")
	}
	if !to.T5.Equal(from.T5) {
		t.Errorf("Field T5 should be copied")
	}
	if !to.T6.Equal(from.T6) {
		t.Errorf("Field T6 should be copied")
	}
}

func TestNestedPrivateData(t *testing.T) {
	type hasPrivate struct {
		data int
	}

	type hasMembers struct {
		Member hasPrivate
	}

	src := hasMembers{
		Member: hasPrivate{
			data: 42,
		},
	}
	var shallow hasMembers
	err := copier.Copy(&shallow, &src)
	if err != nil {
		t.Errorf("could not complete shallow copy")
	}
	if !reflect.DeepEqual(&src, &shallow) {
		t.Errorf("shallow copy faild")
	}

	var deep hasMembers
	err = copier.CopyWithOption(&deep, &src, copier.Option{DeepCopy: true})
	if err != nil {
		t.Errorf("could not complete deep copy")
	}
	if !reflect.DeepEqual(&src, &deep) {
		t.Errorf("deep copy faild")
	}

	if !reflect.DeepEqual(&shallow, &deep) {
		t.Errorf("unexpected difference between shallow and deep copy")
	}
}

func TestDeepMapCopyTime(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(time.Second)
	from := []map[string]interface{}{
		{
			"t1": t1,
			"t2": &t2,
		},
	}
	to := make([]map[string]interface{}, len(from))

	err := copier.CopyWithOption(&to, from, copier.Option{DeepCopy: true})
	if err != nil {
		t.Error("should not error")
	}
	if len(to) != len(from) {
		t.Errorf("slice should be copied")
	}
	if !to[0]["t1"].(time.Time).Equal(from[0]["t1"].(time.Time)) {
		t.Errorf("nested time ptr should be copied")
	}
	if !to[0]["t2"].(*time.Time).Equal(*from[0]["t2"].(*time.Time)) {
		t.Errorf("nested time ptr should be copied")
	}
}

func TestCopySimpleTime(t *testing.T) {
	from := time.Now()
	to := time.Time{}

	err := copier.Copy(&to, from)
	if err != nil {
		t.Error("should not error")
	}
	if !from.Equal(to) {
		t.Errorf("to (%v) value should equal from (%v) value", to, from)
	}
}

func TestDeepCopySimpleTime(t *testing.T) {
	from := time.Now()
	to := time.Time{}

	err := copier.CopyWithOption(&to, from, copier.Option{DeepCopy: true})
	if err != nil {
		t.Error("should not error")
	}
	if !from.Equal(to) {
		t.Errorf("to (%v) value should equal from (%v) value", to, from)
	}
}

type TimeWrapper struct {
	time.Time
}

func TestDeepCopyAnonymousFieldTime(t *testing.T) {
	from := TimeWrapper{time.Now()}
	to := TimeWrapper{}

	err := copier.CopyWithOption(&to, from, copier.Option{DeepCopy: true})
	if err != nil {
		t.Error("should not error")
	}
	if !from.Equal(to.Time) {
		t.Errorf("to (%v) value should equal from (%v) value", to.Time, from.Time)
	}
}

func TestSqlNullFiled(t *testing.T) {
	type sqlStruct struct {
		MkExpiryDateStart sql.NullString
		MkID              sql.NullInt64
		MkExpiryDateType  sql.NullInt32
	}

	type dataStruct struct {
		MkExpiryDateStart string
		MkID              int64
		MkExpiryDateType  int32
	}

	from := sqlStruct{
		MkID:              sql.NullInt64{Int64: 3, Valid: true},
		MkExpiryDateType:  sql.NullInt32{Int32: 4, Valid: true},
		MkExpiryDateStart: sql.NullString{String: "5", Valid: true},
	}

	to := dataStruct{}

	err := copier.Copy(&to, from)
	if err != nil {
		t.Error("should not error")
	}
	if from.MkID.Int64 != to.MkID {
		t.Errorf("to (%v) value should equal from (%v) value", to.MkID, from.MkID.Int64)
	}

	if from.MkExpiryDateStart.String != to.MkExpiryDateStart {
		t.Errorf("to (%v) value should equal from (%v) value", to.MkExpiryDateStart, from.MkExpiryDateStart.String)
	}

	if from.MkExpiryDateType.Int32 != to.MkExpiryDateType {
		t.Errorf("to (%v) value should equal from (%v) value", to.MkExpiryDateType, from.MkExpiryDateType.Int32)
	}
}

func TestEmptySlice(t *testing.T) {
	type Str1 string
	type Str2 string
	type Input1 struct {
		Val Str1
	}
	type Input2 struct {
		Val Str2
	}
	to := []*Input1(nil)
	from := []*Input2{}
	err := copier.Copy(&to, &from)
	if err != nil {
		t.Error("should not error")
	}
	if from == nil {
		t.Error("from should be empty slice not nil")
	}

	to = []*Input1(nil)
	from = []*Input2(nil)
	err = copier.Copy(&to, &from)
	if err != nil {
		t.Error("should not error")
	}
	if from != nil {
		t.Error("from should be empty slice nil")
	}
}

func TestNestedNilPointerStruct(t *testing.T) {
	type destination struct {
		Title string
	}

	type NestedSource struct {
		ID int
	}

	type source struct {
		*NestedSource
		Title string
	}

	from := &source{
		Title: "A title to be copied",
	}

	to := destination{}

	err := copier.Copy(&to, from)
	if err != nil {
		t.Error("should not error")
	}

	if from.Title != to.Title {
		t.Errorf("to (%v) value should equal from (%v) value", to.Title, from.Title)
	}
}

type testValuer struct {
	Value interface{}
}

func (v testValuer) CopyValue() interface{} {
	return v.Value
}

func TestCopyValuer(t *testing.T) {
	to := struct {
		Value string
	}{
		Value: "initial",
	}

	from := struct {
		Value testValuer
	}{
		Value: testValuer{Value: "override"},
	}

	err := copier.Copy(&to, from)
	if err != nil {
		t.Errorf("should not error: %v", err)
	}

	if to.Value != from.Value.Value {
		t.Errorf("to (%v) value should equal to from (%v) value", to.Value, from.Value.Value)
	}

	toStr := ""
	fromStr := testValuer{Value: "override"}
	err = copier.Copy(&toStr, fromStr)
	if err != nil {
		t.Errorf("should not error: %v", err)
	}

	if toStr != fromStr.Value {
		t.Errorf("toStr (%v) value should equal to fromStr (%v) value", toStr, fromStr.Value)
	}
}

func TestMapIncompatibleKeys(t *testing.T) {
	to := map[struct{}]any{}

	from := map[int64]any{}

	err := copier.Copy(&to, from)
	if err == nil {
		t.Errorf("should error")
		return
	}
	if !errors.Is(err, copier.ErrMapKeyNotMatch) {
		t.Errorf("error should be ErrMapKeyNotMatch: %v", err)
	}
}

func TestDeepCopyShortSliceIntoLongSlice(t *testing.T) {
	type testStrct struct {
		Value []string
	}

	to := testStrct{
		Value: []string{"a", "b", "c", "d", "e"},
	}

	from := testStrct{
		Value: []string{"f", "g"},
	}

	err := copier.CopyWithOption(&to, from, copier.Option{DeepCopy: true})
	if err != nil {
		t.Errorf("should not error: %v", err)
	}

	if len(to.Value) != len(from.Value) { // Values "c", "d", "e" are not expected
		t.Errorf("to (%v) value len should equal to from (%v) value len", len(to.Value), len(from.Value))
	}
}

func TestDeepCopyShortStructSliceIntoLongSlice(t *testing.T) {
	type subStrctFrom struct {
		Name string
	}

	type subStrctTo struct {
		Name           string
		SomeOtherField string // Make it so the struct is not directly convertible
	}

	type testStrctFrom struct {
		Value []*subStrctFrom
	}

	type testStrctTo struct {
		Value []*subStrctTo
	}

	to := testStrctTo{
		Value: []*subStrctTo{{Name: "a"}, {Name: "b"}, {Name: "c"}, {Name: "d"}, {Name: "e"}},
	}

	from := testStrctFrom{
		Value: []*subStrctFrom{{"f"}, {"g"}},
	}

	err := copier.CopyWithOption(&to, from, copier.Option{DeepCopy: true})
	if err != nil {
		t.Errorf("should not error: %v", err)
	}

	if len(to.Value) != len(from.Value) { // Values "c", "d", "e" are not expected
		t.Errorf("to (%v) value len should equal to from (%v) value len", len(to.Value), len(from.Value))
	}
}

func TestDeepCopyShortMapIntoLongMap(t *testing.T) {
	type testStrct struct {
		Value map[string]string
	}

	to := testStrct{
		Value: map[string]string{"a": "b", "c": "d", "e": "f"},
	}

	from := testStrct{
		Value: map[string]string{"g": "h", "i": "j"},
	}

	err := copier.CopyWithOption(&to, from, copier.Option{DeepCopy: true})
	if err != nil {
		t.Errorf("should not error: %v", err)
	}

	if len(to.Value) != len(from.Value) { // keys "a", "c" and "e" are not expected
		t.Errorf("to (%v) value len should equal to from (%v) value len", len(to.Value), len(from.Value))
	}
}
