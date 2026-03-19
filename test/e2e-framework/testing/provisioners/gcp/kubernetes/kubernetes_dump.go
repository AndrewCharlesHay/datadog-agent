package gcpkubernetes

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func DumpOpenshiftClusterState(ctx context.Context, name string) (ret string, err error) {
	var out strings.Builder
	defer func() {
		ret = out.String()
	}()

	fmt.Fprintf(&out, "stack name: '%s'\n", name)

	args := []string{
		"--project", "'datadog-agent-sandbox'",
		"compute",
		"instances",
		"list",
		"--filter='labels.team=agent-contint AND labels.managed-by=pulumi AND labels.username=alexandre-lavigne'",
	}

	cmd := exec.Command("gcloud", args...)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(&out, "failed to get instances list: %s\n", err.Error())
	} else {
		cmdOut, cmdErr := cmd.CombinedOutput()
		fmt.Fprintf(&out, "output:\n%s\nerrors:\n%s", string(cmdOut), cmdErr.Error())
	}

	

	return
}
