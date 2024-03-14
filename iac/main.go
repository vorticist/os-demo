package main

import (
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func main() {
	app := cdk8s.NewApp(nil)
	networkChart := NewNetworkSetupChart(app, "network-setup", nil)
	metallbConfigChart := NewMetallbConfigChart(app, "metallb-config", nil)
	metallbConfigChart.AddDependency(networkChart)

	gpuOperatorChart := NewGpuOperatorChart(app, "gpu-op-chart", nil)
	gpuOperatorChart.AddDependency(metallbConfigChart)

	aiChart := NewAIChart(app, "ai-iac", nil)
	aiChart.AddDependency(gpuOperatorChart)

	app.Synth()
}
