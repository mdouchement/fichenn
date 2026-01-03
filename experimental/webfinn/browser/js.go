package main

import "syscall/js"

func dump(o js.Value) {
	js.Global().Get("console").Call("log", o)
}
