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
	"os"
	"os/exec"
	"time"

	"k8s.io/ingress-nginx/internal/nginx"
	"k8s.io/klog/v2"
	"net/http"
)

func main() {
	// We are starting this in preStop -  its sending a term to the nginx-ingress-controller, which is actually a wrapper
	// for the nginx process.  But the nginx process does receive a term signal in return.
	err := exec.Command("bash", "-c", "pkill -SIGTERM -f nginx-ingress-controller").Run()
	if err != nil {
		klog.Errorf("error terminating ingress controller!: %s", err)
		os.Exit(1)
	}

	healthPort := os.Getenv("HEALTH_PORT")
	hostName := os.Getenv("HOSTNAME")
	// wait for the NGINX process to terminate
	// For loop setup to wait monitor health endpoints.  If the nginx service is so bad that it can't return
	// the /metrics endpoint from th health port - it should be shutdown ASAP.

	timer := time.NewTicker(time.Second * 30)
	for range timer.C {
		// The metrics endpoint is used and the response code is checked for simplicity and to check if the
		// service is truly health or not.  Possible improvement is to check active connections - and shutdown
		// the service when active connections are 0
		resp, err := http.Get(fmt.Sprintf("http://localhost:%v/metrics", healthPort))

		if err != nil {
			klog.Errorf("error pulling health check!: %s", err)
			err := exec.Command("bash", "-c", "pkill -SIGKILL -f nginx-ingress-controller").Run()
			if err != nil {
				klog.Errorf("error killing ingress controller!: %s", err)
				os.Exit(1)
			}
			// If the prestop script is failing to shutdown the service when it detects it being unhealthy -
			// we need to break the loop and exit the script.  Fallback on k8s and nginx to shut the service down
			break
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			klog.Errorf("Unhealthy result from health check: %s", resp.StatusCode)
			err := exec.Command("bash", "-c", "pkill -SIGKILL -f nginx-ingress-controller").Run()
			if err != nil {
				klog.Errorf("error killing ingress controller!: %s", err)
				os.Exit(1)
			}
			break
		}

		// No reason for the script to continue if nginx has stopped.
		if !nginx.IsRunning() {
			timer.Stop()
			break
		}
	}
}

