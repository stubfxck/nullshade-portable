package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	update "github.com/inconshreveable/go-update"
)

const (
	repoOwner = "stubfxck"
	repoName  = "zen-browser-portable"
)

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

func fetchLatestRelease() (*ghRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ZenBrowserPortable-Launcher")
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API вернул статус %d", resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func pickZipAsset(rel *ghRelease, arch string) (*ghAsset, error) {
	for i := range rel.Assets {
		a := &rel.Assets[i]
		if strings.HasSuffix(a.Name, ".zip") && strings.Contains(a.Name, arch) {
			return a, nil
		}
	}
	return nil, fmt.Errorf("не нашёл .zip для архитектуры %q в релизе %s", arch, rel.TagName)
}

// downloadFile качает url в dest. При showProgress печатает процент прямо
// в текущую строку консоли (используется только в блокирующем режиме —
// в фоновом режиме консоль уже скрыта, печатать некому).
func downloadFile(url, dest string, size int64, showProgress bool) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "ZenBrowserPortable-Launcher")

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("скачивание вернуло статус %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	var src io.Reader = resp.Body
	if showProgress {
		pw := &progressWriter{total: size}
		src = io.TeeReader(resp.Body, pw)
		defer pw.finish()
	}
	_, err = io.Copy(out, src)
	return err
}

type progressWriter struct {
	total   int64
	written int64
	last    time.Time
}

func (w *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.written += int64(n)
	if time.Since(w.last) > 200*time.Millisecond {
		w.last = time.Now()
		w.print()
	}
	return n, nil
}

func (w *progressWriter) print() {
	pct := 0
	if w.total > 0 {
		pct = int(w.written * 100 / w.total)
	}
	fmt.Printf("\r"+ansiCyan+"> "+ansiReset+"Скачиваю... %3d%% (%.1f/%.1f МБ)  ",
		pct, float64(w.written)/1024/1024, float64(w.total)/1024/1024)
}

func (w *progressWriter) finish() {
	w.written = w.total
	w.print()
	fmt.Println()
}

// extractUpdateZip распаковывает из скачанного релиза только то, что нужно
// для обновления: содержимое App/Zen/ (в appZenTarget), сам лаунчер
// (ZenBrowserPortable.exe) и version.json. Data/ и прочее из архива не трогаем.
func extractUpdateZip(zipPath, appZenTarget string) (exeBytes []byte, versionBytes []byte, err error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, nil, err
	}
	defer r.Close()

	if err := os.RemoveAll(appZenTarget); err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(appZenTarget, 0o755); err != nil {
		return nil, nil, err
	}

	for _, f := range r.File {
		name := strings.ReplaceAll(f.Name, "\\", "/")
		switch {
		case name == "ZenBrowserPortable.exe":
			exeBytes, err = readZipFile(f)
			if err != nil {
				return nil, nil, err
			}
		case name == "version.json":
			versionBytes, err = readZipFile(f)
			if err != nil {
				return nil, nil, err
			}
		case strings.HasPrefix(name, "App/Zen/"):
			rel := strings.TrimPrefix(name, "App/Zen/")
			if rel == "" {
				continue
			}
			target := filepath.Join(appZenTarget, filepath.FromSlash(rel))
			if f.FileInfo().IsDir() {
				if err := os.MkdirAll(target, 0o755); err != nil {
					return nil, nil, err
				}
				continue
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return nil, nil, err
			}
			if err := extractZipFileTo(f, target); err != nil {
				return nil, nil, err
			}
		}
	}
	if exeBytes == nil {
		return nil, nil, fmt.Errorf("в архиве не нашёлся ZenBrowserPortable.exe")
	}
	return exeBytes, versionBytes, nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func extractZipFileTo(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	mode := f.Mode()
	if mode == 0 {
		mode = 0o644
	}
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}

// swapAppZen атомарно (насколько это возможно на NTFS через rename) подменяет
// App\Zen на распакованную новую версию. Старую версию переименовывает в
// App\Zen.old и удаляет только после успешной подмены — если что-то пошло
// не так, откатывается.
func swapAppZen(root, stagedAppZen string) error {
	target := filepath.Join(root, "App", "Zen")
	backup := filepath.Join(root, "App", "Zen.old")
	os.RemoveAll(backup)

	if _, err := os.Stat(target); err == nil {
		if err := os.Rename(target, backup); err != nil {
			return fmt.Errorf("не смог отложить старую версию (закрой браузер и попробуй снова): %w", err)
		}
	}
	if err := os.Rename(stagedAppZen, target); err != nil {
		os.Rename(backup, target) // откат
		return err
	}
	os.RemoveAll(backup)
	return nil
}

// applySelfUpdate безопасно подменяет сам файл лаунчера (текущий исполняемый
// файл) на диске — через github.com/inconshreveable/go-update. Библиотека
// сама разруливает трюк с заменой файла работающего процесса на Windows
// (запись во временный файл рядом + rename поверх), плюс проверяет checksum.
func applySelfUpdate(exeBytes []byte) error {
	checksum := sha256.Sum256(exeBytes)
	err := update.Apply(bytes.NewReader(exeBytes), update.Options{
		Checksum: checksum[:],
	})
	if err != nil {
		if rerr := update.RollbackError(err); rerr != nil {
			return fmt.Errorf("откат после неудачного обновления тоже не удался: %v (исходная ошибка: %v)", rerr, err)
		}
		return err
	}
	return nil
}

