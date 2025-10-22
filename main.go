package main

import (
    "fmt"
    "log"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

func main() {
    // Load HealingPolicy from CRD
    cfg := LoadHealingPolicy()

    // Build Kubernetes clientset
    kubeconfig := "/Users/adhirajchanana/.kube/config"
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        log.Fatalf("❌ Failed to load kubeconfig: %v", err)
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatalf("❌ Failed to create clientset: %v", err)
    }

    fmt.Println("🧠 Autopilot Controller starting with CRD policies...")
    watchPods(clientset, cfg)
}
