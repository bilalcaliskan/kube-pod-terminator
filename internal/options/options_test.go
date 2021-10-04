package options

import "testing"

func TestGetKubePodTerminatorOptions(t *testing.T) {
	t.Log("fetching default options.KubePodTerminatorOptions")
	opts := GetKubePodTerminatorOptions()
	t.Logf("fetched default options.KubePodTerminatorOptions, %v\n", opts)
}
