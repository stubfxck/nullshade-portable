# Zen Browser Portable

*[Читать по-русски](README-PORTABLE_RU.md)* · 💬 Discord: [discord.gg/eCQYpRx8Wv](https://discord.gg/eCQYpRx8Wv)

A portable build of Zen Browser. Everything lives inside this folder — profile,
settings, extensions, history. Carry it on a USB flash drive or an external SSD.

## Launching

```text
ZenBrowserPortable.exe             — run this
Support\Start-ZenPortable.bat      — fallback, if the .exe is blocked by policy
```

(`Support\` is an internal folder, normally hidden; it holds the fallback .bat
and the version files used for auto-update. No need to touch it.)

## Where your data lives

```text
Data/profile   — profile (settings, extensions, history)
Data/cache     — cache
Data/temp      — temp files
App/Zen        — browser binaries
```

## Updates

`ZenBrowserPortable.exe` checks for new versions on every launch and updates
`App\Zen\` (and itself) automatically — no manual zip downloads needed.
Behavior is configurable in `Data\launcher-config.json` (created on first launch):

- `autoUpdateEnabled: false` — turn off update checking entirely.
- `updateMode: "block"` — install the update before Zen launches.
- `updateMode: "background"` (default) — launch immediately, the update
  downloads in the background and applies on the next launch.

## Important

- Close the browser before removing the USB drive.
- Don't run two instances from the same folder at once.

---

## Maximum locality (zero-trace mode)

This build is designed to write NOTHING outside its own folder:

| Protection | How it's done |
|---|---|
| Profile, cache, extensions, DRM modules | `-profile Data\profile` — everything stays inside |
| Temp files | TEMP/TMP → `Data\temp` |
| AppData environment variables | APPDATA/LOCALAPPDATA → `Data\appdata` |
| Crash dumps in %APPDATA%\zen | disabled (`MOZ_CRASHREPORTER_DISABLE=1`) |
| Telemetry pings | disabled (prefs + policies.json) |
| Auto-update (writes to C:\ProgramData) | disabled (policies.json) |
| "Default browser" registry entry | disabled (`DontCheckDefaultBrowser`) |
| Notification registration in the registry (AppUserModelID) | notifications are drawn by the browser itself (`alerts.useSystemBackend=false`) |
| Taskbar jump list | never created (`browser.taskbar.lists.*=false`) |
| Stray folders accidentally created in AppData | removed by `ZenBrowserPortable.exe` after the browser closes (only the ones that didn't exist before launch) |

IMPORTANT:

1. Always launch via `ZenBrowserPortable.exe`. Cleanup on exit only happens in the .exe launcher (the .bat is a fallback, with no cleanup).
2. Never run `App\Zen\zen.exe` directly — it will create an empty profile in your system AppData.

### Honest limits (unavoidable without a sandbox/VM)

Traces created by Windows itself when launching any .exe (not the browser — the OS):

- Prefetch (`C:\Windows\Prefetch\ZEN.EXE-*.pf`) — evidence the program ran;
- SRUM / event logs / icon cache — system-level statistics;
- the page file (may contain fragments of the process's memory).

This is a limitation of any portable software, no exceptions (including
PortableApps.com). If you need ABSOLUTE zero traces on the machine, run this
same folder inside a sandbox: Windows Sandbox (built into Win10/11 Pro) or
Sandboxie-Plus (free) — those also intercept the OS's own writes.
