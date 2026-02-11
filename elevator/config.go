package elevator

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// TODO: Decide to only use config.go OR con_load.go. Learn how to use bufio

// LoadConfig reads a config file using the same "--key value" lines.
// Keys and enum values are treated case-insensitively.
func LoadConfig(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m := make(map[string]string)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "--") {
			rest := strings.TrimPrefix(line, "--")
			// split into key and the rest as value (value may contain spaces)
			parts := strings.Fields(rest)
			if len(parts) >= 1 {
				key := strings.ToLower(parts[0])
				val := ""
				if len(parts) >= 2 {
					// value is everything after the first space
					idx := strings.Index(rest, " ")
					if idx >= 0 {
						val = strings.TrimSpace(rest[idx+1:])
					}
				}
				m[key] = val
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return m, nil
}

// GetString assigns the string value for key (case-insensitive) to dest.
func GetString(cfg map[string]string, key string, dest *string) bool {
	if v, ok := cfg[strings.ToLower(key)]; ok {
		*dest = v
		return true
	}
	return false
}

// GetInt parses an integer value for key and stores it in dest.
func GetInt(cfg map[string]string, key string, dest *int) bool {
	if v, ok := cfg[strings.ToLower(key)]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			*dest = i
			return true
		}
	}
	return false
}

// GetEnum looks up the enum string for key and maps it via mapping (case-insensitive).
// Example mapping: map[string]int{"en1": 0, "en2": 1}
func GetEnum(cfg map[string]string, key string, mapping map[string]int, dest *int) bool {
	if v, ok := cfg[strings.ToLower(key)]; ok {
		if val, ok2 := mapping[strings.ToLower(v)]; ok2 {
			*dest = val
			return true
		}
	}
	return false
}
