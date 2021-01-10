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
	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	longyiv1 "longyi.com/api/v1"
)

// OutLoadBalancerReconciler reconciles a OutLoadBalancer object
type OutLoadBalancerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Reconcile controller
func (r *OutLoadBalancerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("outloadbalancer", req.NamespacedName)

	// your logic here
	load := &longyiv1.OutLoadBalancer{}
	if err := r.Get(ctx, req.NamespacedName, load); err != nil {
		log.Info("unable to fetch load: %v", err)
	}

	var labels map[string]string
	labels = make(map[string]string, 1)
	labels["outLoad"] = "true"

	var annotations map[string]string
	annotations = make(map[string]string, 1)
	annotations["outLoad"] = "true"

	loadMetadata := metav1.ObjectMeta{
		Name:        load.Spec.LoadName,
		Namespace:   req.Namespace,
		Labels:      labels,
		Annotations: annotations,
	}
	// create Sevcice
	loadServiceSpec := corev1.ServiceSpec{
		Ports:                    []corev1.ServicePort{{Name: "http", Protocol: corev1.ProtocolTCP, Port: 80, TargetPort: intstr.IntOrString{IntVal: load.Spec.OutPort}, NodePort: 0}},
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

	loadService := corev1.Service{
		Spec:       loadServiceSpec,
		ObjectMeta: loadMetadata,
	}
	r.Create(ctx, &loadService)

	// create Endpoint
	loadEndpintSubset := []corev1.EndpointSubset{
		{
			Addresses: []corev1.EndpointAddress{{
				IP: load.Spec.OutIP,
			}},
			NotReadyAddresses: []corev1.EndpointAddress{},
			Ports: []corev1.EndpointPort{
				{
					Name:     "http",
					Port:     load.Spec.OutPort,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
	loadEndpoint := corev1.Endpoints{
		Subsets:    loadEndpintSubset,
		ObjectMeta: loadMetadata,
	}
	r.Create(ctx, &loadEndpoint)

	// create ingress
	loadIngressSpec := v1beta1.IngressSpec{
		Backend: nil,
		Rules: []v1beta1.IngressRule{
			{
				Host: load.Spec.OutHost,
				IngressRuleValue: v1beta1.IngressRuleValue{
					HTTP: &v1beta1.HTTPIngressRuleValue{
						Paths: []v1beta1.HTTPIngressPath{
							{
								Path: "/",
								Backend: v1beta1.IngressBackend{
									ServiceName: load.Spec.LoadName,
									ServicePort: intstr.IntOrString{IntVal: 80},
								},
							},
						},
					},
				},
			},
		},
	}
	loadIngress := v1beta1.Ingress{
		ObjectMeta: loadMetadata,
		Spec:       loadIngressSpec,
	}
	r.Create(ctx, &loadIngress)

	return ctrl.Result{}, nil
}

// SetupWithManager set
func (r *OutLoadBalancerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&longyiv1.OutLoadBalancer{}).
		Complete(r)
}
