package main

import (
    "log"
    "os"
    "path/filepath"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
)

func getKubeConfig() *rest.Config {
    // Try to use in-cluster config first (for when running inside Kubernetes)
    config, err := rest.InClusterConfig()
    if err == nil {
        log.Println("🟢 Using in-cluster configuration")
        return config
    }


    // Try environment variable (Docker / external testing)
    kubeconfigEnv := os.Getenv("KUBECONFIG")
    if kubeconfigEnv != "" {
        log.Printf("⚙️ Using kubeconfig from environment: %s", kubeconfigEnv)
        config, err = clientcmd.BuildConfigFromFlags("", kubeconfigEnv)
        if err == nil {
            return config
        }
        log.Printf("❌ Failed to load kubeconfig from env: %v", err)
    }

    // Fallback: default location
    home, _ := os.UserHomeDir()
    defaultPath := filepath.Join(home, ".kube", "config")
    log.Printf("⚙️ Using default kubeconfig: %s", defaultPath)
    config, err = clientcmd.BuildConfigFromFlags("", defaultPath)
    if err != nil {
        log.Fatalf("❌ Failed to load kubeconfig: %v", err)
    }

    return config
}

func main() {
    config := getKubeConfig()

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatalf("❌ Failed to create clientset: %v", err)
    }
    _=clientset

    log.Println("✅ Kubernetes client initialized successfully!")
    // then continue to start your watcher / autopilot logic here
}
