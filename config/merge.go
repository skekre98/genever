package config

func mergeMaps(dst, src map[string]any) {
	for k, v := range src {
		if mv, ok := v.(map[string]any); ok {
			if existing, ok := dst[k].(map[string]any); ok {
				mergeMaps(existing, mv)
				continue
			}
		}
		dst[k] = v
	}
}
