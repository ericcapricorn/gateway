package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"zc-common-go/common"
)

var i int

// test close channel
func test_close() (err error) {
	ch := make(chan int, 10)
	go func() {
		select {
		case test := <-ch:
			fmt.Println("hahahaha", test)
		}
	}()

	close(ch)
	defer func() {
		info := recover()
		if info != nil {
			fmt.Printf("the request queue is closed:err[%v]\n", info)
			err = errors.New("closed")
		}
	}()
	// nil if not write the closed device
	ch <- 1
	return err
}

func main() {
	err := test_panic(false)
	if err != nil {
		fmt.Println("fatal error:", err)
	}
	err = test_panic(true)
	fmt.Println("modify the panic error:", err)
	err = test_panic(true)
	fmt.Println("modify the panic error:", err)

	// closed channel panic
	err = test_close()
	fmt.Println(err)

	// timeout
	result := make(chan int, 60)
	go test_channel(result)

	for i := 0; i < 10; i++ {
		result <- 1
		time.Sleep(time.Second)
	}
}

func test_channel(result chan int) {
	timer := time.After(time.Second * 5)
	for {
		select {
		case <-result:
			fmt.Println("receive a message", time.Now())
		case <-timer:
			fmt.Println("wait timeout", time.Now())
			return
		}
	}
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
