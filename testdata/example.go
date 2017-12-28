package main

func foo(
	buf []byte, // owned
) {
}

func main() {
	var b []byte
	foo(b)

	// this operation is illegal because main no longer owns b
	foo(b)
}
