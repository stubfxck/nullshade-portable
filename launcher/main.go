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
	appZenExe := filepath.Join(root, "App", "Zen", zenBinaryName)

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
			warn(t(
				"version.json не найден — пропускаю проверку обновлений (ручная/локальная сборка?).",
				"version.json not found — skipping update check (manual/local build?).",
			))
		}
	} else {
		step(t("Автообновление выключено (Data\\launcher-config.json).", "Auto-update disabled (Data\\launcher-config.json)."))
	}

	if cfg.PrivateTabModEnabled {
		checkAndInstallPrivateTabMod(root, dataDir)
	}

	if _, err := os.Stat(appZenExe); err != nil {
		fail(fmt.Errorf(t("%s не найден: %s", "%s not found: %s"), zenBinaryName, appZenExe))
	}

	step(t("Запускаю Zen...", "Launching Zen..."))
	hideConsoleWindow()
	runZen(root, appZenExe, dataDir)
}

func runZen(root, app, dataDir string) {
	profile := filepath.Join(dataDir, "profile")
	temp := filepath.Join(dataDir, "temp")

	dirs := append([]string{profile, temp}, runtimeExtraDirs(dataDir)...)
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fail(err)
		}
	}

	// Снимок «до запуска»: каталоги вне portable-папки, которые браузер
	// теоретически может создать (см. systemLeftoverCandidates в
	// platform_windows.go/platform_linux.go). Удалим после выхода только те,
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
	cmd.Dir = filepath.Dir(app)
	cmd.Env = append(os.Environ(), runtimeEnvOverrides(dataDir)...)

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
