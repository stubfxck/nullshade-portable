//go:build windows

package main

// detectRussian спрашивает Windows про язык интерфейса пользователя
// (GetUserDefaultUILanguage) и сравнивает основной language ID с русским
// (0x19) — не обращая внимания на региональный вариант (ru-RU, ru-BY и т.д.,
// младший байт LANGID).
func detectRussian() bool {
	proc := kernel32.NewProc("GetUserDefaultUILanguage")
	ret, _, _ := proc.Call()
	langID := uint16(ret)
	const primaryLangMask = 0x3ff
	const langRussian = 0x19
	return (langID & primaryLangMask) == langRussian
}
