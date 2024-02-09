package main

import (
	"github.com/arkusnexus/ai-demo/iac/imports/k8s"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

type NetworkSetupChartProps struct {
	cdk8s.ChartProps
}

func NewNetworkSetupChart(scope constructs.Construct, id string, props *NetworkSetupChartProps) cdk8s.Chart {
	namespace := jsii.String("metallb-system")
	var cprops cdk8s.ChartProps
	if props != nil {
		cprops = props.ChartProps
	}
	cprops.Namespace = namespace
	chart := cdk8s.NewChart(scope, jsii.String(id), &cprops)

	/************************** metallb       ********************************/
	k8s.NewKubeNamespace(chart, namespace, &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{Name: namespace},
	})

	cdk8s.NewHelm(chart, jsii.String("metallb-helm"), &cdk8s.HelmProps{
		Chart:     jsii.String("metallb/metallb"),
		HelmFlags: &[]*string{jsii.String("--namespace"), namespace},
		Values:    &map[string]interface{}{},
	})
	/************************** metallb       ********************************/

	return chart
}
