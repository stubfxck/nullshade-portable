//go:build linux

package main

import (
	"os"
	"path/filepath"
)

const (
	// Имя исполняемого файла самого Zen внутри App/Zen/ (официальный
	// linux-x86_64/-aarch64 тарбол zen-browser/desktop называет бинарник
	// просто "zen", без расширения).
	zenBinaryName = "zen"
	// Имя лаунчера, которое ищем внутри скачанного релиза для самообновления.
	launcherAssetName = "ZenBrowserPortable"
	// Расширение файла релиза на GitHub для этой платформы — tar.gz, а не
	// zip: сохраняет unix-права (бит "исполняемый") нативно и без сюрпризов
	// в любом архиваторе, в отличие от zip, где это зависит от конкретной
	// реализации инструмента, который его создал/распаковал.
	releaseFileExt = ".tar.gz"
)

// hideSupportDir на Linux — no-op. Атрибут "скрытый файл" через точку в
// начале имени тут не то же самое, что Windows FILE_ATTRIBUTE_HIDDEN
// (переименование сломало бы все пути, которые ссылаются на Support\ по
// имени), и на Linux никто не ждёт, что рядом лежащая папка будет спрятана —
// это не соответствует привычкам пользователей этой платформы.
func hideSupportDir(root string) {}

// systemLeftoverCandidates — Firefox/Zen на Linux при явном -profile почти
// не пишет мимо профиля, но кэш и крэш-репорты по умолчанию всё равно целятся
// в $XDG_CACHE_HOME (обычно ~/.cache) — держим список на случай, если что-то
// всё же проскочит мимо переопределений в runtimeEnvOverrides.
func systemLeftoverCandidates() []string {
	var out []string
	home, _ := os.UserHomeDir()
	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" && home != "" {
		cacheHome = filepath.Join(home, ".cache")
	}
	if cacheHome != "" {
		out = append(out, filepath.Join(cacheHome, "zen"), filepath.Join(cacheHome, "mozilla"))
	}
	if home != "" {
		out = append(out, filepath.Join(home, ".mozilla"))
	}
	return out
}

// runtimeExtraDirs — дополнительные каталоги (кроме profile/temp), которые
// нужно создать в Data/ до запуска Zen.
func runtimeExtraDirs(dataDir string) []string {
	return []string{filepath.Join(dataDir, "cache")}
}

// runtimeEnvOverrides — переменные окружения, которые направляют Zen писать
// кэш и временные файлы внутрь portable-папки вместо $HOME/.cache и /tmp.
func runtimeEnvOverrides(dataDir string) []string {
	return []string{
		"TMPDIR=" + filepath.Join(dataDir, "temp"),
		"XDG_CACHE_HOME=" + filepath.Join(dataDir, "cache"),
		// Крэш-репортер по умолчанию пишет вне portable-папки — выключаем.
		"MOZ_CRASHREPORTER_DISABLE=1",
	}
}
