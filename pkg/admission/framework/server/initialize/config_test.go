package initialize

import (
	"testing"
)

func TestConfigValidate(t *testing.T) {
	configFilePath := "../../../../../etc/handler.yaml"
	c, e := ReadHandlerConfigFromFile(configFilePath)
	if e != nil {
		t.Fatalf("ReadHandlerConfigFromFile %s failed, %v", configFilePath, e)
	}
	e = c.Validate()
	if e != nil {
		t.Fatalf("Validate failed, %v", e)
	}
	t.Log(c.String())
}
