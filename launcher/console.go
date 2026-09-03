package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
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

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	user32               = syscall.NewLazyDLL("user32.dll")
	procGetConsoleMode   = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode   = kernel32.NewProc("SetConsoleMode")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	procShowWindow       = user32.NewProc("ShowWindow")
	procMessageBoxW      = user32.NewProc("MessageBoxW")
)

// enableColors включает ANSI escape-коды в консоли Windows (ENABLE_VIRTUAL_TERMINAL_PROCESSING).
// Без этого цвета печатались бы как мусорные символы на старых консолях.
func enableColors() {
	h := os.Stdout.Fd()
	var mode uint32
	r, _, _ := procGetConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode)))
	if r == 0 {
		return
	}
	const enableVirtualTerminalProcessing = 0x0004
	procSetConsoleMode.Call(h, uintptr(mode|enableVirtualTerminalProcessing))
}

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

// hideConsoleWindow прячет окно консоли (процесс продолжает жить и после этого —
// нужен, чтобы дождаться закрытия Zen и убрать за собой мусор в AppData).
func hideConsoleWindow() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd != 0 {
		const swHide = 0
		procShowWindow.Call(hwnd, swHide)
	}
}

// notify показывает системный MessageBox — используется, чтобы сообщить
// о готовом фоновом обновлении, когда консоль уже скрыта, а Zen на экране.
func notify(title, text string) {
	t, err1 := syscall.UTF16PtrFromString(title)
	m, err2 := syscall.UTF16PtrFromString(text)
	if err1 != nil || err2 != nil {
		return
	}
	const mbIconInformation = 0x40
	const mbTopmost = 0x40000
	procMessageBoxW.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), uintptr(mbIconInformation|mbTopmost))
}

func pauseForError() {
	fmt.Println()
	fmt.Println(ansiGray + "Нажмите Enter, чтобы закрыть..." + ansiReset)
	var discard string
	fmt.Scanln(&discard)
}
