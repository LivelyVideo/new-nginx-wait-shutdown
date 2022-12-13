/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
