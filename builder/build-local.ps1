<#
.SYNOPSIS
  Zen Browser Portable — режим сборки из исходников (Source).
  Клонирует zen-browser/desktop, собирает браузер и упаковывает portable.

  ТРЕБОВАНИЯ (см. https://docs.zen-browser.app/contribute/desktop/building):
  - Git, Node.js (npm)
  - MozillaBuild (в PATH)
  - 7-Zip (в PATH)
  - Visual Studio + workload "Desktop development with C++"
  - ~40+ ГБ свободного места, сборка занимает часы

.EXAMPLE
  .\build-local.ps1                    # ветка dev (актуальный код)
  .\build-local.ps1 -Ref 1.21.9b      # конкретный релизный тег
#>
param(
    [string]$Repo = "https://github.com/zen-browser/desktop.git",
    [string]$Ref = "dev",
    [string]$WorkDir = "work",
    [string]$OutputDir = "output"
)

$ErrorActionPreference = "Stop"

$Root = (Get-Location).Path
$SourceDir = Join-Path $Root "$WorkDir\source"
$Out = Join-Path $Root $OutputDir
New-Item -ItemType Directory -Force $Out | Out-Null

# --- 1. Клонировать / обновить исходники --------------------------------------
if (-not (Test-Path $SourceDir)) {
    Write-Host "== Cloning $Repo ($Ref)..."
    git clone $Repo $SourceDir --depth 10
} else {
    Write-Host "== Updating source..."
    Push-Location $SourceDir
    git fetch --depth 10 origin $Ref
    Pop-Location
}
Push-Location $SourceDir
git checkout $Ref

# --- 2. Сборка по официальному flow -------------------------------------------
Write-Host "== npm install..."
npm install

Write-Host "== npm run init (bootstrap Firefox, долго)..."
npm run init

Write-Host "== npm run build (сборка, очень долго)..."
npm run build

# --- 3. Найти папку с zen.exe и нужным runtime --------------------------------
Write-Host "== Searching for build output..."
$Candidates = Get-ChildItem -Recurse -Filter "zen.exe" -ErrorAction SilentlyContinue |
    Where-Object {
        (Test-Path (Join-Path $_.Directory.FullName "omni.ja")) -or
        (Test-Path (Join-Path $_.Directory.FullName "browser"))
    }
if (-not $Candidates) {
    Write-Host "zen.exe not found. Try clean rebuild:"
    Write-Host "  npm run reset-ff; npm run init; npm run build"
    throw "Build output not found"
}
$BinDir = ($Candidates | Select-Object -First 1).Directory.FullName
Write-Host "   Found: $BinDir"

$Commit = git rev-parse --short HEAD
Pop-Location

# --- 4. Упаковать portable (общая логика с release-режимом) --------------------
$PkgName = "ZenBrowserPortable-src-$Ref-$Commit-win64"
$Pkg = Join-Path $Root "$WorkDir\$PkgName"
if (Test-Path $Pkg) { Remove-Item -Recurse -Force $Pkg }

New-Item -ItemType Directory -Force `
    (Join-Path $Pkg "App\Zen"), `
    (Join-Path $Pkg "Data\profile"), `
    (Join-Path $Pkg "Data\cache"), `
    (Join-Path $Pkg "Data\temp") | Out-Null

Copy-Item -Recurse -Force "$BinDir\*" (Join-Path $Pkg "App\Zen")
Copy-Item (Join-Path $PSScriptRoot "template\Start-ZenPortable.bat") $Pkg
Copy-Item (Join-Path $PSScriptRoot "template\README-PORTABLE.md") $Pkg

@"
Zen Browser Portable
Mode: built-from-source
Ref: $Ref
Commit: $Commit
Built: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
Source: $Repo
"@ | Set-Content -Encoding UTF8 (Join-Path $Pkg "VERSION.txt")

$ZipPath = Join-Path $Out "$PkgName.zip"
if (Test-Path $ZipPath) { Remove-Item -Force $ZipPath }
Compress-Archive -Path "$Pkg\*" -DestinationPath $ZipPath

Write-Host ""
Write-Host "DONE:"
Write-Host "  Folder: $Pkg"
Write-Host "  Zip:    $ZipPath"
