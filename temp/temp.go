package main

import (
	"fmt"
	"strconv"
	"time"
	"zc-gateway/common"
)

var i int

func main() {
	err := test_panic(false)
	if err != nil {
		fmt.Println("fatal error:", err)
	}
	err = test_panic(true)
	fmt.Println("modify the panic error:", err)
	err = test_panic(true)
	fmt.Println("modify the panic error:", err)
}

func test_panic(on bool) (err error) {
	defer func() {
		temp := recover()
		if temp != nil {
			fmt.Println("catch the panic:", temp)
			err = common.ErrInvalidMsg
		}
	}()
	if on {
		err = common.ErrEntryExist
		panic("panic on so happend")
	}
	return
}

func test() {
	msg := "abc" + "cde" + strconv.Itoa(124)
	fmt.Println(msg)
	var i int
	go func() {
		for i < 100 {
			i++
			fmt.Println("routine 1:", i)
			time.Sleep(time.Second)
		}
	}()
	go func(j int) {
		for j < 200 {
			j++
			fmt.Println("routine 2:", j, i)
			time.Sleep(time.Second / 2)
		}
	}(i)
	fmt.Println("main:", i)
	time.Sleep(10 * time.Second)
	fmt.Println("main:", i)
}
