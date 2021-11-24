package options

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetKubePodTerminatorOptions(t *testing.T) {
	t.Log("fetching default options.KubePodTerminatorOptions")
	opts := GetKubePodTerminatorOptions()
	assert.NotNil(t, opts)
	t.Logf("fetched default options.KubePodTerminatorOptions, %v\n", opts)
}
