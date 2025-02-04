package cmd

import (
	"bytes"
	"testing"
)

func Test_execute(t *testing.T) {
	actual := new(bytes.Buffer)

	RootCmd.SetOut(actual)
	RootCmd.SetErr(actual)
	RootCmd.SetArgs([]string{"kubeconfig", "--trace", "--noproxy", "pandora"})
	RootCmd.Execute()
}
