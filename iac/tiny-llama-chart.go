package main

import (
	"encoding/base64"
	"fmt"
	"github.com/arkusnexus/ai-demo/iac/imports/k8s"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	"os"
)

type TinyLlamaChartProps struct {
	cdk8s.ChartProps
}

func NewTinyLlamaChart(scope constructs.Construct, id string, props *TinyLlamaChartProps) cdk8s.Chart {
	namespace := "tiny-llama"
	var cprops cdk8s.ChartProps
	if props != nil {
		cprops = props.ChartProps
	}
	cprops.Namespace = jsii.String(namespace)
	chart := cdk8s.NewChart(scope, jsii.String(id), &cprops)

	k8s.NewKubeNamespace(chart, &namespace, &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{Name: &namespace},
	})

	/************************** registry auth ********************************/
	registrySecretName := "registry-secrets"
	rawAuth := fmt.Sprintf("%v:%v", os.Getenv("DOCKER_USER"), os.Getenv("DOCKER_PASSWORD"))
	auth := base64.StdEncoding.EncodeToString([]byte(rawAuth))
	rawDockercfgjson := fmt.Sprintf(`
	{
	  "auths": {
		"%v": {
		  "auth": "%v",
		  "email": "%v"
		}
	  }
	}
	`, os.Getenv("DOCKER_REGISTRY_SERVER"), auth, os.Getenv("DOCKER_EMAIL"))
	dockercfgjson := base64.StdEncoding.EncodeToString([]byte(rawDockercfgjson))
	k8s.NewKubeSecret(chart, &registrySecretName, &k8s.KubeSecretProps{
		Metadata: &k8s.ObjectMeta{
			Name:      &registrySecretName,
			Namespace: &namespace,
		},
		Type: jsii.String("kubernetes.io/dockerconfigjson"),
		Data: &map[string]*string{
			".dockerconfigjson": &dockercfgjson,
		},
	})
	/************************** registry auth ********************************/

	labels := map[string]*string{
		"app": jsii.String("arkusnexus-demo-be"),
	}

	appName := "tiny-llama"
	deploymentName := fmt.Sprintf("%v-deployment", appName)
	k8s.NewKubeDeployment(chart, jsii.String(deploymentName), &k8s.KubeDeploymentProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String(deploymentName),
			Namespace: jsii.String(namespace),
		},
		Spec: &k8s.DeploymentSpec{
			Replicas: jsii.Number(1),
			Selector: &k8s.LabelSelector{
				MatchLabels: &labels,
			},
			Template: &k8s.PodTemplateSpec{
				Metadata: &k8s.ObjectMeta{
					Labels: &labels,
				},
				Spec: &k8s.PodSpec{
					Containers: &[]*k8s.Container{{
						Name:  jsii.String("text-generation-webui"),
						Image: jsii.String("aiarkusnexus/tinyllama:latest"),
						Ports: &[]*k8s.ContainerPort{{
							ContainerPort: jsii.Number(7860),
						}},
					}},
				},
			},
		},
	})

	serviceName := fmt.Sprintf("%v-service", appName)
	port := float64(7860)
	k8s.NewKubeService(chart, jsii.String(serviceName), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{},
		Spec: &k8s.ServiceSpec{
			Ports: &[]*k8s.ServicePort{
				{
					Port: &port,
				},
			},
			Selector: &labels,
			Type:     jsii.String("LoadBalancer"),
		},
	})

	return chart
}
