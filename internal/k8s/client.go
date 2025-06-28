package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Change it to in Cluster whe deploy
// config, err := rest.InClusterConfig()
// 	if err != nil {
// 		return nil, err
// 	}'


func InitKubeClient() (*kubernetes.Clientset, error) {
	kubeconfig := ""

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}
