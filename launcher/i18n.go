// i18n.go — минимальная локализация консоли лаунчера. Если основной язык
// системы (Windows UI language) русский — выводим текст на русском, иначе
// на английском. Никаких файлов ресурсов/каталогов сообщений — на таком
// объёме текста хватает пары строк RU/EN прямо по месту вызова.
package main

var isRussian = detectRussian()

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

// t возвращает ru, если система на русском, иначе en. Обёртка нужна только
// ради читаемости на месте вызова (t("...", "...") короче, чем if/else).
func t(ru, en string) string {
	if isRussian {
		return ru
	}
	return en
}
