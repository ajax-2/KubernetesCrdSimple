/*

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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	dmeov1 "local.com/demo/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DemoReconciler reconciles a Demo object
type DemoReconciler struct {
	client.Client
	Log logr.Logger
}

// Reconcile demo
func (r *DemoReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("demo", req.NamespacedName)

	// your logic here
	demo := &dmeov1.Demo{}
	if err := r.Get(ctx, req.NamespacedName, demo); err != nil {
		log.Info("unable to fetch demo: %v", err)
	}

	var labels map[string]string
	labels = make(map[string]string, 2)
	labels["demo"] = "true"
	labels["crd"] = "true"

	var annotations map[string]string
	annotations = make(map[string]string, 1)
	annotations["demo"] = "true"

	demoMetadata := metav1.ObjectMeta{
		Name:        demo.Spec.DemoName,
		Namespace:   req.Namespace,
		Labels:      labels,
		Annotations: annotations,
	}

	demoServiceSpec := corev1.ServiceSpec{
		Ports:                    []corev1.ServicePort{{Name: "http", Protocol: corev1.ProtocolTCP, Port: demo.Spec.DemoPort, TargetPort: intstr.IntOrString{Type: 0, IntVal: 80, StrVal: "80"}, NodePort: 0}},
		Selector:                 nil,
		ClusterIP:                "",
		Type:                     corev1.ServiceTypeClusterIP,
		ExternalIPs:              []string{},
		SessionAffinity:          "",
		LoadBalancerIP:           "",
		LoadBalancerSourceRanges: []string{},
		ExternalName:             "",
		ExternalTrafficPolicy:    "",
		HealthCheckNodePort:      0,
		PublishNotReadyAddresses: false,
		SessionAffinityConfig:    &corev1.SessionAffinityConfig{},
	}
	demoService := corev1.Service{
		Spec:       demoServiceSpec,
		ObjectMeta: demoMetadata,
	}
	r.Create(ctx, &demoService)

	return ctrl.Result{}, nil
}

// SetupWithManager Demo
func (r *DemoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dmeov1.Demo{}).
		Complete(r)
}
