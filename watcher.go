package main

import (
    "context"
    "log"
    "time"

    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
)

func watchPods(clientset *kubernetes.Clientset, cfg HealPolicy) {
    watcher, err := clientset.CoreV1().Pods("default").Watch(context.TODO(), metav1.ListOptions{})
    if err != nil {
        log.Fatalf("❌ Failed to start watcher: %v", err)
    }

    healed := make(map[string]time.Time)
    retryCount := make(map[string]int)

    for event := range watcher.ResultChan() {
        pod, ok := event.Object.(*corev1.Pod)
        if !ok {
            continue
        }

        name := pod.Name
        labels := pod.Labels
        phase := pod.Status.Phase
        reason := ""

        for _, cs := range pod.Status.ContainerStatuses {
            if cs.State.Waiting != nil {
                reason = cs.State.Waiting.Reason
            } else if cs.State.Terminated != nil {
                reason = cs.State.Terminated.Reason
            }
        }

        // Skip ignored reasons
        for _, r := range cfg.IgnoreReasons {
            if reason == r {
                log.Printf("🟡 Ignoring pod '%s' (reason: %s)\n", name, reason)
                goto NEXT
            }
        }

        // Handle unhealthy pods
        if reason == "CrashLoopBackOff" || reason == "Error" || phase == corev1.PodFailed {
            if labels["autopilot"] == "true" {
                // Cooldown check
                if last, ok := healed[name]; ok && time.Since(last) < time.Duration(cfg.CooldownSeconds)*time.Second {
                    continue
                }

                // Retry limit check
                retryCount[name]++
                if retryCount[name] > cfg.MaxRetries {
                    log.Printf("🚫 Max retries reached for pod '%s'. Skipping further heals.\n", name)
                    goto NEXT
                }

                log.Printf("🚨 Healing pod '%s' (reason: %s, attempt %d/%d)\n",
                    name, reason, retryCount[name], cfg.MaxRetries)

                err := clientset.CoreV1().Pods("default").Delete(context.TODO(), name, metav1.DeleteOptions{})
                if err != nil {
                    log.Printf("❌ Failed to delete pod '%s': %v\n", name, err)
                } else {
                    healed[name] = time.Now()
                    log.Printf("✅ Deleted pod '%s' for recovery.\n", name)
                }
            }
        }

    NEXT:
    }
}
