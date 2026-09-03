// ZenBrowserPortable.exe — нативный лаунчер с режимом «ноль следов» и автообновлением.
//
// При каждом запуске:
//  1. Если раньше в фоне докачалось обновление — ставит его первым делом (быстро, локально).
//  2. Проверяет GitHub Releases на новую версию (можно выключить в Data\launcher-config.json).
//  3. Запускает App/Zen/zen.exe с профилем внутри Data/, ничего не читая из системы.
//  4. Ждёт закрытия браузера и убирает случайно созданные каталоги в системном AppData —
//     но ТОЛЬКО те, которых не существовало до запуска.
//
// Компилируется без CGO: go build.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	enableColors()
	banner()

	exePath, err := os.Executable()
	if err != nil {
		fail(err)
	}
	root := filepath.Dir(exePath)
	dataDir := filepath.Join(root, "Data")
	appZenExe := filepath.Join(root, "App", "Zen", "zen.exe")

	// Support\ (bat-запуск, version.json) не для обычного использования —
	// прячем каждый раз, на случай если архиватор при распаковке не сохранил
	// атрибут (Compress-Archive/7-Zip это не гарантируют).
	hideSupportDir(root)

	_ = os.MkdirAll(dataDir, 0o755)
	cfg := loadOrCreateConfig(dataDir)

	applyPendingUpdateIfAny(root, dataDir)

	if cfg.AutoUpdateEnabled {
		if ver, err := readVersionJSON(root); err == nil {
			checkAndHandleUpdate(root, dataDir, ver, cfg)
		} else {
			warn("version.json не найден — пропускаю проверку обновлений (ручная/локальная сборка?).")
		}
	} else {
		step("Автообновление выключено (Data\\launcher-config.json).")
	}

	if _, err := os.Stat(appZenExe); err != nil {
		fail(fmt.Errorf("zen.exe не найден: %s", appZenExe))
	}

	step("Запускаю Zen...")
	hideConsoleWindow()
	runZen(root, appZenExe, dataDir)
}

func runZen(root, app, dataDir string) {
	profile := filepath.Join(dataDir, "profile")
	temp := filepath.Join(dataDir, "temp")
	appdataR := filepath.Join(dataDir, "appdata", "Roaming")
	appdataL := filepath.Join(dataDir, "appdata", "Local")

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

	// Run (а не Start): ждём закрытия браузера, чтобы прибраться за ним
	// и чтобы фоновая докачка обновления (если она идёт) успела дожить до конца.
	runErr := cmd.Run()

	// Уборка: удаляем только то, чего не было до запуска.
	for _, p := range missingBefore {
		_ = os.RemoveAll(p)
	}

	if runErr != nil {
		fail(runErr)
	}
}

// hideSupportDir ставит атрибут Hidden на папку Support\. Не критично, если
// не получится (например, папки ещё нет при самой первой распаковке) —
// молча пропускаем.
func hideSupportDir(root string) {
	dir := filepath.Join(root, "Support")
	if _, err := os.Stat(dir); err != nil {
		return
	}
	p, err := syscall.UTF16PtrFromString(dir)
	if err != nil {
		return
	}
	_ = syscall.SetFileAttributes(p, syscall.FILE_ATTRIBUTE_HIDDEN)
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
	errLine(err.Error())
	// Дублируем в файл на случай, если консоль почему-то не видна (например,
	// EXE запущен нестандартным способом).
	exePath, _ := os.Executable()
	logPath := filepath.Join(filepath.Dir(exePath), "launcher-error.log")
	_ = os.WriteFile(logPath, []byte(err.Error()+"\n"), 0o644)
	pauseForError()
	os.Exit(1)
}
