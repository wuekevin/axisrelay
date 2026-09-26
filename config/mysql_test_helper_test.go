package config

import "testing"

func setTestMySQLConfig(t *testing.T) {
	t.Helper()
	t.Setenv("AXISRELAY_DATABASE_DRIVER", "mysql")
	t.Setenv("AXISRELAY_DATABASE_HOST", "127.0.0.1")
	t.Setenv("AXISRELAY_DATABASE_PORT", "3306")
	t.Setenv("AXISRELAY_DATABASE_USER", "root")
	t.Setenv("AXISRELAY_DATABASE_PASSWORD", "")
	t.Setenv("AXISRELAY_DATABASE_NAME", "axisrelay_config_test")
}
