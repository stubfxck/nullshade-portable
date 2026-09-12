//go:build windows

package main

import (
	"os"
	"path/filepath"
	"syscall"
)

const (
	// Имя исполняемого файла самого Zen внутри App\Zen\.
	zenBinaryName = "zen.exe"
	// Имя лаунчера, которое ищем внутри скачанного релиза для самообновления.
	launcherAssetName = "ZenBrowserPortable.exe"
	// Расширение файла релиза на GitHub для этой платформы.
	releaseFileExt = ".zip"
)

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

// runtimeExtraDirs — дополнительные каталоги (кроме profile/temp), которые
// нужно создать в Data\ до запуска Zen, чтобы ниже было куда указать через
// переменные окружения.
func runtimeExtraDirs(dataDir string) []string {
	return []string{
		filepath.Join(dataDir, "appdata", "Roaming"),
		filepath.Join(dataDir, "appdata", "Local"),
	}
}

// runtimeEnvOverrides — переменные окружения, которые направляют Zen писать
// AppData-подобные и временные файлы внутрь portable-папки вместо системных путей.
func runtimeEnvOverrides(dataDir string) []string {
	temp := filepath.Join(dataDir, "temp")
	appdataR := filepath.Join(dataDir, "appdata", "Roaming")
	appdataL := filepath.Join(dataDir, "appdata", "Local")
	return []string{
		"TEMP=" + temp,
		"TMP=" + temp,
		"APPDATA=" + appdataR,
		"LOCALAPPDATA=" + appdataL,
		// Крэш-репортер пишет дампы в %APPDATA%\zen\Crash Reports (вне portable) — выкл.
		"MOZ_CRASHREPORTER_DISABLE=1",
	}
}
