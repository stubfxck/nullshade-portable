<#
.SYNOPSIS
  Zen Browser Portable — быстрый режим (Release).
  Скачивает официальный релиз Zen Browser, распаковывает установщик
  и собирает portable-папку + zip. Ничего не устанавливает в систему.

.EXAMPLE
  .\package-release.ps1                 # последняя версия, x86_64
  .\package-release.ps1 -Version 1.21.9b
  .\package-release.ps1 -Arch arm64
#>
param(
    [string]$Version = "latest",
    [ValidateSet("x86_64", "arm64")]
    [string]$Arch = "x86_64",
    [string]$OutputDir = "output",
    [string]$WorkDir = "work"
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

$Root = (Get-Location).Path
$Work = Join-Path $Root $WorkDir
$Out  = Join-Path $Root $OutputDir
New-Item -ItemType Directory -Force $Work, $Out | Out-Null

# --- 1. Определить версию через GitHub API -----------------------------------
Write-Host "== Resolving Zen Browser version..."
if ($Version -eq "latest") {
    $release = Invoke-RestMethod "https://api.github.com/repos/zen-browser/desktop/releases/latest"
} else {
    $release = Invoke-RestMethod "https://api.github.com/repos/zen-browser/desktop/releases/tags/$Version"
}
$Tag = $release.tag_name
Write-Host "   Version: $Tag"

# --- 2. Скачать установщик ----------------------------------------------------
$AssetName = if ($Arch -eq "arm64") { "zen.installer-arm64.exe" } else { "zen.installer.exe" }
$Asset = $release.assets | Where-Object { $_.name -eq $AssetName }
if (-not $Asset) { throw "Asset $AssetName not found in release $Tag" }

$InstallerPath = Join-Path $Work $AssetName
Write-Host "== Downloading $AssetName ($([math]::Round($Asset.size / 1MB)) MB)..."
Invoke-WebRequest $Asset.browser_download_url -OutFile $InstallerPath

# --- 3. Распаковать NSIS-установщик через 7-Zip -------------------------------
$SevenZip = @(
    "7z",
    "$env:ProgramFiles\7-Zip\7z.exe",
    "${env:ProgramFiles(x86)}\7-Zip\7z.exe"
) | Where-Object { Get-Command $_ -ErrorAction SilentlyContinue } | Select-Object -First 1
if (-not $SevenZip) { throw "7-Zip not found. Install it: winget install 7zip.7zip" }

$ExtractDir = Join-Path $Work "extracted"
if (Test-Path $ExtractDir) { Remove-Item -Recurse -Force $ExtractDir }
Write-Host "== Extracting installer..."
& $SevenZip x $InstallerPath "-o$ExtractDir" -y | Out-Null

# В Firefox-style NSIS установщиках приложение лежит в папке core
$CoreDir = Join-Path $ExtractDir "core"
if (-not (Test-Path (Join-Path $CoreDir "zen.exe"))) {
    $ZenExe = Get-ChildItem $ExtractDir -Recurse -Filter "zen.exe" | Select-Object -First 1
    if (-not $ZenExe) { throw "zen.exe not found inside installer" }
    $CoreDir = $ZenExe.Directory.FullName
}
Write-Host "   App files: $CoreDir"

# --- 4. Собрать portable-структуру --------------------------------------------
$PkgName = "ZenBrowserPortable-$Tag-win-$Arch"
$Pkg = Join-Path $Work $PkgName
if (Test-Path $Pkg) { Remove-Item -Recurse -Force $Pkg }

Write-Host "== Building portable package..."
New-Item -ItemType Directory -Force `
    (Join-Path $Pkg "App\Zen"), `
    (Join-Path $Pkg "Data\profile"), `
    (Join-Path $Pkg "Data\cache"), `
    (Join-Path $Pkg "Data\temp"), `
    (Join-Path $Pkg "Support") | Out-Null

Copy-Item -Recurse -Force "$CoreDir\*" (Join-Path $Pkg "App\Zen")

# Portable-переопределения: не обновляться, не трогать систему, не слать отчёты.
# Причины каждой строки описаны в README (раздел «Почему portable может сломаться»).
@"
// Portable build overrides
pref("app.update.service.enabled", false);
pref("app.update.background.scheduling.enabled", false);
pref("app.update.auto", false);
// Не предлагать сделать браузером по умолчанию (иначе в реестр Windows
// запишется путь к portable-папке, который сломается при смене буквы диска)
pref("browser.shell.checkDefaultBrowser", false);
// Телеметрия и «эксперименты» — пишут пинги в %APPDATA%\zen вне portable-папки
pref("toolkit.telemetry.enabled", false);
pref("toolkit.telemetry.unified", false);
pref("datareporting.healthreport.uploadEnabled", false);
pref("datareporting.policy.dataSubmissionEnabled", false);
pref("app.normandy.enabled", false);
// Крэш-репорты не отправлять автоматически (папка Crash Reports живёт вне portable)
pref("browser.crashReports.unsubmittedCheck.autoSubmit2", false);
// Маркер portable-режима (используется сообществом portable-сборок Zen)
pref("zen.portable.mode", true);
// --- «Ноль следов» в системе ---
// Системные toast-уведомления Windows требуют регистрации AppUserModelID
// в реестре HKCU — рисуем уведомления средствами самого браузера
pref("alerts.useSystemBackend", false);
// Jump list («недавнее/частое» у иконки на панели задач) — пишется браузером
// в файлы CustomDestinations профиля Windows — не создаём вообще
pref("browser.taskbar.lists.enabled", false);
pref("browser.taskbar.lists.frequent.enabled", false);
pref("browser.taskbar.lists.recent.enabled", false);
pref("browser.taskbar.lists.tasks.enabled", false);
"@ | Set-Content -Encoding ASCII (Join-Path $Pkg "App\Zen\defaults\pref\portable.js") -ErrorAction SilentlyContinue

# policies.json — более жёсткий уровень, чем prefs: пользователь не сможет
# случайно включить апдейтер из настроек (обновление portable = замена App\Zen)
$DistDir = Join-Path $Pkg "App\Zen\distribution"
New-Item -ItemType Directory -Force $DistDir | Out-Null
@"
{
  "policies": {
    "DisableAppUpdate": true,
    "DisableTelemetry": true,
    "DontCheckDefaultBrowser": true,
    "DisableFirefoxStudies": true
  }
}
"@ | Set-Content -Encoding ASCII (Join-Path $DistDir "policies.json")

$TemplateDir = Join-Path $PSScriptRoot "template"
$SupportDir = Join-Path $Pkg "Support"
# Всё, что не нужно видеть при обычном использовании (запасной .bat,
# машиночитаемая версия), лежит в Support\ и помечается скрытой папкой —
# на виду только .exe, README и App/Data. См. hideSupportDir() в launcher/main.go:
# он же перепрячет папку на диске пользователя, если та расшилась при копировании.
Copy-Item (Join-Path $TemplateDir "Start-ZenPortable.bat") $SupportDir
Copy-Item (Join-Path $TemplateDir "README-PORTABLE.md") $Pkg
Copy-Item (Join-Path $TemplateDir "README-PORTABLE_RU.md") $Pkg

@"
Zen Browser Portable
Version: $Tag
Arch: win-$Arch
Mode: release-repackage
Built: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
Source: https://github.com/zen-browser/desktop/releases/tag/$Tag
"@ | Set-Content -Encoding UTF8 (Join-Path $SupportDir "VERSION.txt")

# version.json — машиночитаемая версия для автообновления в лаунчере.
# portableTag должен буквально совпадать с tag_name релиза в этом репозитории
# (см. .github/workflows/build-portable.yml: tag_name: portable-$Tag).
$BuiltAtIso = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
@{
    portableTag = "portable-$Tag"
    zenTag      = $Tag
    arch        = "win-$Arch"
    builtAt     = $BuiltAtIso
} | ConvertTo-Json | Set-Content -Encoding UTF8 (Join-Path $SupportDir "version.json")

# Атрибут Hidden на Support\ здесь НЕ ставим: это транзитная папка сборки
# ($Pkg живёт только в work\, не попадает пользователю), а wildcard-раскрытие
# "$Pkg\*" в Compress-Archive молча пропускает скрытые элементы — если
# пометить Support\ скрытым до зипа, он целиком вылетает из архива (поймано
# на реальном прогоне CI). Реальное скрытие для пользователя делает
# hideSupportDir() в launcher/main.go на каждом запуске — этого достаточно,
# ставить атрибут ещё и здесь незачем.

# --- 5. Лаунчер .exe (если доступен Go; иначе остаётся .bat) ----------------
$LauncherSrc = Join-Path $PSScriptRoot "..\launcher"
if ((Get-Command go -ErrorAction SilentlyContinue) -and (Test-Path (Join-Path $LauncherSrc "main.go"))) {
    Write-Host "== Building ZenBrowserPortable.exe launcher..."
    Push-Location $LauncherSrc
    $env:GOOS = "windows"
    $env:GOARCH = if ($Arch -eq "arm64") { "arm64" } else { "amd64" }
    $env:CGO_ENABLED = "0"
    # Без -H=windowsgui: лаунчеру нужна консоль, чтобы показывать прогресс
    # проверки/установки обновлений (см. launcher/console.go). Окно прячется
    # само (hideConsoleWindow) в момент запуска Zen — see main.go.
    go build -ldflags "-s -w" -o (Join-Path $Pkg "ZenBrowserPortable.exe") .
    Pop-Location
    if (-not (Test-Path (Join-Path $Pkg "ZenBrowserPortable.exe"))) {
        throw "Launcher build failed"
    }
} else {
    Write-Host "== Go not found: skipping .exe launcher, .bat will be used"
}

# --- 6. Проверки пакета (smoke tests) — защита от битых сборок -----------------
Write-Host "== Validating package..."
$Required = @(
    "App\Zen\zen.exe",
    "App\Zen\omni.ja",
    "Support\Start-ZenPortable.bat",
    "Support\VERSION.txt",
    "Support\version.json",
    "README-PORTABLE.md",
    "README-PORTABLE_RU.md"
)
foreach ($r in $Required) {
    if (-not (Test-Path (Join-Path $Pkg $r))) { throw "VALIDATION FAILED: missing $r" }
}
$ZenSize = (Get-Item (Join-Path $Pkg "App\Zen\zen.exe")).Length
if ($ZenSize -lt 300KB) { throw "VALIDATION FAILED: zen.exe suspiciously small ($ZenSize bytes)" }
$AppSize = (Get-ChildItem (Join-Path $Pkg "App\Zen") -Recurse | Measure-Object -Property Length -Sum).Sum
if ($AppSize -lt 100MB) { throw "VALIDATION FAILED: App/Zen total size too small ($([math]::Round($AppSize/1MB)) MB)" }
Write-Host "   OK: zen.exe present, App/Zen = $([math]::Round($AppSize/1MB)) MB"

# --- 7. Zip --------------------------------------------------------------------
$ZipPath = Join-Path $Out "$PkgName.zip"
if (Test-Path $ZipPath) { Remove-Item -Force $ZipPath }
Write-Host "== Creating $ZipPath ..."
Compress-Archive -Path "$Pkg\*" -DestinationPath $ZipPath

Write-Host ""
Write-Host "DONE:"
Write-Host "  Folder: $Pkg"
Write-Host "  Zip:    $ZipPath"
