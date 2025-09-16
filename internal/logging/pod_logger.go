package logging

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const MaxConcurrentTails = 20 // max simultaneous pod log tails

var (
	activePods = make(map[string]bool)
	mu         sync.Mutex
	sem        = make(chan struct{}, MaxConcurrentTails) // semaphore to limit concurrency
)

// TailPodLogs streams logs from a single pod/container safely
func TailPodLogs(namespace, podName, containerName string) {
	defer func() { <-sem }() // release semaphore when done

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

	options := &corev1.PodLogOptions{Container: containerName, Follow: true}

	for {
		req := clientset.CoreV1().Pods(namespace).GetLogs(podName, options)
		stream, err := req.Stream(context.Background())
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		reader := bufio.NewReader(stream)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break // reconnect if stream ends
			}

			line = cleanLine(line)
			if line == "" {
				continue // skip empty lines
			}

			// Check if line is already JSON
			if !isJSON(line) {
				event := LogEvent{
					Timestamp: time.Now(),
					Line:      line,
					Pod:       podName,
					Container: containerName,
					Namespace: namespace,
				}
				data, _ := json.Marshal(event)
				fmt.Println(string(data))
			} else {
				fmt.Println(line)
			}
		}

		stream.Close()
		time.Sleep(2 * time.Second) // retry if stream closed
	}
}

// TailAllPods dynamically tails logs from all pods in all namespaces
// Optionally provide namespace="", or labelSelector="" to filter
func TailAllPods(namespace string, labelSelector string) {
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

	targetNamespace := namespace
	if targetNamespace == "" {
		targetNamespace = metav1.NamespaceAll
	}

	for {
		watcher, err := clientset.CoreV1().Pods(targetNamespace).Watch(context.Background(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			fmt.Println("Error creating pod watcher:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		ch := watcher.ResultChan()
		for event := range ch {
			pod, ok := event.Object.(*corev1.Pod)
			if !ok || pod.Status.Phase != corev1.PodRunning {
				continue
			}

			// Skip system pods and agent itself
			if pod.Namespace == "kube-system" || pod.Namespace == "kube-public" || pod.Name == "observability-agent" {
				continue
			}

			for _, container := range pod.Spec.Containers {
				key := fmt.Sprintf("%s/%s/%s", pod.Namespace, pod.Name, container.Name)

				mu.Lock()
				if _, exists := activePods[key]; !exists {
					activePods[key] = true
					sem <- struct{}{} // acquire semaphore
					go TailPodLogs(pod.Namespace, pod.Name, container.Name)
				}
				mu.Unlock()
			}
		}

		watcher.Stop()
		time.Sleep(1 * time.Second)
	}
}

// cleanLine trims whitespace and newline characters
func cleanLine(line string) string {
	return strings.TrimSpace(line)
}

// isJSON checks if a string is valid JSON
func isJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}
