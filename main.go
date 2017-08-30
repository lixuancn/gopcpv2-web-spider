package main

import (
	"fmt"
)

func main(){
	var list map[string]map[string]int
	list = map[string]map[string]int{}
	list["a"] = map[string]int{"b":10, "c":20, "d":30}
	fmt.Println(list)
	for _, v := range list{
		delete(v, "b")
	}
	fmt.Println(list)

}