// checkAndHandleUpdate сверяет локальную версию с последним релизом на GitHub
// и, в зависимости от cfg.UpdateMode, либо ставит обновление сразу (блокируя
// запуск), либо запускает Zen немедленно и качает обновление в фоне.
func checkAndHandleUpdate(root, dataDir string, ver *localVersion, cfg launcherConfig) {
	step("Проверяю обновления...")
	rel, err := fetchLatestRelease()
	if err != nil {
		warn("Не смог проверить обновления (нет сети или GitHub недоступен): " + err.Error())
		return
	}
	if rel.TagName == ver.PortableTag {
		ok("Версии совпадают (" + ver.PortableTag + ")")
		return
	}

	asset, err := pickZipAsset(rel, ver.Arch)
	if err != nil {
		warn(err.Error())
		return
	}

	step(fmt.Sprintf("Найдено обновление: %s -> %s", ver.PortableTag, rel.TagName))

	if cfg.UpdateMode == "block" {
		if err := downloadAndApplyBlocking(root, dataDir, asset, rel.TagName); err != nil {
			warn("Не смог установить обновление: " + err.Error() + " — запускаю текущую версию.")
			return
		}
		ok("Обновление " + rel.TagName + " установлено.")
		return
	}

	// background: не задерживаем пользователя, качаем параллельно с работой браузера.
	step("Скачаю обновление в фоне, пока пользуешься браузером.")
	go func() {
		if err := downloadAndStageUpdate(dataDir, asset, rel.TagName); err != nil {
			return // тихо: попробуем снова при следующем запуске
		}
		notify("Zen Browser Portable",
			"Обновление "+rel.TagName+" скачано.\nУстановится автоматически при следующем запуске.")
	}()
}

func downloadAndApplyBlocking(root, dataDir string, asset *ghAsset, tag string) error {
	tmpDir := filepath.Join(dataDir, "temp", "update")
	os.RemoveAll(tmpDir)
	defer os.RemoveAll(tmpDir)
	zipPath := filepath.Join(tmpDir, "update.zip")

	if err := downloadFile(asset.BrowserDownloadURL, zipPath, asset.Size, true); err != nil {
		return err
	}

	step("Распаковываю...")
	stagedAppZen := filepath.Join(tmpDir, "App-Zen-new")
	exeBytes, versionBytes, err := extractUpdateZip(zipPath, stagedAppZen)
	if err != nil {
		return err
	}

	step("Устанавливаю...")
	if err := swapAppZen(root, stagedAppZen); err != nil {
		return err
	}
	if versionBytes != nil {
		_ = os.WriteFile(filepath.Join(root, "version.json"), versionBytes, 0o644)
	}
	if err := applySelfUpdate(exeBytes); err != nil {
		warn("Браузер обновлён, но не смог обновить сам лаунчер: " + err.Error())
	}
	return nil
}

// downloadAndStageUpdate качает и распаковывает обновление в Data\pending-update\,
// НЕ применяя его сразу — применится при следующем запуске (applyPendingUpdateIfAny),
// чтобы не трогать файлы, пока Zen может их использовать.
func downloadAndStageUpdate(dataDir string, asset *ghAsset, tag string) error {
	pending := filepath.Join(dataDir, "pending-update")
	os.RemoveAll(pending)
	zipPath := filepath.Join(dataDir, "temp", "pending-update.zip")
	defer os.Remove(zipPath)

	if err := downloadFile(asset.BrowserDownloadURL, zipPath, asset.Size, false); err != nil {
		return err
	}

	stagedAppZen := filepath.Join(pending, "App-Zen")
	exeBytes, versionBytes, err := extractUpdateZip(zipPath, stagedAppZen)
	if err != nil {
		os.RemoveAll(pending)
		return err
	}
	if err := os.WriteFile(filepath.Join(pending, "launcher.exe"), exeBytes, 0o644); err != nil {
		os.RemoveAll(pending)
		return err
	}
	if versionBytes != nil {
		_ = os.WriteFile(filepath.Join(pending, "version.json"), versionBytes, 0o644)
	}

	manifest, _ := json.Marshal(map[string]string{"tag": tag})
	if err := os.WriteFile(filepath.Join(pending, "manifest.json"), manifest, 0o644); err != nil {
		os.RemoveAll(pending)
		return err
	}
	return nil
}

// applyPendingUpdateIfAny ставит обновление, скачанное в фоне при прошлом
// запуске (см. downloadAndStageUpdate). Вызывается в начале main(), до
// проверки сети — установка из уже скачанных файлов занимает секунды.
func applyPendingUpdateIfAny(root, dataDir string) {
	pending := filepath.Join(dataDir, "pending-update")
	manifestPath := filepath.Join(pending, "manifest.json")

	mb, err := os.ReadFile(manifestPath)
	if err != nil {
		return // нет отложенного обновления — обычный запуск
	}
	var manifest struct {
		Tag string `json:"tag"`
	}
	if json.Unmarshal(mb, &manifest) != nil {
		os.RemoveAll(pending)
		return
	}

	step("Устанавливаю ранее скачанное обновление (" + manifest.Tag + ")...")

	stagedAppZen := filepath.Join(pending, "App-Zen")
	if _, err := os.Stat(stagedAppZen); err != nil {
		warn("Отложенное обновление повреждено, пропускаю.")
		os.RemoveAll(pending)
		return
	}
	if err := swapAppZen(root, stagedAppZen); err != nil {
		warn("Не смог установить отложенное обновление: " + err.Error())
		return // pending не чистим — попробуем ещё раз в следующий запуск
	}
	if vb, err := os.ReadFile(filepath.Join(pending, "version.json")); err == nil {
		_ = os.WriteFile(filepath.Join(root, "version.json"), vb, 0o644)
	}
	if exeBytes, err := os.ReadFile(filepath.Join(pending, "launcher.exe")); err == nil {
		if err := applySelfUpdate(exeBytes); err != nil {
			warn("Не смог обновить сам лаунчер: " + err.Error())
		}
	}
	os.RemoveAll(pending)
	ok("Обновление " + manifest.Tag + " установлено.")
}
