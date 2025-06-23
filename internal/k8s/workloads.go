package k8s

import (
	"context"

	"github.com/tabed23/k8s-resource-tuner/internal/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListDeployments(clientset *kubernetes.Clientset, namespaces string) ([]models.WorkLoad, error) {
	deployClient := clientset.AppsV1().Deployments(namespaces)
	deployments, err := deployClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var workloads []models.WorkLoad
	for _, d := range deployments.Items {
		var containers []models.ContainerSpec
		for _, c := range d.Spec.Template.Spec.Containers {
			containers = append(containers, models.ContainerSpec{
				Name: c.Name,
				Resources: models.ResourceConfig{
					Request: c.Resources.Requests,
					Limits:  c.Resources.Limits,
				},
			})
		}
		workload := models.WorkLoad{
			Namespace:  d.Namespace,
			Name:       d.Name,
			Kind:       "Deployment",
			Containers: containers,
			Labels:     d.Labels,
		}
		workloads = append(workloads, workload)
	}
	return workloads, nil
}
