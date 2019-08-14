package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mattbaird/jsonpatch"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
)

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

// Config contains the server (the webhook) cert and key.
type Config struct {
	CertFile string
	KeyFile  string
}
type admitFunc func(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse


func configTLS(config Config) *tls.Config {
	sCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		klog.Fatalf("config=%#v Error: %v", config, err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
		// TODO: uses mutual tls after we agree on what cert the apiserver should use.
		//ClientAuth: tls.RequireAndVerifyClientCert,
	}
}

func apply(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	klog.Info("Entering apply in ExtendedResourceToleration webhook")
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if ar.Request.Resource != podResource {
		klog.Errorf("expect resource to be %s", podResource)
		return nil
	}

	raw := ar.Request.Object.Raw
	pod := corev1.Pod{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
		klog.Error(err)
		return toAdmissionResponse(err)
	}
	reviewResponse := v1beta1.AdmissionResponse{}
	reviewResponse.Allowed = true
	podCopy := pod.DeepCopy()
	klog.V(1).Infof("Examining pod: %v\n", pod.GetName())

	// Ignore if exclusion annotation is present
	if podAnnotations := pod.GetAnnotations(); podAnnotations != nil {
		klog.Info(fmt.Sprintf("Looking at pod annotations, found: %v", podAnnotations))
		if _, isMirrorPod := podAnnotations[corev1.MirrorPodAnnotationKey]; isMirrorPod {
			return &reviewResponse
		}
	}

	// find resource requests and add toleration
	// Copied from : https://github.com/kubernetes/kubernetes/blob/master/plugin/pkg/admission/extendedresourcetoleration/admission.go
	resources := sets.String{}
	for _, container := range pod.Spec.Containers {
		for resourceName := range container.Resources.Requests {
			if isExtendedResourceName(resourceName) {
				resources.Insert(string(resourceName))
			}
		}
	}
	for _, container := range pod.Spec.InitContainers {
		for resourceName := range container.Resources.Requests {
			if isExtendedResourceName(resourceName) {
				resources.Insert(string(resourceName))
			}
		}
	}

	if resources.Len() == 0 {
		return &reviewResponse
	}

	// Doing .List() so that we get a stable sorted list.
	// This allows us to test adding tolerations for multiple extended resources.
	for _, resource := range resources.List() {
		if !addOrUpdateTolerationInPod(&pod, &corev1.Toleration{
			Key:      resource,
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoSchedule,
		}) {
			return &reviewResponse
		}
		klog.Infof("applied extendedresourcetoleration: %s successfully on Pod: %+v ", resource, pod.GetName())
	}

	podCopyJSON, err := json.Marshal(podCopy)
	if err != nil {
		return toAdmissionResponse(err)
	}
	podJSON, err := json.Marshal(pod)
	if err != nil {
		return toAdmissionResponse(err)
	}
	klog.Infof("PodCopy json: %s ", podCopyJSON)
	klog.Infof("pod json: %s ", podJSON)
	jsonPatch, err := jsonpatch.CreatePatch(podCopyJSON, podJSON)
	if err != nil {
		klog.Infof("patch error: %+v", err)
		return toAdmissionResponse(err)
	}
	jsonPatchBytes, _ := json.Marshal(jsonPatch)
	klog.Infof("jsonPatch json: %s", jsonPatchBytes)

	reviewResponse.Patch = jsonPatchBytes
	pt := v1beta1.PatchTypeJSONPatch
	reviewResponse.PatchType = &pt
	return &reviewResponse
}

func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	var reviewResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		klog.Error(err)
		reviewResponse = toAdmissionResponse(err)
	} else {
		reviewResponse = admit(ar)
	}

	response := v1beta1.AdmissionReview{}
	if reviewResponse != nil {
		response.Response = reviewResponse
		response.Response.UID = ar.Request.UID
	}
	// reset the Object and OldObject, they are not needed in a response.
	ar.Request.Object = runtime.RawExtension{}
	ar.Request.OldObject = runtime.RawExtension{}

	resp, err := json.Marshal(response)
	if err != nil {
		klog.Error(err)
	}
	if _, err := w.Write(resp); err != nil {
		klog.Error(err)
	}
}

func serveERT(w http.ResponseWriter, r *http.Request) {
	serve(w, r, apply)
}

func main() {
	var config Config
	flag.StringVar(&config.CertFile, "tlsCertFile", "/etc/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&config.KeyFile, "tlsKeyFile", "/etc/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()
	klog.InitFlags(nil)

	http.HandleFunc("/apply-ert", serveERT)

	server := &http.Server{
		Addr:      ":443",
		TLSConfig: configTLS(config),
	}
	klog.Info(fmt.Sprintf("About to start serving webhooks: %#v", server))
	server.ListenAndServeTLS("", "")
}
