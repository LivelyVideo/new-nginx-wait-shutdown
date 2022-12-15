package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"k8s.io/ingress-nginx/internal/nginx"

	"k8s.io/klog/v2"
)

func main() {

	err := exec.Command("bash", "-c", "pkill -SIGTERM -f nginx-ingress-controller").Run()
	if err != nil {
		klog.Errorf("error terminating ingress controller!: %s", err)
		os.Exit(1)
	}

	healthPort := os.Getenv("HEALTH_PORT")
	// wait for the NGINX process to terminate
	timer := time.NewTicker(time.Second * 30)
	for range timer.C {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%v/metrics", healthPort))
		if err != nil {
			klog.Errorf("error pulling metrics endpoint!: %s", err)
			err := exec.Command("bash", "-c", "pkill -SIGKILL -f nginx-ingress-controller").Run()
			if err != nil {
				klog.Errorf("error killing ingress controller!: %s", err)
				os.Exit(1)
			}
			break
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			klog.Errorf("Unhealthy result from metrics endpoint: %s", resp.StatusCode)
			err := exec.Command("bash", "-c", "pkill -SIGKILL -f nginx-ingress-controller").Run()
			if err != nil {
				klog.Errorf("error killing ingress controller!: %s", err)
				os.Exit(1)
			}
			break
		}

		if !nginx.IsRunning() {
			timer.Stop()
			break
		}
	}
}
