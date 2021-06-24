package utils

import (
	"testing"
)

func TestInitConfig(t *testing.T) {
	err := InitConfig("../../config/config.yaml")

	if err != nil {
		t.Errorf(err.Error())
	}

	if Settings.LeakGlobals.Version == 0.0 {
		t.Errorf("Could not parse config file")
	}

	return
}
