# Zen Browser Portable

*от [Nullshade Studio](https://github.com/stubfxck)*

*[Read this in English](README.md)*

**Неофициальная portable-сборка [Zen Browser](https://zen-browser.app/) для Windows и Linux.**
На Windows — без установки, без прав администратора, без следов в реестре и AppData;
на Linux — вообще ничего не пишет за пределы своей папки. Все данные
(профиль, расширения, история, кэш) живут в одной папке. Можно носить на
флешке или внешнем SSD.

Приватные вкладки без нового окна и другие моды — отдельно, в
[nullshade-private-tab](https://github.com/stubfxck/nullshade-private-tab).

Сборки обновляются автоматически: GitHub Actions каждый понедельник берёт
свежий официальный релиз Zen, переупаковывает его в portable, прогоняет
проверки и публикует в Releases. Битая сборка проверки не пройдёт и опубликована не будет.

💬 Discord: **[discord.gg/eCQYpRx8Wv](https://discord.gg/eCQYpRx8Wv)**

---

## Скачать

▶ **[Последняя версия — в Releases](../../releases/latest)** — в каждом релизе сразу обе платформы:
`ZenBrowserPortable-<версия>-win-x86_64.zip` и `ZenBrowserPortable-<версия>-linux-x86_64.tar.gz`
(плюс варианты `-arm64`/`-aarch64` для обеих).

### Быстрый старт — Windows

1. Скачайте zip из [Releases](../../releases/latest) и распакуйте в любую папку (диск, флешка, внешний SSD).
2. Запустите **`ZenBrowserPortable.exe`**.
3. Всё. Браузер работает, все данные остаются в папке `Data\`.

> Если SmartScreen ругается на неподписанный .exe — «Подробнее → Выполнить в любом случае»,
> либо используйте запасной `Start-ZenPortable.bat` — результат тот же.

Подробная инструкция для пользователей лежит внутри каждого архива: `README-PORTABLE.md`.

### Быстрый старт — Linux (Debian/Ubuntu/Mint и подобные)

1. Скачайте tar.gz из [Releases](../../releases/latest) и распакуйте: `tar -xzf ZenBrowserPortable-*-linux-x86_64.tar.gz`.
2. Запустите **`./ZenBrowserPortable`** из распакованной папки.
3. Всё, то же самое, что на Windows — данные остаются внутри `Data/`.

> Если в архиве почему-то нет бинарника лаунчера (редкий случай — только если
> сборка прошла без доступного Go), используйте `Support/start-zen-portable.sh` —
> делает то же самое, только без самообновления.

### Что внутри архива

```text
ZenBrowserPortable/
├─ ZenBrowserPortable(.exe)   ← запускать это
├─ README-PORTABLE.md         ← инструкция
├─ App/Zen/                   ← браузер (официальные бинарники Zen, не трогать)
├─ Data/                      ← ВСЕ ваши данные (профиль, кэш, temp, launcher-config.json)
└─ Support/                   ← служебное (на Windows — скрытая папка, на Linux — обычная)
   ├─ Start-ZenPortable.bat   ← запасной запуск на Windows
   ├─ start-zen-portable.sh   ← запасной запуск на Linux
   ├─ VERSION.txt             ← версия и дата сборки (для людей)
   └─ version.json            ← версия для автообновления (для лаунчера)
```

На Windows на виду только `.exe`, инструкция и папки `App`/`Data` — всё
служебное убрано в скрытую `Support\` (лаунчер сам следит, чтобы она
оставалась скрытой, даже если архиватор при распаковке этот атрибут не
перенёс). На Linux `Support/` не прячется — это не то, что пользователи этой
платформы ожидают от обычной папки, поэтому она просто видна.

### Обновление

`ZenBrowserPortable.exe` сам проверяет новые версии при каждом запуске и
обновляет `App\Zen\` (и себя) — вручную скачивать zip не нужно.

Поведение настраивается в `Data\launcher-config.json` (создаётся при первом
запуске, переживает обновления — лежит в `Data\`, как и весь ваш профиль):

```json
{
  "autoUpdateEnabled": true,
  "updateMode": "background",
  "privateTabModEnabled": false
}
```

- `autoUpdateEnabled: false` — полностью выключить проверку обновлений.
- `updateMode: "block"` — скачать и поставить обновление ДО запуска Zen (дольше, зато сразу на последней версии).
- `updateMode: "background"` — запустить Zen сразу, обновление скачается в фоне и встанет автоматически при следующем запуске (по умолчанию).
- `privateTabModEnabled: true` — установить и держать актуальным [мод приватных вкладок](https://github.com/stubfxck/nullshade-private-tab) (по умолчанию выключено).

Обновляется только `App\Zen\` — ваш профиль в `Data\` при этом не трогается.

(Штатный апдейтер самого Firefox/Zen внутри браузера по-прежнему отключён
намеренно: он пишет в системные папки и может сломать portable-режим.
Обновляет всё именно лаунчер снаружи.)

---

## Режим «ноль следов»

Сборка спроектирована так, чтобы ничего не писать за пределы своей папки:

- профиль, кэш, DRM-модули, временные файлы — внутри `Data\`;
- телеметрия, крэш-репорты, автообновление — отключены;
- запись в реестр (браузер по умолчанию, уведомления, jump list) — отключена;
- лаунчер после закрытия браузера убирает случайно созданные папки в AppData
  (не трогая данные установленного Zen/Firefox).

Полная таблица защит и честные пределы (Prefetch и прочие следы уровня самой Windows) —
в `README-PORTABLE.md` внутри архива.

Два правила:

1. Запускайте только через лаунчер `ZenBrowserPortable`/`ZenBrowserPortable.exe` (уборка следов есть только в нём).
2. Не запускайте `App/Zen/zen`/`App\Zen\zen.exe` напрямую — создаст пустой профиль в системном доме/AppData.

---

## Как это работает (для тех, кто хочет собрать сам)

Этот репозиторий — не форк браузера, а **билдер**. Бинарники берутся из
официальных релизов [zen-browser/desktop](https://github.com/zen-browser/desktop/releases)
без каких-либо модификаций кода — добавляются только лаунчер и portable-настройки.

| Файл | Назначение |
|---|---|
| `.github/workflows/build-portable.yml` | автосборка обеих платформ: каждый понедельник + вручную через Run workflow (джобы `portable-windows` и `portable-linux`) |
| `builder/package-release.ps1` | Windows: скачивает релиз → распаковывает → собирает portable → проверяет → zip |
| `builder/package-release.sh` | Linux: та же задача, на bash — скачивает официальный tar.xz → собирает portable → проверяет (включая то, что исполняемый бит реально пережил упаковку) → tar.gz |
| `builder/build-local.ps1` | альтернатива: полная сборка из исходников под Windows (локально, долго) |
| `launcher/*.go` | исходник лаунчера, разнесён по build-тегам там, где платформы реально отличаются (`platform_windows.go`/`platform_linux.go`, `console_windows.go`/`console_linux.go`, `archive_windows.go`/`archive_linux.go` — zip vs tar.gz при самообновлении) — всё остальное (`main.go`, `update.go`, `version.go`, `mod.go`) общее |
| `builder/template/` | файлы, которые кладутся в каждый пакет (Start-ZenPortable.bat, start-zen-portable.sh, README-PORTABLE.md) |

### Собрать локально

Windows (нужен только 7-Zip; Go — опционально для .exe-лаунчера):

```powershell
powershell -ExecutionPolicy Bypass -File .\builder\package-release.ps1                  # последняя версия
powershell -ExecutionPolicy Bypass -File .\builder\package-release.ps1 -Version 1.21.9b # конкретная
powershell -ExecutionPolicy Bypass -File .\builder\package-release.ps1 -Arch arm64      # Windows ARM
```

Linux (нужны `curl`, `jq`, `tar`; Go — опционально для бинарника лаунчера):

```bash
./builder/package-release.sh                    # последняя версия, x86_64
./builder/package-release.sh --version 1.21.9b
./builder/package-release.sh --arch aarch64      # Linux ARM
```

Результат: `output/ZenBrowserPortable-<версия>-win-<арх>.zip` или
`output/ZenBrowserPortable-<версия>-linux-<арх>.tar.gz`.

### Полная сборка из исходников (продвинутый вариант)

Требования по [официальной документации Zen](https://docs.zen-browser.app/contribute/desktop/building):
Git, Node.js, MozillaBuild, 7-Zip, Visual Studio (Desktop development with C++), 40+ ГБ, несколько часов.

```powershell
powershell -ExecutionPolicy Bypass -File .\builder\build-local.ps1 -Ref 1.21.9b
```

---

## FAQ

**Это официальный проект Zen?**
Нет. Это независимая обёртка. Сам браузер — официальные, неизменённые бинарники
из релизов zen-browser/desktop. Весь код обёртки открыт в этом репозитории,
каждая сборка воспроизводима через публичный лог в Actions.

**Почему браузер (Zen) сам не предлагает обновиться?**
Так задумано — штатный апдейтер Firefox/Zen отключён, потому что пишет в системные
папки. Вместо этого обновляет `ZenBrowserPortable.exe` снаружи, при каждом запуске
(см. раздел «Обновление» выше).

**Работает с флешки?**
Да. Для скорости и корректной работы песочницы браузера рекомендуется NTFS (не FAT32).

**Пропали данные после запуска!**
Скорее всего, был запущен `App\Zen\zen.exe`/`App/Zen/zen` напрямую — он создал пустой
профиль в системе. Данные целы: закройте браузер и запустите лаунчер
(`ZenBrowserPortable.exe`/`ZenBrowserPortable`).

**Как проверить, что portable-режим работает?**
Откройте `about:profiles` — активный профиль должен указывать на `...\Data\profile`.

---

## Лицензия и благодарности

- Zen Browser — [zen-browser/desktop](https://github.com/zen-browser/desktop) (MPL-2.0).
- Скрипты и лаунчер этого репозитория можно свободно использовать и форкать.
