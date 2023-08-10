package server

import (
	"context"
	"fmt"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type ProdServer interface {
	ListenAndServe(context.Context) error
	RegisterHealthZHandler(http.Handler)
}

type server struct{}

// New generates a blah server.
func New(kubeconfig string) (*server, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error generating kubeconfig: %w", err)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("kube config err: %w", err)
	}

	s := &server{}

	return s, err
}

// healthzHandler is an HTTP handler for the healthz API.
type healthzHandler struct {
	clientset *kubernetes.Clientset
}

// ServeHTTP implements http.Handler
func (h *healthzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pods, err := h.clientset.CoreV1().Pods("").List(r.Context(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintf(w, "healthz: OK\nThere are %d pods in the cluster\n", len(pods.Items))
}

