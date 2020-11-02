package main

import (
	"context"
	"encoding/json"
	"github.com/k8s-autoops/autoops"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"net/http"
	"os"
)

const (
	AnnotationKeyIngressClass = "autoops.enforce-ingress-class"
)

func exit(err *error) {
	if *err != nil {
		log.Println("exited with error:", (*err).Error())
		os.Exit(1)
	} else {
		log.Println("exited")
	}
}

func main() {
	var err error
	defer exit(&err)

	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	var client *kubernetes.Clientset
	if client, err = autoops.InClusterClient(); err != nil {
		return
	}

	s := &http.Server{
		Addr: ":443",
		Handler: autoops.NewMutatingAdmissionHTTPHandler(
			func(ctx context.Context, request *admissionv1.AdmissionRequest, patches *[]map[string]interface{}) (err error) {
				var buf []byte
				if buf, err = request.Object.MarshalJSON(); err != nil {
					return
				}
				var ing networkingv1beta1.Ingress
				if err = json.Unmarshal(buf, &ing); err != nil {
					return
				}
				// 获取命名空间并检查特定注解
				var ns *corev1.Namespace
				if ns, err = client.CoreV1().Namespaces().Get(ctx, request.Namespace, metav1.GetOptions{}); err != nil {
					return
				}
				if ns.Annotations == nil {
					return
				}
				if ns.Annotations[AnnotationKeyIngressClass] == "" {
					return
				}
				// 增加注解
				if ing.Annotations == nil {
					*patches = append(*patches, map[string]interface{}{
						"op":    "replace",
						"path":  "/metadata/annotations",
						"value": map[string]interface{}{},
					})
				}
				*patches = append(*patches, map[string]interface{}{
					"op":    "replace",
					"path":  "/metadata/annotations/kubernetes.io~1ingress.class",
					"value": ns.Annotations[AnnotationKeyIngressClass],
				})
				return
			},
		),
	}

	if err = autoops.RunAdmissionServer(s); err != nil {
		return
	}
}
