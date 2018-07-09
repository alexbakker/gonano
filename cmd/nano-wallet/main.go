package main

import "syscall"

func main() {
	// set the umask of this process to 077
	// this ensures all written files are only readable/writable by the current user
	syscall.Umask(077)
}
