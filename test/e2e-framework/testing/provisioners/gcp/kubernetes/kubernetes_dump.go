package gcpkubernetes

import (
	"context"
	"fmt"
	"strings"

	gcpapi "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
)

func DumpOpenshiftClusterState(ctx context.Context, name string) (ret string, err error) {
	var out strings.Builder
	defer func() {
		ret = out.String()
	}()

	fmt.Fprintf(&out, "stack name: '%s'\n", name)

	gcpClient, err := gcpapi.NewInstancesRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to init GCP instance client: %w", err)
	}

	instanceFilter := "labels.managed-by=pulumi AND labels.stack=" + name
	instanceIterator := gcpClient.List(ctx, &computepb.ListInstancesRequest{
		Filter: &instanceFilter,
	}).All()

	for instance, _ := range instanceIterator {
		fmt.Fprintf(&out, "instance: %+v\n", instance)
	}

	return
}
