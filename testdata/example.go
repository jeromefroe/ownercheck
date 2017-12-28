package main

func foo(
	buf []byte, // owned
) {
}

func main() {
	var b []byte
	foo(b)

	// illegal
	foo(b)
}
