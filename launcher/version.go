package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// localVersion — то, что package-release.ps1 кладёт в version.json при сборке.
// PortableTag сравнивается напрямую с tag_name релиза на GitHub (формат "portable-1.21.16b").
type localVersion struct {
	PortableTag string `json:"portableTag"`
	ZenTag      string `json:"zenTag"`
	Arch        string `json:"arch"`
	BuiltAt     string `json:"builtAt"`
}

func readVersionJSON(root string) (*localVersion, error) {
	b, err := os.ReadFile(filepath.Join(root, "Support", "version.json"))
	if err != nil {
		// Совместимость со старыми установками (до переезда в Support\):
		// version.json лежал прямо в корне пакета. Само подчистится при
		// следующем успешном обновлении (см. cleanupLegacyVersionJSON).
		b, err = os.ReadFile(filepath.Join(root, "version.json"))
		if err != nil {
			return nil, err
		}
	}
	var v localVersion
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// cleanupLegacyVersionJSON убирает version.json из корня пакета, если он там
// остался от установки, сделанной до переезда версии в Support\.
func cleanupLegacyVersionJSON(root string) {
	legacy := filepath.Join(root, "version.json")
	if _, err := os.Stat(legacy); err == nil {
		_ = os.Remove(legacy)
	}
}

// launcherConfig живёт в Data\launcher-config.json — то есть переживает
// обновления (Data\ никогда не перезаписывается), в отличие от файлов
// в корне пакета.
type launcherConfig struct {
	Comment              string `json:"_comment"`
	AutoUpdateEnabled    bool   `json:"autoUpdateEnabled"`
	UpdateMode           string `json:"updateMode"` // "block" или "background"
	PrivateTabModEnabled bool   `json:"privateTabModEnabled"`
}

func defaultConfig() launcherConfig {
	return launcherConfig{
		Comment: "autoUpdateEnabled: true/false — включить/выключить проверку обновлений. " +
			"updateMode: \"block\" (скачать и поставить обновление перед запуском Zen) " +
			"или \"background\" (запустить Zen сразу, обновление скачается в фоне и встанет при следующем запуске). " +
			"privateTabModEnabled: true/false — установить и держать актуальным мод приватных вкладок " +
			"(github.com/stubfxck/nullshade-private-tab), выключено по умолчанию.",
		AutoUpdateEnabled:    true,
		UpdateMode:           "background",
		PrivateTabModEnabled: false,
	}
}

func loadOrCreateConfig(dataDir string) launcherConfig {
	path := filepath.Join(dataDir, "launcher-config.json")
	if b, err := os.ReadFile(path); err == nil {
		var cfg launcherConfig
		if json.Unmarshal(b, &cfg) == nil && (cfg.UpdateMode == "block" || cfg.UpdateMode == "background") {
			return cfg
		}
	}
	cfg := defaultConfig()
	saveConfig(dataDir, cfg)
	return cfg
}

func saveConfig(dataDir string, cfg launcherConfig) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(dataDir, "launcher-config.json"), b, 0o644)
}
