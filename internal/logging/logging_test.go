package logging

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetLogger(t *testing.T) {
	t.Log("getting logger")
	logger := GetLogger()
	assert.NotNil(t, logger)
	t.Log("will try logger for debugging")
	logger.Info("this is a test log by *zap.Logger!")
}
