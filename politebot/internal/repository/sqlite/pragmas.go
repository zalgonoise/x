package sqlite

func ReadOnlyPragmas() map[string]string {
	return map[string]string{
		"journal_mode": "off",
		"busy_timeout": "5000",
		"synchronous":  "off",
		"cache_size":   "1000000000",
		"foreign_keys": "true",
		"temp_store":   "memory",
		"optimize":     "",
	}
}

func ReadWritePragmas() map[string]string {
	return map[string]string{
		"journal_mode": "wal",
		"busy_timeout": "5000",
		"synchronous":  "normal",
		"cache_size":   "1000000000",
		"mmap_size":    "30000000000",
		"foreign_keys": "true",
		"temp_store":   "memory",
	}
}
