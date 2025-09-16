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

// TailAllPods tails logs from all pods in a namespace.
// If namespace is empty, tails logs from all namespaces.
func TailAllPods(namespace string) {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("Error creating in-cluster config:", err)
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Error creating Kubernetes clientset:", err)
		return
	}

	listOptions := metav1.ListOptions{}
	targetNamespace := namespace
	if namespace == "" {
		targetNamespace = metav1.NamespaceAll
	}

	podList, err := clientset.CoreV1().Pods(targetNamespace).List(context.Background(), listOptions)
	if err != nil {
		fmt.Printf("Error listing pods in namespace '%s': %v\n", targetNamespace, err)
		return
	}

	for _, pod := range podList.Items {
		for _, container := range pod.Spec.Containers {
			go func(ns, podName, containerName string) {
				if err := TailPodLogs(ns, podName, containerName); err != nil {
					fmt.Printf("Error tailing logs for pod %s container %s: %v\n", podName, containerName, err)
				}
			}(pod.Namespace, pod.Name, container.Name)
		}
	}
}
