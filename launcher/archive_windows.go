//go:build windows

package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// extractUpdateZip распаковывает из скачанного релиза (.zip на Windows)
// только то, что нужно для обновления: содержимое App/Zen/ (в appZenTarget),
// сам лаунчер и файлы Support/*. Data/ и прочее из архива не трогаем.
//
// Поддержка старых установок: если в архиве почему-то нет Support/ (старый
// билдер), version.json ищем и в корне архива — так уже обновлённый лаунчер
// не спотыкается о зип, собранный до переезда в Support\.
func extractUpdateZip(zipPath, appZenTarget string) (*updatePayload, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	if err := os.RemoveAll(appZenTarget); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(appZenTarget, 0o755); err != nil {
		return nil, err
	}

	payload := &updatePayload{support: map[string][]byte{}}

	for _, f := range r.File {
		name := strings.ReplaceAll(f.Name, "\\", "/")
		switch {
		case name == launcherAssetName:
			payload.exe, err = readZipFile(f)
			if err != nil {
				return nil, err
			}
		case name == "Support/version.json", name == "version.json":
			payload.version, err = readZipFile(f)
			if err != nil {
				return nil, err
			}
		case strings.HasPrefix(name, "Support/"):
			rel := strings.TrimPrefix(name, "Support/")
			if rel == "" || f.FileInfo().IsDir() {
				continue
			}
			b, err := readZipFile(f)
			if err != nil {
				return nil, err
			}
			payload.support[rel] = b
		case strings.HasPrefix(name, "App/Zen/"):
			rel := strings.TrimPrefix(name, "App/Zen/")
			if rel == "" {
				continue
			}
			target := filepath.Join(appZenTarget, filepath.FromSlash(rel))
			if f.FileInfo().IsDir() {
				if err := os.MkdirAll(target, 0o755); err != nil {
					return nil, err
				}
				continue
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return nil, err
			}
			if err := extractZipFileTo(f, target); err != nil {
				return nil, err
			}
		}
	}
	if payload.exe == nil {
		return nil, fmt.Errorf(t("в архиве не нашёлся %s", "%s not found in the archive"), launcherAssetName)
	}
	return payload, nil
}
