package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/lavelle96/seldon-deployment/seldon"
	v1seldonapi "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	seldonclientset "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var ns, deploymentFilePath string
	flag.StringVar(&ns, "namespace", "", "namespace")
	flag.StringVar(&deploymentFilePath, "filename", "deployment.yml", "seldon deployment filename")
	flag.Parse()

	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	log.Println("Using kubeconfig file: ", kubeconfig)
	log.Println("Using namespace: ", ns)
	log.Println()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	seldonClientSet, err := seldonclientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deployment := &v1seldonapi.SeldonDeployment{}
	err = ParseDeploymentFromFile(deploymentFilePath, deployment)
	if err != nil {
		panic(err)
	}

	seldonDeploymentController := seldon.NewSeldonDeploymentController(seldonClientSet, ns)
	ctx := context.Background()
	err = seldonDeploymentController.CreateSeldonDeployment(ctx, seldonClientSet, ns, deployment)
	if err != nil {
		panic(err)
	}

	err = seldonDeploymentController.WaitUntilReplicaNumberIsReached(ctx, 1)
	if err != nil {
		panic(err)
	}

	replicas := int32(2)

	err = seldonDeploymentController.UpdateNumberOfReplicas(ctx, replicas)
	if err != nil {
		panic(err)
	}

	err = seldonDeploymentController.WaitUntilReplicaNumberIsReached(ctx, replicas)
	if err != nil {
		panic(err)
	}

	err = seldonDeploymentController.DeleteDeployment(ctx)
	if err != nil {
		panic(err)
	}
}

func ParseDeploymentFromFile(path string, deployment *v1seldonapi.SeldonDeployment) error {
	filename, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var decodingSerializer = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	if _, _, err := decodingSerializer.Decode(yamlFile, nil, deployment); err != nil {
		return err
	}
	return nil
}
