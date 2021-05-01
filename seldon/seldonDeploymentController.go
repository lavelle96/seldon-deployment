package seldon

import (
	"context"
	"log"
	"time"

	seldonapi "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	seldonclientset "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SeldonDeploymentController struct {
	seldonClientSet *seldonclientset.Clientset
	namespace       string
	deployment      *seldonapi.SeldonDeployment
}

type SeldonDeploymentControllerInterface interface {
	CreateSeldonDeployment(ctx context.Context, clientSet *seldonclientset.Clientset, namespace string, deployment *seldonapi.SeldonDeployment) error
	WaitUntilReplicaNumberIsReached(ctx context.Context, replicas int32) error
	UpdateNumberOfReplicas(ctx context.Context, replicas int32) error
	DeleteDeployment(ctx context.Context) error
}

func NewSeldonDeploymentController(clientSet *seldonclientset.Clientset, namespace string) SeldonDeploymentControllerInterface {
	return &SeldonDeploymentController{
		seldonClientSet: clientSet,
		namespace:       namespace,
	}
}

func (sD *SeldonDeploymentController) CreateSeldonDeployment(ctx context.Context, clientSet *seldonclientset.Clientset, namespace string, deployment *seldonapi.SeldonDeployment) error {
	deployment, err := sD.seldonClientSet.MachinelearningV1().SeldonDeployments(sD.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	} else {
		sD.deployment = deployment
		sD.seldonClientSet = clientSet
		sD.namespace = namespace
		log.Printf("Created seldon deployment %s successfully\n", deployment.Name)
		return nil
	}
}

func (sD *SeldonDeploymentController) WaitUntilReplicaNumberIsReached(ctx context.Context, replicas int32) error {
	const delay = 4 * time.Second
	available := false
	for !available {
		updatedDeployment, err := sD.seldonClientSet.MachinelearningV1().SeldonDeployments(sD.namespace).Get(ctx, sD.deployment.Name, metav1.GetOptions{})
		sD.deployment = updatedDeployment
		if err != nil {
			return err
		}
		for _, deploymentStatus := range updatedDeployment.Status.DeploymentStatus {
			if deploymentStatus.AvailableReplicas != replicas {
				log.Printf("Deployment currently not at desired replica count: %d, available replicas: %d \n", replicas, deploymentStatus.AvailableReplicas)
			} else {
				available = true
			}
		}

		if !available {
			time.Sleep(delay)
		} else {
			break
		}
	}
	return nil
}

func (sD *SeldonDeploymentController) UpdateNumberOfReplicas(ctx context.Context, replicas int32) error {
	updatedDeployment, err := sD.seldonClientSet.MachinelearningV1().SeldonDeployments(sD.namespace).Get(ctx, sD.deployment.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	updatedDeployment.Spec.Replicas = &replicas
	for i := range updatedDeployment.Spec.Predictors {
		updatedDeployment.Spec.Predictors[i].Replicas = &replicas
		for j := range updatedDeployment.Spec.Predictors[i].ComponentSpecs {
			updatedDeployment.Spec.Predictors[i].ComponentSpecs[j].Replicas = &replicas
		}
	}
	_, err = sD.seldonClientSet.MachinelearningV1().SeldonDeployments(sD.namespace).Update(ctx, updatedDeployment, metav1.UpdateOptions{})
	if err != nil {
		return err
	} else {
		sD.deployment = updatedDeployment
		return nil
	}
}

func (sD *SeldonDeploymentController) DeleteDeployment(ctx context.Context) error {
	err := sD.seldonClientSet.MachinelearningV1().SeldonDeployments(sD.namespace).Delete(ctx, sD.deployment.Name, metav1.DeleteOptions{})
	if err == nil {
		log.Printf("Deleted seldon deployment %s successfully\n", sD.deployment.Name)
		return nil
	} else {
		return err
	}
}
