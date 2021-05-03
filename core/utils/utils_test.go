package utils

import (
	"testing"
)

func TestInitConfig(t *testing.T) {
	InitConfig("../../config/config.yaml")

	if Settings.LeakGlobals.Version == 0.0 {
		t.Errorf("Could not parse config file")
	}
	return
}
