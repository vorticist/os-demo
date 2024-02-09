package main

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

type MetallbConfigChartProps struct {
	cdk8s.ChartProps
}

func NewMetallbConfigChart(scope constructs.Construct, id string, props *NetworkSetupChartProps) cdk8s.Chart {
	var cprops cdk8s.ChartProps
	if props != nil {
		cprops = props.ChartProps
	}
	chart := cdk8s.NewChart(scope, jsii.String(id), &cprops)

	cdk8s.NewInclude(chart, jsii.String("metallb-config"), &cdk8s.IncludeProps{
		Url: jsii.String("cluster/network/metallb-config.yaml"),
	})

	return chart
}
