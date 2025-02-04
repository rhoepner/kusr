package cmd

import (
	"os"
	"testing"
)

func Test_executeInit(t *testing.T) {

	RootCmd.SetOut(os.Stdout)
	RootCmd.SetErr(os.Stdout)
	RootCmd.SetArgs([]string{"init"})
	RootCmd.Execute()
}
