package main

import (
    "context"
    "log"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/client-go/dynamic"
    "k8s.io/client-go/rest"
)

type HealPolicy struct {
    MaxRetries      int
    CooldownSeconds int
    IgnoreReasons   []string
}

func LoadHealingPolicy() HealPolicy {
    // Try in-cluster config, fallback to kubeconfig
    config, err := rest.InClusterConfig()
    if err != nil {
        log.Printf("⚙️ Running outside cluster, using default kubeconfig.")
        return HealPolicy{MaxRetries: 3, CooldownSeconds: 15, IgnoreReasons: []string{"OOMKilled"}}
    }

    dynClient, err := dynamic.NewForConfig(config)
    if err != nil {
        log.Fatalf("❌ Failed to create dynamic client: %v", err)
    }

    gvr := schema.GroupVersionResource{
        Group:    "autopilot.io",
        Version:  "v1",
        Resource: "healingpolicies",
    }

    hp, err := dynClient.Resource(gvr).Namespace("default").Get(context.TODO(), "default-policy", metav1.GetOptions{})
    if err != nil {
        log.Fatalf("❌ Failed to get HealingPolicy: %v", err)
    }

    spec := hp.Object["spec"].(map[string]interface{})
    policy := HealPolicy{
        MaxRetries:      int(spec["maxRetries"].(int64)),
        CooldownSeconds: int(spec["cooldownSeconds"].(int64)),
    }

    if v, ok := spec["ignoreReasons"]; ok {
        for _, r := range v.([]interface{}) {
            policy.IgnoreReasons = append(policy.IgnoreReasons, r.(string))
        }
    }

    log.Printf("✅ Loaded HealingPolicy from cluster: %+v\n", policy)
    return policy
}
