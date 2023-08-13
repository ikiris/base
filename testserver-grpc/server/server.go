package server

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/teraptra/base/testserver-grpc/proto"
	v1a "k8s.io/api/apps/v1"
	v1m "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type server struct {
	pb.DepManServiceServer
	clientSet *kubernetes.Clientset
}

// New generates a blah server.
func New(kubeconfig string) (*server, error) {
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

	s := &server{
		clientSet: clientSet,
	}

	return s, err
}

func (s *server) getDeployment(ctx context.Context, ns, dn string) (*v1a.Deployment, error) {
	d, err := s.clientSet.AppsV1().Deployments(ns).Get(ctx, dn, v1m.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed retrieving deployment: %w", err)
	}

	return d, nil
}

func (s *server) Get(ctx context.Context, req *pb.ReplicaRequest) (*pb.Deployment, error) {
	d, err := s.getDeployment(ctx, "core", "server")
	if err != nil {
		return nil, err
	}

	ret := &pb.Deployment{
		Name:     d.Name,
		Replicas: uint32(*d.Spec.Replicas),
	}
	return ret, nil
}

func (s *server) Set(ctx context.Context, req *pb.ReplicaRequest) (*pb.ReplicaResponse, error) {
	return nil, errors.ErrUnsupported
}

func (s *server) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	return nil, errors.ErrUnsupported
}
