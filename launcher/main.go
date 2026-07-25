// ZenBrowserPortable.exe — нативный лаунчер с режимом «ноль следов».
//
// Делает три вещи:
//  1. Запускает App/Zen/zen.exe с профилем внутри Data/ (ничего не читает из системы).
//  2. Перенаправляет TEMP/TMP/APPDATA/LOCALAPPDATA внутрь Data/ для дочернего процесса.
//  3. Ждёт закрытия браузера и удаляет каталоги, которые Zen/Firefox могли создать
//     в системном AppData — но ТОЛЬКО те, которых не существовало до запуска
//     (чтобы не тронуть данные установленного Zen/Firefox на этой машине).
//
// Компилируется без внешних зависимостей: go build.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	exePath, err := os.Executable()
	if err != nil {
		fail(err)
	}
	root := filepath.Dir(exePath)

	app := filepath.Join(root, "App", "Zen", "zen.exe")
	profile := filepath.Join(root, "Data", "profile")
	temp := filepath.Join(root, "Data", "temp")
	appdataR := filepath.Join(root, "Data", "appdata", "Roaming")
	appdataL := filepath.Join(root, "Data", "appdata", "Local")

	if _, err := os.Stat(app); err != nil {
		fail(fmt.Errorf("zen.exe not found: %s", app))
	}

	for _, dir := range []string{profile, temp, appdataR, appdataL} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fail(err)
		}
	}

	// Снимок «до запуска»: каталоги в настоящем системном AppData, которые
	// браузер теоретически может создать. Удалим после выхода только те,
	// которых сейчас нет (существующие = чужие данные, их не трогаем).
	candidates := systemLeftoverCandidates()
	missingBefore := make([]string, 0, len(candidates))
	for _, p := range candidates {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			missingBefore = append(missingBefore, p)
		}
	}

	args := append([]string{"-profile", profile, "-no-remote"}, os.Args[1:]...)
	cmd := exec.Command(app, args...)
	cmd.Dir = filepath.Join(root, "App", "Zen")
	cmd.Env = append(os.Environ(),
		// Временные файлы — внутрь portable-папки
		"TEMP="+temp,
		"TMP="+temp,
		// Всё, что резолвится через переменные окружения, — тоже внутрь
		"APPDATA="+appdataR,
		"LOCALAPPDATA="+appdataL,
		// Крэш-репортер пишет дампы в %APPDATA%\zen\Crash Reports (вне portable) — выкл.
		"MOZ_CRASHREPORTER_DISABLE=1",
	)

	// Run (а не Start): ждём закрытия браузера, чтобы прибраться за ним.
	runErr := cmd.Run()

	// Уборка: удаляем только то, чего не было до запуска.
	for _, p := range missingBefore {
		_ = os.RemoveAll(p)
	}

	if runErr != nil {
		fail(runErr)
	}
}

// systemLeftoverCandidates — известные места, куда Firefox-база может
// создать (обычно пустые) каталоги даже в portable-режиме.
func systemLeftoverCandidates() []string {
	var out []string
	home, _ := os.UserHomeDir()
	roaming := os.Getenv("APPDATA")
	local := os.Getenv("LOCALAPPDATA")
	if roaming == "" && home != "" {
		roaming = filepath.Join(home, "AppData", "Roaming")
	}
	if local == "" && home != "" {
		local = filepath.Join(home, "AppData", "Local")
	}
	var locallow string
	if home != "" {
		locallow = filepath.Join(home, "AppData", "LocalLow")
	}
	for _, base := range []string{roaming, local, locallow} {
		if base == "" {
			continue
		}
		out = append(out,
			filepath.Join(base, "zen"),
			filepath.Join(base, "Mozilla"),
		)
	}
	return out
}

func fail(err error) {
	// Пишем лог рядом с лаунчером, т.к. GUI-приложение не имеет консоли.
	exePath, _ := os.Executable()
	logPath := filepath.Join(filepath.Dir(exePath), "launcher-error.log")
	_ = os.WriteFile(logPath, []byte(err.Error()+"\n"), 0o644)
	os.Exit(1)
}
