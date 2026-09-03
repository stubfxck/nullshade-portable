@echo off
setlocal

rem Лежит в Support\, корень пакета - на уровень выше.
set "ROOT=%~dp0..\"
set "APP=%ROOT%App\Zen"
set "PROFILE=%ROOT%Data\profile"
set "CACHE=%ROOT%Data\cache"
set "TEMP_DIR=%ROOT%Data\temp"

if not exist "%PROFILE%" mkdir "%PROFILE%"
if not exist "%CACHE%" mkdir "%CACHE%"
if not exist "%TEMP_DIR%" mkdir "%TEMP_DIR%"

set "APPDATA_DIR=%ROOT%Data\appdata\Roaming"
set "LOCALAPPDATA_DIR=%ROOT%Data\appdata\Local"
if not exist "%APPDATA_DIR%" mkdir "%APPDATA_DIR%"
if not exist "%LOCALAPPDATA_DIR%" mkdir "%LOCALAPPDATA_DIR%"

set "TEMP=%TEMP_DIR%"
set "TMP=%TEMP_DIR%"
rem Все, что резолвится через переменные окружения - внутрь portable-папки
set "APPDATA=%APPDATA_DIR%"
set "LOCALAPPDATA=%LOCALAPPDATA_DIR%"
rem Крэш-репортер пишет в %%APPDATA%%\zen вне portable-папки - отключаем
set "MOZ_CRASHREPORTER_DISABLE=1"

start "" "%APP%\zen.exe" -profile "%PROFILE%" -no-remote

endlocal
