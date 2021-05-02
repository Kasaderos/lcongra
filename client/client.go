package main

import "fmt"

type A struct {
	a chan string
}

func (a *A) f() {
	fmt.Println("got")
	<-a.a
}

func main() {
	ch := make(chan string, 1)
	a := A{ch}
	b := a
	_ = b
	ch <- "1"
	a.f()
}
