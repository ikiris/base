package server

import (
	"context"
	"fmt"

	v1a "k8s.io/api/apps/v1"
	v1m "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type kwrapper struct {
	clientSet *kubernetes.Clientset
}

func newK8(kubeconfig string) (*kwrapper, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error generating kubeconfig: %w", err)
	}

	// create the clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("kube config err: %w", err)
	}

	return &kwrapper{clientSet: clientSet}, nil
}

func (s *kwrapper) getDeployment(ctx context.Context, ns, dn string) (*v1a.Deployment, error) {
	d, err := s.clientSet.AppsV1().Deployments(ns).Get(ctx, dn, v1m.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed retrieving deployment: %w", err)
	}

	return d, nil
}

func (s *kwrapper) getDeployments(ctx context.Context, ns string) (*v1a.DeploymentList, error) {
	l, err := s.clientSet.AppsV1().Deployments(ns).List(ctx, v1m.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed retrieving deployments: %w", err)
	}
	return l, nil
}
