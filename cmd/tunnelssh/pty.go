package main

// tunnelssh PTY code

// https://github.com/google/goexpect/blob/master/expect.go
// https://github.com/google/goterm/blob/master/term/ssh.go

// TerminalMode mode
type TerminalMode int

// Mode
const (
	TerminalModeAuto TerminalMode = iota
	TerminalModeNone
	TerminalModeForce
)
