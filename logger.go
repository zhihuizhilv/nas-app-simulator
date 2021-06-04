package main

import "fmt"

type Loger struct {
	funName string
}

func (l *Loger) Begin(name string) *Loger {
	fmt.Println(name + " begin")
	return l
}

func (l *Loger) End() {
	fmt.Println(l.funName + " begin")
}

func NewFuncTracer(name string) *Loger {
	var l Loger
	l.Begin(name)
	return &l
}