package server

import (
	"context"
	"errors"

	pb "github.com/teraptra/base/testserver-grpc/proto"
)

type server struct {
	pb.DepManServiceServer
	kw *kwrapper
}

// New generates a blah server.
func New(kubeconfig string) (*server, error) {
	kw, err := newK8(kubeconfig)
	if err != nil {
		return nil, err
	}

	return &server{kw: kw}, nil
}

// Get fetches basic deployment info by name and namespace.
func (s *server) Get(ctx context.Context, req *pb.ReplicaRequest) (*pb.Deployment, error) {
	d, err := s.kw.getDeployment(ctx, req.GetNamespace(), req.GetName())
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

// List fetches basic list of all deployments by namespace.
func (s *server) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	dl, err := s.kw.getDeployments(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	ret := &pb.ListResponse{}
	var deps []*pb.Deployment
	for _, d := range dl.Items {
		deps = append(deps, &pb.Deployment{
			Name: d.GetName(),
		})
	}
	ret.Deployment = deps
	return ret, nil
}
