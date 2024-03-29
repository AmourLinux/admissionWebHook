/*
Copyright 2018 The Kubernetes Authors.

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
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	// TODO: try this library to see if it generates correct json patch
	// https://github.com/mattbaird/jsonpatch
)

// toAdmissionResponse is a helper function to create an AdmissionResponse
// with an embedded error
func toAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

// admitFunc is the type we use for all of our validators and mutators
type admitFunc func(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse

// serve handles the http portion of a request prior to handing to an admit
// function
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

	klog.V(2).Info(fmt.Sprintf("handling request: %s", body))

	// The AdmissionReview that was sent to the webhook
	requestedAdmissionReview := v1beta1.AdmissionReview{}

	// The AdmissionReview that will be returned
	responseAdmissionReview := v1beta1.AdmissionReview{}

	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		klog.Error(err)
		responseAdmissionReview.Response = toAdmissionResponse(err)
	} else {
		// pass to admitFunc
		responseAdmissionReview.Response = admit(requestedAdmissionReview)
	}

	// Return the same UID
	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	klog.V(2).Info(fmt.Sprintf("sending response: %v", responseAdmissionReview.Response))

	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		klog.Error(err)
	}
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}

func serveAlwaysDeny(w http.ResponseWriter, r *http.Request) {
	serve(w, r, alwaysDeny)
}

func serveAddLabel(w http.ResponseWriter, r *http.Request) {
	serve(w, r, addLabel)
}

func servePods(w http.ResponseWriter, r *http.Request) {
	serve(w, r, admitPods)
}

func serveAttachingPods(w http.ResponseWriter, r *http.Request) {
	serve(w, r, denySpecificAttachment)
}

func serveMutatePods(w http.ResponseWriter, r *http.Request) {
	serve(w, r, mutatePods)
}

func serveConfigmaps(w http.ResponseWriter, r *http.Request) {
	serve(w, r, admitConfigMaps)
}

func serveMutateConfigmaps(w http.ResponseWriter, r *http.Request) {
	serve(w, r, mutateConfigmaps)
}

func serveCustomResource(w http.ResponseWriter, r *http.Request) {
	serve(w, r, admitCustomResource)
}

func serveMutateCustomResource(w http.ResponseWriter, r *http.Request) {
	serve(w, r, mutateCustomResource)
}

//func serveCRD(w http.ResponseWriter, r *http.Request) {
//	serve(w, r, admitCRD)
//}

func serveCheckRepo(w http.ResponseWriter, r *http.Request) {
	serve(w, r, checkRepo)
}

func main() {
	var err error
	var config = Config{
		CertFile: "/Users/caipengfei/Applications/goTemp/pki/amourlinux-admission.pem",
		KeyFile: "/Users/caipengfei/Applications/goTemp/pki/amourlinux-admission-key.pem",
	}
	config.InitFlags()

	klog.InitFlags(flag.CommandLine)
	if err = flag.Set("v", "2"); err != nil {
		klog.Fatalf("Failed to set flag: %v\n", err)
	}
	if err = flag.Set("logtostderr", "true"); err != nil {
		klog.Fatalf("Failed to set flag: %v\n", err)
	}

	flag.Parse()

	http.HandleFunc("/always-deny", serveAlwaysDeny)
	http.HandleFunc("/add-label", serveAddLabel)
	http.HandleFunc("/pods", servePods)
	http.HandleFunc("/pods/attach", serveAttachingPods)
	http.HandleFunc("/mutating-pods", serveMutatePods)
	http.HandleFunc("/configmaps", serveConfigmaps)
	http.HandleFunc("/mutating-configmaps", serveMutateConfigmaps)
	http.HandleFunc("/custom-resource", serveCustomResource)
	http.HandleFunc("/mutating-custom-resource", serveMutateCustomResource)
	//http.HandleFunc("/crd", serveCRD)
	http.HandleFunc("/pods/check-repo", serveCheckRepo)
	server := &http.Server{
		Addr:      ":443",
		TLSConfig: configTLS(config),
	}

	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			klog.Fatalf("Failed to listen and serve webhook server: %v\n", err)
		}
	}()
	klog.V(2).Infoln("Server started port: 443")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	<-signalChan

	klog.V(2).Infof("Got OS shutdown signal, shutting down webhook server gracefully...\n")
	server.Shutdown(context.Background())
}
