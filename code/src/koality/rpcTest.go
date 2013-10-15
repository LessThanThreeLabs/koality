package main

import (
	"fmt"
	"koality/resources/rpc"
	"reflect"
)

func main() {
	connection := rpc.NewConnection()

	fmt.Printf("Getting user 17...\n")
	user, err := connection.Users.Read.Get(17)
	if err != nil {
		fmt.Println("Error: %v", err)
	} else {
		fmt.Println(user)
	}
}

func testReflection() {
	c := map[string]int{"there": 14}
	arr := []interface{}{7, "hello", c}
	fmt.Println(arr)

	blah := Blah(7)

	method, _ := reflect.TypeOf(blah).MethodByName("hello")
	fmt.Println(method.Func.Type().In(1))

	arr2 := []reflect.Value{
		reflect.ValueOf(arr[0]),
		reflect.ValueOf(arr[1]),
	}
	method.Func.Call(arr2)
}

type Blah int

func (blah Blah) hello(a int, b string) {
	fmt.Println("hello there!")
}
