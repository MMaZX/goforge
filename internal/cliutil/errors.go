package cliutil

// ConfigError wraps errors from loading configuration or connecting to the
// database, so main can distinguish "your setup is wrong" (exit 2) from
// "the migration run failed" (exit 1).
type ConfigError struct {
	Err error
}

func (e *ConfigError) Error() string { return e.Err.Error() }
func (e *ConfigError) Unwrap() error { return e.Err }
