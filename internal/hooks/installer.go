package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const supervisorMarker = "claude-code-supervisor"

// Claude Code hook format:
// {"PreToolUse": [{"matcher": "...", "hooks": [{"type": "command", "command": "...", "timeout": 1000}]}]}
// matcher is a tool name, pipe-separated list, or "" to match all.

type hookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout"`
}

type hookMatcherEntry struct {
	Matcher string        `json:"matcher"`
	Hooks   []hookCommand `json:"hooks"`
	Marker  string        `json:"marker,omitempty"`
}

func settingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}

func readSettings(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]any), nil
		}
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", path, err)
	}
	return settings, nil
}

func writeSettings(path string, settings map[string]any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", dir, err)
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal settings: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func buildHookCommand(hookType, apiURL string) string {
	// Claude Code passes hook data as JSON on stdin (not env vars).
	// We read stdin with jq, add our hook type, and POST to the Supervisor.
	endpoint := apiURL + "/api/events"

	switch hookType {
	case "PreToolUse":
		return fmt.Sprintf(
			`cat | jq -c '. + {hook: "pre_tool_use", session: .session_id, tool: .tool_name, input: .tool_input, cwd: .cwd}' | curl -s -m 1 -X POST %s -H 'Content-Type: application/json' -d @- 2>/dev/null; true`,
			endpoint,
		)
	case "PostToolUse":
		return fmt.Sprintf(
			`cat | jq -c '. + {hook: "post_tool_use", session: .session_id, tool: .tool_name, input: .tool_input, output: .tool_output}' | curl -s -m 1 -X POST %s -H 'Content-Type: application/json' -d @- 2>/dev/null; true`,
			endpoint,
		)
	case "Notification":
		return fmt.Sprintf(
			`cat | jq -c '. + {hook: "notification", session: .session_id, title: .notification_title, body: .notification_body}' | curl -s -m 1 -X POST %s -H 'Content-Type: application/json' -d @- 2>/dev/null; true`,
			endpoint,
		)
	default:
		return ""
	}
}

func Install(apiPort int) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}

	settings, err := readSettings(path)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("http://localhost:%d", apiPort)

	hooksMap, ok := settings["hooks"].(map[string]any)
	if !ok {
		hooksMap = make(map[string]any)
	}

	hookTypes := []string{"PreToolUse", "PostToolUse", "Notification"}

	for _, ht := range hookTypes {
		entry := hookMatcherEntry{
			Matcher: "", // match all tools
			Hooks: []hookCommand{{
				Type:    "command",
				Command: buildHookCommand(ht, apiURL),
				Timeout: 1000,
			}},
			Marker: supervisorMarker,
		}

		existing, _ := hooksMap[ht].([]any)

		// Remove any existing supervisor matcher entries (idempotent)
		var filtered []any
		for _, e := range existing {
			if m, ok := e.(map[string]any); ok {
				if m["marker"] == supervisorMarker {
					continue
				}
			}
			filtered = append(filtered, e)
		}

		// Convert entry to map[string]any for JSON
		entryBytes, _ := json.Marshal(entry)
		var entryMap map[string]any
		json.Unmarshal(entryBytes, &entryMap)

		filtered = append(filtered, entryMap)
		hooksMap[ht] = filtered
	}

	settings["hooks"] = hooksMap
	return writeSettings(path, settings)
}

func Uninstall() error {
	path, err := settingsPath()
	if err != nil {
		return err
	}

	settings, err := readSettings(path)
	if err != nil {
		return err
	}

	hooksMap, ok := settings["hooks"].(map[string]any)
	if !ok {
		return nil // No hooks to remove
	}

	uninstallTypes := []string{"PreToolUse", "PostToolUse", "Notification"}

	for _, ht := range uninstallTypes {
		existing, _ := hooksMap[ht].([]any)

		var filtered []any
		for _, e := range existing {
			if m, ok := e.(map[string]any); ok {
				if m["marker"] == supervisorMarker {
					continue
				}
			}
			filtered = append(filtered, e)
		}

		if len(filtered) == 0 {
			delete(hooksMap, ht)
		} else {
			hooksMap[ht] = filtered
		}
	}

	if len(hooksMap) == 0 {
		delete(settings, "hooks")
	} else {
		settings["hooks"] = hooksMap
	}

	return writeSettings(path, settings)
}
