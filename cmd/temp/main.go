package main

import (
	"fmt"
	"strings"
	"time"
)


func main(){
	fmt.Printf("%s\n", time.Now().Format(time.RFC3339))
	fmt.Println(strings.Split("n23","n")[1])
	body := map[string]bool{
		"k2": true,
	}

	fmt.Println(body["k2"])
	fmt.Println(body["k3"])
}
