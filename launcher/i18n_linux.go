//go:build linux

package main

import (
	"os"
	"strings"
)

// detectRussian смотрит на переменные окружения, которыми Linux сообщает
// программам язык интерфейса — в том порядке, в котором их же учитывает
// gettext: LANGUAGE (список через ':', GNU-расширение) важнее LC_ALL,
// который важнее LC_MESSAGES, который важнее LANG. Достаточно, чтобы
// ПЕРВЫЙ непустой из них начинался с "ru" (ru_RU.UTF-8, ru_BY, ru — все
// варианты) или заглавного региона, региональный вариант не важен.
func detectRussian() bool {
	for _, name := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"} {
		v := os.Getenv(name)
		if v == "" {
			continue
		}
		// LANGUAGE может быть списком через ':' — берём первый элемент.
		first := strings.SplitN(v, ":", 2)[0]
		return strings.HasPrefix(strings.ToLower(first), "ru")
	}
	return false
}
