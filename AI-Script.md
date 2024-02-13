# Intro

- Who are we?

With recent development of AI technologies, new possibilites are available for developers to create new products leveraging AI. However, most of the tools that are available are designed to be used as a service over the cloud, and even if you can install some of these services in your private cloud infrastructure, for some cases it may represent a risk on data privacy, 

- Here we explain how tools are used in AI workflows to fine tune existing models.
    - Dataset preprocessing
    - Iterate training

Additionally, setting up a workflow environment to train or serve an AI model is not excactly as straightforward as setting it up in a 'as a service' environment, that's why, for this tutorial, we put together a sample workflow environment to train and serve an image recognition model using open source tools and technologies that can be installed in your own private cloud or even in your own hardware on premises. The project leverages K8s to be able to eaisily setup and configure the tools needed to train the model but also showcases how a custom app can be quickly integrated into the environment to serve the model itself. A basic understanding of k8s components is recommended in order to follow along this demo.

We're presenting a couple of examples with pretrained models that then get fine tuned, but much of the principles shown here can be used to setup an environment to train from scratch if that's what your project requires.

# Sample Use Case 1

For our first use case we will be demonstrating how to setup an environment for what has become the hello world of AI models, train and serve an image recognition model, using YOLOv8. We will fine tune the model in order for it to be able to identify different kinds of tacos in images.

The way a workflow looks for this use case is as follows:
- We need a dataset to fine tune our model, usually videos or images that will be tagger in order to use in the training and validation processes. The images should be as close as the real case the model will be dealing with and should contain the object to identify, in our case that's pictures of tacos.
- We'll need a tool to tag our dataset efectively identifying every taco that appears in the dataset. For this we will use LabelStudio in our environment.
- The dataset needs to be divided into thain, test and validation sub sets // TODO: mention the ideal proportion of each sub set compared to the total dataset size
- The training process itself is pretty straigthforward, the important bit being to pay attention to the accuracy against the validation set until we reach the desired accuracy. We will be using a Jupyter notebook and a python script for this.
- Once the model has been fine tuned, we can also include a custom app in our environment to serve the model. Ideally, serving the model, should be done in a separate environment than the training, but for the purposes of this project we bundled everything in a single environment.

