package main

import (
	"fmt"
	"time"
)

func main() {
	a := time.Now()
	json, err := a.MarshalJSON()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(json))
}
