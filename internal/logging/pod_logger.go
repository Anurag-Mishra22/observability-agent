package logging

import (
	"bufio"
	"context"
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
		Sink(ParseLog(err.Error(), "error", podName, containerName, namespace))
		return
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		Sink(ParseLog(err.Error(), "error", podName, containerName, namespace))
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

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			parsed := ParseLog(line, "pod", podName, containerName, namespace)
			Sink(parsed)
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
		Sink(ParseLog(err.Error(), "error", "", "", ""))
		return
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		Sink(ParseLog(err.Error(), "error", "", "", ""))
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
			Sink(ParseLog(err.Error(), "error", "", "", targetNamespace))
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
				key := pod.Namespace + "/" + pod.Name + "/" + container.Name

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
