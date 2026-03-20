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
	"k8s.io/client-go/tools/clientcmd"

	"github.com/DataDog/datadog-agent/test/e2e-framework/testing/utils/infra"
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

	sshClient, err := infra.SshConnectToInstance(vmIP, "22", "gce")
	if err != nil {
		fmt.Fprintf(&out, "failed to ssh to the VM %s can't reach for kube config file", vmName)
		return
	}
	defer sshClient.Close()

	sshOutput, err := infra.SshRunCommand(
		sshClient,
		"cat .kube/config",
	)
	if err != nil {
		return "", err
	}

	kubeConfig, err := clientcmd.Load(sshOutput)
	if err != nil {
		return "", err
	}

	fmt.Fprintf(&out, "Found kubeconfiguration:\n\n%s\n\n", string(sshOutput))

	var _ = kubeConfig

	return
}
