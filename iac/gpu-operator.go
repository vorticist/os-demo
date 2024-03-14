package main

import (
	"github.com/arkusnexus/ai-demo/iac/imports/k8s"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

type GpuOperatorChartProps struct {
	cdk8s.ChartProps
}

func NewGpuOperatorChart(scope constructs.Construct, id string, props *GpuOperatorChartProps) cdk8s.Chart {
	namespace := "gpu-operator"
	var cprops cdk8s.ChartProps
	if props != nil {
		cprops = props.ChartProps
	}
	cprops.Namespace = jsii.String(namespace)
	chart := cdk8s.NewChart(scope, jsii.String(id), &cprops)

	k8s.NewKubeNamespace(chart, &namespace, &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: &namespace,
			Labels: &map[string]*string{
				"pod-security.kubernetes.io/enforce": jsii.String("privileged"),
			},
		},
	})

	cdk8s.NewHelm(chart, jsii.String("gpu-op"), &cdk8s.HelmProps{
		Chart: jsii.String("nvidia/gpu-operator"),
		HelmFlags: &[]*string{
			jsii.String("--wait"),
			jsii.String("--create-namespace"),
			jsii.String("--namespace"), jsii.String(namespace),
		},
		Version: jsii.String("23.3.2"),
		Values:  &map[string]interface{}{
			// "driver": map[string]interface{}{
			// 	"enabled": true,
			// },
			// "mig": map[string]interface{}{
			// 	"strategy": "mixed",
			// },
			// "migManager": map[string]interface{}{
			// 	"config": map[string]interface{}{
			// 		"name": "mig-parted-config",
			// 	},
			// },
		},
	})

	return chart
}
