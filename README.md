# Zen Browser Portable

*by [Nullshade Studio](https://github.com/stubfxck)*

*[Читать по-русски](README_RU.md)*

**Unofficial portable build of [Zen Browser](https://zen-browser.app/) for Windows and Linux.**
No installer, no admin rights, no registry/AppData traces on Windows or
writes outside its own folder on Linux — every bit of data (profile,
extensions, history, cache) lives in one folder. Carry it on a USB flash
drive or an external SSD.

Private tabs without a new window, and other mods, live separately in
[nullshade-private-tab](https://github.com/stubfxck/nullshade-private-tab).

Builds update themselves automatically: every Monday, GitHub Actions grabs
the latest official Zen release, repackages it as portable, runs validation
checks, and publishes to Releases. A broken build fails validation and never
gets published.

💬 Discord: **[discord.gg/eCQYpRx8Wv](https://discord.gg/eCQYpRx8Wv)**

---

## Download

▶ **[Latest release](../../releases/latest)** — every release carries both platforms:
`ZenBrowserPortable-<version>-win-x86_64.zip` and `ZenBrowserPortable-<version>-linux-x86_64.tar.gz`
(plus `-arm64`/`-aarch64` variants for both).

### Quick start — Windows

1. Download the zip from [Releases](../../releases/latest) and extract it anywhere (drive, USB stick, external SSD).
2. Run **`ZenBrowserPortable.exe`**.
3. That's it. The browser runs, and all your data stays inside the `Data\` folder.

> If SmartScreen complains about the unsigned .exe — "More info → Run anyway",
> or use the fallback `Start-ZenPortable.bat` instead — same result.

A detailed guide for end users ships inside every archive: `README-PORTABLE.md`.

### Quick start — Linux (Debian/Ubuntu/Mint and similar)

1. Download the tar.gz from [Releases](../../releases/latest) and extract it: `tar -xzf ZenBrowserPortable-*-linux-x86_64.tar.gz`.
2. Run **`./ZenBrowserPortable`** from inside the extracted folder.
3. That's it — same deal as Windows, all data stays inside `Data/`.

> No launcher binary in the archive for some reason (rare — only happens if a
> build ran without Go available)? Use `Support/start-zen-portable.sh` instead;
> it does the same job minus self-updating.

### What's inside the archive

```text
ZenBrowserPortable/
├─ ZenBrowserPortable(.exe)   ← run this
├─ README-PORTABLE.md         ← end-user guide
├─ App/Zen/                   ← the browser (official Zen binaries, don't touch)
├─ Data/                      ← ALL your data (profile, cache, temp, launcher-config.json)
└─ Support/                   ← internal (hidden on Windows, a plain folder on Linux)
   ├─ Start-ZenPortable.bat   ← Windows fallback launcher
   ├─ start-zen-portable.sh   ← Linux fallback launcher
   ├─ VERSION.txt             ← version and build date (human-readable)
   └─ version.json            ← version marker for auto-update (machine-readable)
```

On Windows, only the `.exe`, the guide, and `App`/`Data` are visible by
default — everything internal is tucked into the hidden `Support\` folder
(the launcher re-hides it on every run, in case the archiver that extracted
the zip didn't preserve the attribute). Linux doesn't hide `Support/` — that's
not a convention this platform's users expect, so it's just left visible.

### Updates

`ZenBrowserPortable.exe` checks for new versions on every launch and updates
`App\Zen\` (and itself) automatically — no manual zip re-downloads needed.

Behavior is configurable in `Data\launcher-config.json` (created on first
launch, survives updates since it lives in `Data\` along with the rest of
your profile):

```json
{
  "autoUpdateEnabled": true,
  "updateMode": "background",
  "privateTabModEnabled": false
}
```

- `autoUpdateEnabled: false` — turn off update checking entirely.
- `updateMode: "block"` — download and install the update before Zen launches (slower, but always on the latest version).
- `updateMode: "background"` — launch Zen immediately, download in the background, apply automatically on the next launch (default).
- `privateTabModEnabled: true` — install and keep up to date the [private-tab mod](https://github.com/stubfxck/nullshade-private-tab) (opt-in, off by default).

Only `App\Zen\` gets updated — your profile in `Data\` is never touched.

(Firefox/Zen's own built-in updater is still intentionally disabled: it writes
to system folders and can break portable mode. The launcher handles all
updates from the outside instead.)

---

## Zero-trace mode

The build is designed to write nothing outside its own folder:

- profile, cache, DRM modules, temp files — all inside `Data\`;
- telemetry, crash reports, auto-update — disabled;
- registry writes (default browser, notifications, jump list) — disabled;
- the launcher cleans up any stray folders the browser created in AppData
  after it closes (without touching data from an installed Zen/Firefox).

The full table of protections and honest limitations (Prefetch and other
OS-level traces) is in `README-PORTABLE.md` inside the archive.

Two rules:

1. Always launch via the `ZenBrowserPortable`/`ZenBrowserPortable.exe` launcher (trace cleanup only happens there).
2. Never run `App/Zen/zen`/`App\Zen\zen.exe` directly — it will create an empty profile in your system home/AppData.

---

## How it works (for anyone who wants to build it themselves)

This repository is not a browser fork — it's a **builder**. Binaries come
straight from the official [zen-browser/desktop](https://github.com/zen-browser/desktop/releases)
releases with no code modifications — only the launcher and portable settings
are added on top.

| File | Purpose |
|---|---|
| `.github/workflows/build-portable.yml` | auto-build for both platforms: every Monday + manually via Run workflow (`portable-windows` and `portable-linux` jobs) |
| `builder/package-release.ps1` | Windows: downloads the release → extracts → assembles portable → validates → zips |
| `builder/package-release.sh` | Linux: same job, bash — downloads the official tar.xz → assembles portable → validates (incl. checking the executable bit survived) → tar.gz |
| `builder/build-local.ps1` | alternative: full Windows build from source (local, slow) |
| `launcher/*.go` | launcher source, split by build tag where the two platforms genuinely differ (`platform_windows.go`/`platform_linux.go`, `console_windows.go`/`console_linux.go`, `archive_windows.go`/`archive_linux.go` for zip vs tar.gz update payloads) — everything else (`main.go`, `update.go`, `version.go`, `mod.go`) is shared |
| `builder/template/` | files bundled into every package (Start-ZenPortable.bat, start-zen-portable.sh, README-PORTABLE.md) |

### Build locally

Windows (needs only 7-Zip; Go is optional, for the .exe launcher):

```powershell
powershell -ExecutionPolicy Bypass -File .\builder\package-release.ps1                  # latest version
powershell -ExecutionPolicy Bypass -File .\builder\package-release.ps1 -Version 1.21.9b # specific version
powershell -ExecutionPolicy Bypass -File .\builder\package-release.ps1 -Arch arm64      # Windows ARM
```

Linux (needs `curl`, `jq`, `tar`; Go is optional, for the launcher binary):

```bash
./builder/package-release.sh                          # latest version, x86_64
./builder/package-release.sh --version 1.21.9b
./builder/package-release.sh --arch aarch64            # Linux ARM
```

Output: `output/ZenBrowserPortable-<version>-win-<arch>.zip` or
`output/ZenBrowserPortable-<version>-linux-<arch>.tar.gz`.

### Full build from source (advanced)

Requirements per the [official Zen documentation](https://docs.zen-browser.app/contribute/desktop/building):
Git, Node.js, MozillaBuild, 7-Zip, Visual Studio (Desktop development with C++), 40+ GB free space, a few hours.

```powershell
powershell -ExecutionPolicy Bypass -File .\builder\build-local.ps1 -Ref 1.21.9b
```

---

## FAQ

**Is this an official Zen project?**
No. This is an independent wrapper. The browser itself is the official,
unmodified binaries from zen-browser/desktop releases. All of the wrapper's
code is open in this repository, and every build is reproducible via the
public Actions log.

**Why doesn't Zen offer to update itself?**
By design — Firefox/Zen's own updater is disabled because it writes to
system folders. `ZenBrowserPortable.exe` handles updates from the outside
instead, on every launch (see "Updates" above).

**Does it work from a USB drive?**
Yes. On Windows, NTFS is recommended over FAT32 for speed and correct browser
sandboxing. On Linux, make sure the filesystem preserves Unix permissions
(ext4 works fine; avoid mounting from exFAT/FAT32 without permission options).

**My data disappeared after launching!**
`App\Zen\zen.exe`/`App/Zen/zen` was probably run directly — it created an
empty profile in the system. Your data is safe: close the browser and launch
the launcher (`ZenBrowserPortable.exe`/`ZenBrowserPortable`) instead.

**How do I verify portable mode is actually working?**
Open `about:profiles` — the active profile should point at `...\Data\profile`.

---

## License and credits

- Zen Browser — [zen-browser/desktop](https://github.com/zen-browser/desktop) (MPL-2.0).
- The scripts and launcher in this repository are free to use and fork.
