package main

import (
	"github.com/arkusnexus/ai-demo/iac/imports/k8s"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

type AIChartProps struct {
	cdk8s.ChartProps
}

func NewAIChart(scope constructs.Construct, id string, props *AIChartProps) cdk8s.Chart {
	namespace := "ai-iac"
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
	//registrySecretName := "registry-secrets"
	//rawAuth := fmt.Sprintf("%v:%v", os.Getenv("DOCKER_USER"), os.Getenv("DOCKER_PASSWORD"))
	//auth := base64.StdEncoding.EncodeToString([]byte(rawAuth))
	//rawDockercfgjson := fmt.Sprintf(`
	//{
	//  "auths": {
	//	"%v": {
	//	  "auth": "%v",
	//	  "email": "%v"
	//	}
	//  }
	//}
	//`, os.Getenv("DOCKER_REGISTRY_SERVER"), auth, os.Getenv("DOCKER_EMAIL"))
	//dockercfgjson := base64.StdEncoding.EncodeToString([]byte(rawDockercfgjson))
	//k8s.NewKubeSecret(chart, &registrySecretName, &k8s.KubeSecretProps{
	//	Metadata: &k8s.ObjectMeta{
	//		Name:      &registrySecretName,
	//		Namespace: &namespace,
	//	},
	//	Type: jsii.String("kubernetes.io/dockerconfigjson"),
	//	Data: &map[string]*string{
	//		".dockerconfigjson": &dockercfgjson,
	//	},
	//})
	/************************** registry auth ********************************/
	/************************** label-studio  ********************************/
	//cdk8s.NewHelm(chart, jsii.String("label-studio"), &cdk8s.HelmProps{
	//	Chart:     jsii.String("heartex/label-studio"),
	//	HelmFlags: &[]*string{jsii.String("--namespace"), jsii.String(namespace)},
	//	Values: &map[string]interface{}{
	//		"app": map[string]interface{}{
	//			"service": map[string]interface{}{
	//				"type": "LoadBalancer",
	//			},
	//		},
	//		"replica": map[string]interface{}{
	//			"replicaCount": 1,
	//		},
	//	},
	//})
	/************************** label-studio  ********************************/
	/************************** jupyterhub    ********************************/
	cdk8s.NewHelm(chart, jsii.String("jupyter-hub"), &cdk8s.HelmProps{
		Chart:     jsii.String("jupyterhub/jupyterhub"),
		HelmFlags: &[]*string{jsii.String("--namespace"), jsii.String(namespace)},
		Version:   jsii.String("3.2.1"),
		Values: &map[string]interface{}{
			"singleuser": map[string]interface{}{
				"profileList": []map[string]interface{}{
					{
						"display_name": "GPU Server",
						"description":  "Spawns a notebook server with access to a GPU",
						"kubespawner_override": map[string]interface{}{
							"extra_resource_limits": map[string]interface{}{
								"nvidia.com/gpu": "1",
							},
						},
					},
				},
			},
		},
	})
	/************************** jupyterhub    ********************************/
	/************************** arkusnexus    ********************************/
	//labels := map[string]*string{
	//	"app": jsii.String("arkusnexus-demo-be"),
	//}
	//
	//appName := "arkusnexus-demo-be"
	//deploymentName := fmt.Sprintf("%v-deployment", appName)
	//k8s.NewKubeDeployment(chart, jsii.String(deploymentName), &k8s.KubeDeploymentProps{
	//	Metadata: &k8s.ObjectMeta{
	//		Name:      jsii.String(deploymentName),
	//		Namespace: jsii.String(namespace),
	//	},
	//	Spec: &k8s.DeploymentSpec{
	//		Replicas: jsii.Number(1),
	//		Selector: &k8s.LabelSelector{
	//			MatchLabels: &labels,
	//		},
	//		Template: &k8s.PodTemplateSpec{
	//			Metadata: &k8s.ObjectMeta{
	//				Labels: &labels,
	//			},
	//			Spec: &k8s.PodSpec{
	//				Containers: &[]*k8s.Container{{
	//					Name:  jsii.String("be-container"),
	//					Image: jsii.String("aiarkusnexus/opensource-demo-be:latest"),
	//					Ports: &[]*k8s.ContainerPort{{
	//						ContainerPort: jsii.Number(8080),
	//					}},
	//				}},
	//			},
	//		},
	//	},
	//})
	//
	//serviceName := fmt.Sprintf("%v-service", appName)
	//targetPort := float64(8080)
	//port := float64(80)
	//k8s.NewKubeService(chart, jsii.String(serviceName), &k8s.KubeServiceProps{
	//	Metadata: &k8s.ObjectMeta{},
	//	Spec: &k8s.ServiceSpec{
	//		Ports: &[]*k8s.ServicePort{
	//			{
	//				Port:       &port,
	//				TargetPort: k8s.IntOrString_FromNumber(&targetPort),
	//			},
	//		},
	//		Selector: &labels,
	//		Type:     jsii.String("LoadBalancer"),
	//	},
	//})
	/************************** arkusnexus    ********************************/
	return chart
}
