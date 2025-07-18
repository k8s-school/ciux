package util

import (
	"os"

	"github.com/spf13/cobra"
)

// See https://github.com/kubernetes/kubectl/blob/b35935138f5330d299e89b1278b5738487e4f015/pkg/cmd/top/top_pod.go#L175
func AddLabelSelectorFlagVar(cmd *cobra.Command, p *string) {
	cmd.Flags().StringVarP(p, "selector", "l", *p, "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2). Matching objects must satisfy all of the specified label constraints.")
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true // file exists
	}
	if os.IsNotExist(err) {
		return false // file does not exist
	}
	// some other error occurred (e.g. permission)
	return false
}
