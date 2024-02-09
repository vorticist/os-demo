package main

import (
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func main() {
	app := cdk8s.NewApp(nil)
	networkChart := NewNetworkSetupChart(app, "network-setup", nil)
	metallbConfigChart := NewMetallbConfigChart(app, "metallb-config", nil)
	metallbConfigChart.AddDependency(networkChart)
	aiChart := NewAIChart(app, "ai-iac", nil)
	aiChart.AddDependency(metallbConfigChart)
	app.Synth()
}
