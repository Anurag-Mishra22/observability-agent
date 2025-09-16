package logging

import (
	"context"
	"fmt"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubernetesMetadataFilter struct {
	clientset *kubernetes.Clientset
	cache     map[string]map[string]string // podName → labels
	mu        sync.RWMutex
}

func NewKubernetesMetadataFilter() (*KubernetesMetadataFilter, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &KubernetesMetadataFilter{
		clientset: clientset,
		cache:     make(map[string]map[string]string),
	}, nil
}

func (f *KubernetesMetadataFilter) Apply(event LogEvent) (*LogEvent, bool) {
	// If already enriched, skip
	if event.Extra == nil {
		event.Extra = make(map[string]interface{})
	}

	// Try cached metadata
	f.mu.RLock()
	if labels, ok := f.cache[event.Pod]; ok {
		event.Extra["labels"] = labels
		f.mu.RUnlock()
		return &event, true
	}
	f.mu.RUnlock()

	// Fetch metadata from Kubernetes API
	pod, err := f.clientset.CoreV1().Pods(event.Namespace).Get(context.Background(), event.Pod, metav1.GetOptions{})
	if err != nil {
		fmt.Println("Metadata fetch error:", err)
		return &event, true
	}

	// Cache labels
	labels := pod.GetLabels()
	f.mu.Lock()
	f.cache[event.Pod] = labels
	f.mu.Unlock()

	event.Extra["labels"] = labels
	event.Extra["nodeName"] = pod.Spec.NodeName
	event.Extra["owner"] = pod.OwnerReferences

	return &event, true
}
