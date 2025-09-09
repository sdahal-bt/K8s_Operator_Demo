package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func randomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func rotateSecret(clientset *kubernetes.Clientset, secret *corev1.Secret) error {
	// Rotate all string data fields
	for k := range secret.Data {
		secret.Data[k] = []byte(randomString(16))
	}
	_, err := clientset.CoreV1().Secrets(secret.Namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	return err
}

func main() {
	namespace := os.Getenv("WATCH_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Starting secretswatcher-operator in namespace: %s\n", namespace)

	for {
		watcher, err := clientset.CoreV1().Secrets(namespace).Watch(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error creating watcher: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		ch := watcher.ResultChan()
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case event, ok := <-ch:
				if !ok {
					break
				}
				if event.Type == watch.Added || event.Type == watch.Modified {
					secret := event.Object.(*corev1.Secret)
					fmt.Printf("Detected secret: %s/%s\n", secret.Namespace, secret.Name)
				}
			case <-ticker.C:
				secrets, err := clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					fmt.Printf("Error listing secrets: %v\n", err)
					continue
				}
				for _, secret := range secrets.Items {
					fmt.Printf("Rotating secret: %s/%s\n", secret.Namespace, secret.Name)
					err := rotateSecret(clientset, &secret)
					if err != nil {
						fmt.Printf("Error rotating secret %s/%s: %v\n", secret.Namespace, secret.Name, err)
					}
				}
			}
		}
	}
}
