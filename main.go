package main

import "fmt"

func printAll(args ...string) {
	for _, arg := range args {
		fmt.Println(arg)
	}
}

func main() {
	printAll("Hello")
	printAll("Hello", "World", "!")
}
