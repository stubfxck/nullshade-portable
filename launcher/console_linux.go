//go:build linux

package main

import "os/exec"

// enableColors на Linux — no-op: терминалы здесь понимают ANSI-коды из
// коробки, отдельно включать нечего (в отличие от старого conhost на Windows).
func enableColors() {}

// hideConsoleWindow на Linux — no-op: лаунчер тут не порождает отдельное окно
// консоли поверх какого-то GUI-процесса, он просто выполняется в том
// терминале, из которого его запустили (или из .desktop-файла — тогда
// терминала не видно вообще, скрывать нечего).
func hideConsoleWindow() {}

// notify — best-effort системное уведомление через notify-send (libnotify),
// если оно установлено. На системах без него (минимальные WM без D-Bus/
// notification daemon) просто тихо ничего не показывает — это не критично,
// у пользователя и так есть терминал с выводом лаунчера.
func notify(title, text string) {
	_ = exec.Command("notify-send", title, text).Run()
}
