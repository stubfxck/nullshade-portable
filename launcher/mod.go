// mod.go — опциональное обновление и установка мода приватных вкладок
// (github.com/stubfxck/nullshade-private-tab). Мод не встроен в сборку —
// включается флагом privateTabModEnabled в Data\launcher-config.json.
//
// Важный нюанс: обновление самого браузера (update.go) целиком заменяет
// App\Zen\ свежераспакованными официальными бинарниками — а значит стирает
// файлы мода, которые лежат внутри App\Zen (config.js и т.д.). Поэтому мод
// не просто проверяется на новую версию, а на КАЖДОМ запуске переналивается
// из локального кэша поверх App\Zen и Data\profile\chrome — это чинит его
// после любого обновления браузера, а не только после обновления самого мода.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type installedMods map[string]string // id мода -> установленная версия

func installedModsPath(dataDir string) string {
	return filepath.Join(dataDir, "profile", "chrome", ".nullshade-mods.json")
}

func readInstalledMods(dataDir string) installedMods {
	b, err := os.ReadFile(installedModsPath(dataDir))
	if err != nil {
		return installedMods{}
	}
	var m installedMods
	if json.Unmarshal(b, &m) != nil {
		return installedMods{}
	}
	return m
}

func writeInstalledMods(dataDir string, m installedMods) {
	path := installedModsPath(dataDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, b, 0o644)
}

func modCacheDir(dataDir, id string) string {
	return filepath.Join(dataDir, "mods-cache", id)
}

func findFirstZipAsset(rel *ghRelease) *ghAsset {
	for i := range rel.Assets {
		if strings.HasSuffix(rel.Assets[i].Name, ".zip") {
			return &rel.Assets[i]
		}
	}
	return nil
}

// checkAndInstallPrivateTabMod — вызывается только если privateTabModEnabled
// стоит true. Проверяет новую версию мода (сеть), при необходимости качает
// её в локальный кэш, а затем ВСЕГДА переналивает из кэша в App\Zen и
// Data\profile\chrome — дёшево (несколько мелких файлов), но чинит мод,
// если его стёр апдейт браузера с прошлого запуска.
func checkAndInstallPrivateTabMod(root, dataDir string) {
	installed := readInstalledMods(dataDir)
	cache := modCacheDir(dataDir, modID)

	rel, err := fetchLatestReleaseFor(modRepoOwner, modRepoName)
	if err != nil {
		warn("Не смог проверить обновление мода private-tab: " + err.Error())
	} else {
		latest := strings.TrimPrefix(rel.TagName, "v")
		if installed[modID] != latest {
			asset := findFirstZipAsset(rel)
			if asset == nil {
				warn("В релизе мода " + rel.TagName + " не нашлось zip.")
			} else {
				step(fmt.Sprintf("Мод private-tab: %s -> %s", displayOrNone(installed[modID]), latest))
				if err := downloadModToCache(cache, asset); err != nil {
					warn("Не смог скачать мод: " + err.Error())
				} else {
					installed[modID] = latest
					writeInstalledMods(dataDir, installed)
				}
			}
		}
	}

	if _, err := os.Stat(cache); err != nil {
		return // мод ещё ни разу не скачался (нет сети при первом включении флага) — нечего накатывать
	}
	if err := applyModFromCache(root, dataDir, cache); err != nil {
		warn("Не смог применить мод private-tab: " + err.Error())
		return
	}
	ok("Мод private-tab установлен (" + installed[modID] + ").")
}

func displayOrNone(v string) string {
	if v == "" {
		return "не установлен"
	}
	return v
}

func downloadModToCache(cache string, asset *ghAsset) error {
	tmp := cache + ".download"
	os.RemoveAll(tmp)
	defer os.RemoveAll(tmp)
	zipPath := filepath.Join(tmp, "mod.zip")

	if err := downloadFile(asset.BrowserDownloadURL, zipPath, asset.Size, false); err != nil {
		return err
	}

	fresh := cache + ".new"
	os.RemoveAll(fresh)
	if err := extractZipPrefix(zipPath, "payload/", fresh); err != nil {
		os.RemoveAll(fresh)
		return err
	}

	os.RemoveAll(cache)
	if err := os.Rename(fresh, cache); err != nil {
		return err
	}
	return nil
}

// applyModFromCache копирует уже скачанные файлы мода поверх App\Zen и
// Data\profile\chrome. Не трогает остальное содержимое этих папок —
// только накладывает файлы мода (create/overwrite, без удаления чужого).
func applyModFromCache(root, dataDir, cache string) error {
	appOverlay := filepath.Join(cache, "app-overlay")
	profileOverlay := filepath.Join(cache, "profile-overlay", "chrome")

	if _, err := os.Stat(appOverlay); err == nil {
		if err := copyTreeOverlay(appOverlay, filepath.Join(root, "App", "Zen")); err != nil {
			return err
		}
	}
	if _, err := os.Stat(profileOverlay); err == nil {
		chromeDir := filepath.Join(dataDir, "profile", "chrome")
		if err := copyTreeOverlay(profileOverlay, chromeDir); err != nil {
			return err
		}
	}
	return nil
}

// copyTreeOverlay рекурсивно копирует src поверх dst: создаёт недостающие
// директории, перезаписывает файлы, ничего не удаляет.
func copyTreeOverlay(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}