## Prerequisites
- A k8s cluster where the environment will be installed. You can have this in the cloud or in prem. For development you can also use minikube or [kind](https://kind.sigs.k8s.io/) as we are doing for this demo.
- `kubectl` [installed and configured](https://kubernetes.io/docs/reference/kubectl/) to connect to the k8s cluster.
- Helm 3 installed and add all the repos needed for the tools we'll be using (`helm repo add <NAME> <URL>`)
    - MetalLB: `metallb https://metallb.github.io/metallb` 
    - LabelStudio: `heartex https://charts.heartex.com/`
    - JupyterHub: `jupyterhub https://hub.jupyter.org/helm-chart/`
- [Docker](https://www.docker.com/get-started/)
- [nodejs](https://nodejs.org/en) & [cdk8s](https://cdk8s.io/docs/latest/cli/installation/) cli
- If running on custom hardware outside a cloud provider, the cluster most likely will need to have installed a custom load balancer. K3s comes with it's own load balancer preinstalled, but for this demo we'll be using [MetalLB](https://metallb.universe.tf/installation/#installation-with-helm).
- Presentation setup:
  - `kind create cluster --config cluster/kind-config.yaml`
  - `kind load docker-image aiarkusnexus/opensource-demo-be:latest`
  - `kubectl apply -f dist/0000-network-setup.k8s.yaml`
  - `kubectl apply -f dist/0001-metallb-config.k8s.yaml`
## Steps
### CDK8s IaC project setup
The first thing we'll want to do is to [create a new CDK8S project](https://cdk8s.io/docs/latest/cli/init/). CDK8S is inspired by AWS' CDK, but was designed for managing infrastructure inside a k8s cluster using code. It is not tied to AWS so it can be used with any other cloud provider or custom hardware as long as there is a k8s cluster accessible with kubectl. 

The recommended template for cdk8s is the typescript one, but for this demo we choose to work with go:
``` bash
mkdir my-demo-folder
cd my-demo-folder
mkdir iac
cd iac
cdk8s init go-app
```
The main.go file is the entry point for the project and it will contain generated code by the cdk8s cli, you can modify that code in place, but we'll be using a separate file to keep things tidy. 

It is also important to run the [cdk8s import command](https://cdk8s.io/docs/latest/cli/import/) as this will import all the base constructs to work with kubernetes.
``` bash
cdk8s import
```

Delete every thing from the main.go file until it looks like this:
``` go
package main

import (
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func main() {
	app := cdk8s.NewApp(nil)
	app.Synth()
}

```
### Create a chart for our AI infrastructure
`cdk8s` uses the concept of charts to bundle resource management, which are different from helm charts, but act pretty much in the same way. You define a set of resources under a cdk8s chart and will generate a single resources file to apply to the cluster. Helcharts can be part of a cdk8s chart.

Create a new file called `ai-iac.go` and we'll add the first of our tools, LabelStudio:
``` go
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

    /************************** label-studio  ********************************/
	cdk8s.NewHelm(chart, jsii.String("label-studio"), &cdk8s.HelmProps{
		Chart:     jsii.String("heartex/label-studio"),
		HelmFlags: &[]*string{jsii.String("--namespace"), jsii.String(namespace)},
		Values: &map[string]interface{}{
			"app": map[string]interface{}{
				"service": map[string]interface{}{
					"type": "LoadBalancer",
				},
			},
			"replica": map[string]interface{}{
				"replicaCount": 1,
			},
		},
	})
	/************************** label-studio  ********************************/
}
```
Make sure to add the helm repo as stated in the prerequisites: `helm repo add heartex https://charts.heartex.com/` 

The first thing we do is to create a namespace and tie it to our cdk8s chart so that every construct defined within it will be created under that same namespace.

``` go
	cprops.Namespace = jsii.String(namespace)
	chart := cdk8s.NewChart(scope, jsii.String(id), &cprops)

	k8s.NewKubeNamespace(chart, &namespace, &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{Name: &namespace},
	})
```

`cdk8s.NewHelm` is the costruct provided to work with helm charts, and the helm chart to use is passed in the Chart property of the `HelmProps` parameter `Chart:     jsii.String("heartex/label-studio")` 

The values passed in the rest of the props, correspond to the configuration provided by the helm chart [installation instructions for Label Studio](https://labelstud.io/guide/install_k8s)

And that's it, this will make sure that when we apply our manifest to the cluster it will be boostraped with the LabelStudio app available to start using it, but before testing that we also need to modify our `main.go` file again.

``` go
...

func main() {
	app := cdk8s.NewApp(nil)

	networkChart := NewNetworkSetupChart(app, "network-setup", nil)
	metallbConfigChart := NewMetallbConfigChart(app, "metallb-config", nil)
	metallbConfigChart.AddDependency(networkChart)

	aiChart := NewAIChart(app, "ai-iac", nil)
	aiChart.AddDependency(metallbConfigChart)
	app.Synth()
}
```
With `aiChart := NewAIChart(app, "ai-iac", nil)` we are instructing cdk8s to take that chart and generate the necessary yaml configuration files to be applied to the cluster.

For this demo we also needed to include two charts to setup MetalLB (which were not shown here, but are part of the source code in case you want to take a look) and that's why we're adding a dependency between them with `metallbConfigChart.AddDependency(networkChart)` and `aiChart.AddDependency(metallbConfigChart)`, this will cause the output to be generated in a three separate yaml files that need to be applied in order.

### Generating k8s files
In order to test these changes, we'll need to run the synth command at the root folder of our IaC project
``` bash
cdk8s synth
```
This will create the dist folser if it does not exists and add the yaml file within it:
``` bash
0000-network-setup.k8s.yaml
0001-metallb-config.k8s.yaml
0002-ai-iac.k8s.yaml
```
### Applying changes to cluster
Make sure that kubectl is configured to use your intended cluster before applying any changes, since in our case we're using kind we can simply run `kubectl cluster-info --context kind-kind` but you could also inspect the nodes to make sure you are connected to the right cluster `kubectl get nodes`

To apply the changes we'll need to run the following command:
``` bash
kubectl apply -f dist/0000-network-setup.k8s.yaml
kubectl apply -f dist/0001-metallb-config.k8s.yaml
kubectl apply -f dist/0002-ai-iac.k8s.yaml
```
If your applying more than one file, you'll need to run them individually and wait until the changes are successfully applied to the cluster before running the next one.

You can monitor the state of the installation using the watch command and wait until all pods are in a running state before moving on to the next file.
``` bash
watch -n -2 sudo kubectl get pods -A
```
Once the Label Studio pods are running, you can look into the services to get the url for accessing the app
``` bash
kubectl get services -A
```
Look for a service named after LabelStudio and take note of its external IP // TODO: use an in-cluster dns service to use custom domain names instead of the raw IPs
- Image here

Input the URL into a browser and you should be greeted by the LabelStudio login screen.
- Image here

Now let's add more stuff to our AI chart.

### Adding Jupyter notebook
Similarily to Label Studio, we will be using a helmchart to add Jupyter Hub to our cluster. Jupyter Hub will allow us to create notebooks in order to train and fine tune our model, it will also allow us to import our dataset once we've preprocessed it with Label Studio.

Add the following code after the helm chart for Labels Studio in the `ai-iac.go` file:
``` go
    ...
	/************************** label-studio  ********************************/
	/************************** jupyterhub    ********************************/
	cdk8s.NewHelm(chart, jsii.String("jupyter-hub"), &cdk8s.HelmProps{
		Chart:     jsii.String("jupyterhub/jupyterhub"),
		HelmFlags: &[]*string{jsii.String("--namespace"), jsii.String(namespace)},
		Version:   jsii.String("3.2.1"),
		Values:    &map[string]interface{}{},
	})
	/************************** jupyterhub    ********************************/
    ...
```
If we generate the k8s files again and apply them to the cluste (`cdk8s synth` and `kubectl apply`), we'll notice that it will only add resources related to jupyter hub and did not perform any changes to our previously defined resources. This is because k8s applies manifest changes incrementaly, detecting changes directly from the yaml files and only modifying resources that are new or have changes in their definitions.

- Image here

If we instead destroy our cluster and generate a new one to apply the changes to, we'll see that it will create all of the resources from the beggining.

### Dataset tagging and model training
- Showcase LabelStudio and create a small sample dataset with tags
- create a jupyter notebook and install ultralytics

```
pip install ultralytics 
pip install opencv-python-headless
```
``` python
from ultralytics import YOLO

model = YOLO('yolov8n.pt') 
results = model.train(data='coco128.yaml', epochs=3, imgsz=640)
```
- Use a python script to fine tune the model using the dataset

### Add Custom Serving app
When you're done fine-tuning your model, you'll want to serve it, for that we can use a kubernetes deployment with a service. We'll also use a docker image to run a custom app that will use our model to detect objects from images or video.

Our custom app will be consisting of a front end application and a rest api server to consume the model as defined in the `Dockerfile`
``` dockerfile
FROM golang:alpine3.19 as builder
COPY . /server
WORKDIR /server

RUN go build -o server .
RUN ls

FROM ultralytics/ultralytics:latest as yolov8
WORKDIR /server
COPY --from=builder /server/server .
COPY --from=builder /server/client ./client
COPY --from=builder /server/static ./static

RUN ls


ENV TZ=US/Pacific
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt update && apt upgrade -y
RUN conda update -y ffmpeg

ENTRYPOINT [ "/server/server" ]
```
Underneath, YoloV8 uses `ffmpeg` to process video, so we need to make sure that it's added in the `Dockerfile`
- Explain server functionality and dockerization
- Add registry secrets to iac chart and explain image loading for pods
- Add custom app deployment and service for consumption.

# Sample Use Case 2

# Train model from scratch (no demo)
