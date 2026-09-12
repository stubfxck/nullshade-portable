package main

import (
	"fmt"
	"time"
)

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiGreen  = "\x1b[38;5;46m"
	ansiCyan   = "\x1b[38;5;51m"
	ansiYellow = "\x1b[38;5;226m"
	ansiRed    = "\x1b[38;5;196m"
	ansiGray   = "\x1b[38;5;240m"
)

// enableColors/hideConsoleWindow/notify — платформенно-специфичные, см.
// console_windows.go и console_linux.go.

func banner() {
	fmt.Println(ansiGreen + ansiBold + "========================================" + ansiReset)
	fmt.Println(ansiGreen + ansiBold + "   Z E N   B R O W S E R   P O R T A B L E" + ansiReset)
	fmt.Println(ansiGreen + ansiBold + "========================================" + ansiReset)
	fmt.Println()
}

func step(msg string) {
	fmt.Println(ansiCyan + "> " + ansiReset + msg)
	time.Sleep(180 * time.Millisecond)
}

func ok(msg string) {
	fmt.Println(ansiGreen + "[ok] " + ansiReset + msg)
	time.Sleep(180 * time.Millisecond)
}

func warn(msg string) {
	fmt.Println(ansiYellow + "[!] " + ansiReset + msg)
}

func errLine(msg string) {
	fmt.Println(ansiRed + "[x] " + ansiReset + msg)
}

func pauseForError() {
	fmt.Println()
	fmt.Println(ansiGray + t("Нажмите Enter, чтобы закрыть...", "Press Enter to close...") + ansiReset)
	var discard string
	fmt.Scanln(&discard)
}
