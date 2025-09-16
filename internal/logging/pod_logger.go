package logging

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// TailPodLogs streams logs from a single pod/container
func TailPodLogs(namespace, podName, containerName string) error {
	// Create in-cluster config (use clientcmd.BuildConfigFromFlags for outside cluster)
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	options := &corev1.PodLogOptions{
		Container: containerName,
		Follow:    true,
	}

	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, options)
	stream, err := req.Stream(context.Background())
	if err != nil {
		return err
	}
	defer stream.Close()

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()
		event := LogEvent{
			Timestamp: time.Now(),
			Line:      line,
			Pod:       podName,
			Container: containerName,
			Namespace: namespace,
		}
		data, _ := json.Marshal(event)
		fmt.Println(string(data))
	}

	return scanner.Err()
}

func TailAllPods(namespace string) {
	config, _ := rest.InClusterConfig()
	clientset, _ := kubernetes.NewForConfig(config)

	pods, _ := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			go TailPodLogs(namespace, pod.Name, container.Name)
		}
	}
}
