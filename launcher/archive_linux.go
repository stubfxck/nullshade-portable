//go:build linux

package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// extractUpdateZip распаковывает из скачанного релиза (.tar.gz на Linux —
// см. releaseFileExt в platform_linux.go) только то, что нужно для
// обновления: содержимое App/Zen/ (в appZenTarget), сам лаунчер и файлы
// Support/*. Имя оставлено как на Windows (extractUpdateZip) ради общего
// call site в update.go — платформа сама решает, зип это или тарбол.
//
// tar-заголовок несёт unix-права файла (Mode) напрямую — в отличие от zip,
// где это зависит от того, чем архив создавался, тут двусмысленности нет:
// восстанавливаем ровно то, что записано, поэтому исполняемый бит на
// новом zen/новом лаунчере переживает распаковку без отдельного chmod.
func extractUpdateZip(archivePath, appZenTarget string) (*updatePayload, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	if err := os.RemoveAll(appZenTarget); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(appZenTarget, 0o755); err != nil {
		return nil, err
	}

	payload := &updatePayload{support: map[string][]byte{}}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		name := strings.TrimPrefix(strings.ReplaceAll(hdr.Name, "\\", "/"), "./")
		switch {
		case name == launcherAssetName:
			b, err := io.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			payload.exe = b
		case name == "Support/version.json", name == "version.json":
			b, err := io.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			payload.version = b
		case strings.HasPrefix(name, "Support/"):
			rel := strings.TrimPrefix(name, "Support/")
			if rel == "" || hdr.Typeflag == tar.TypeDir {
				continue
			}
			b, err := io.ReadAll(tr)
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
			if hdr.Typeflag == tar.TypeDir {
				if err := os.MkdirAll(target, 0o755); err != nil {
					return nil, err
				}
				continue
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return nil, err
			}
			mode := os.FileMode(hdr.Mode)
			if mode == 0 {
				mode = 0o644
			}
			out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
			if err != nil {
				return nil, err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return nil, err
			}
			out.Close()
		}
	}
	if payload.exe == nil {
		return nil, fmt.Errorf(t("в архиве не нашёлся %s", "%s not found in the archive"), launcherAssetName)
	}
	return payload, nil
}
