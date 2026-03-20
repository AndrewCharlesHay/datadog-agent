// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2025-present Datadog, Inc.

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

	var openshiftInstance *computepb.Instance
	for instance := range instanceIterator {
		fmt.Fprintf(&out, "instance: %+v\n", instance)

		if !strings.Contains(instance.GetName(), "fakeint") {
			openshiftInstance = instance
		}
	}

	vmName := openshiftInstance.GetName()
	networks := openshiftInstance.GetNetworkInterfaces()
	if len(networks) == 0 {
		fmt.Fprintf(&out, "the VM %s has 0 interfaces... can't reach for kube config file", vmName)
		return
	}

	if networks[0] == nil {
		fmt.Fprintf(&out, "the VM %s has 1 interface but it's nil :-( can't reach for kube config file", vmName)
		return
	}

	vmIP := networks[0].GetNetworkIP()

	fmt.Fprintf(&out, "Found gcp vm running openshift: %s - IP: %s\n", vmName, vmIP)

	return
}
