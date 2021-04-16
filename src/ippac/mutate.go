package main

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"os"
	"strconv"
	"k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	admissionv1 "k8s.io/api/admission/v1"
)

func isKubeNamespace(ns string) bool {
	return ns == metav1.NamespacePublic || ns == metav1.NamespaceSystem

}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func (mr *myServerHandler) mutserve(w http.ResponseWriter, r *http.Request) {

	var Body []byte
	if r.Body != nil {
		if data , err := ioutil.ReadAll(r.Body); err == nil {
			Body = data
		}
	}

	if len(Body) == 0 {
		fmt.Fprintf(os.Stderr, "Unable to retrieve Body from API")
		http.Error(w,"Empty Body", http.StatusBadRequest)
	}

	fmt.Fprintf(os.Stdout,"Received Request\n")

	if r.URL.Path != "/mutate" {
		fmt.Fprintf(os.Stderr, "Not a Validate URL Path")
		http.Error(w, "Not A Validate URL Path", http.StatusBadRequest)
	}

	// Read the Response from the Kubernetes API and place it in the Request 

	arRequest := &v1.AdmissionReview{}
	if err := json.Unmarshal(Body, arRequest); err != nil {
		fmt.Fprintf(os.Stderr, "Error Unmarsheling the Body request")
		http.Error(w, "Error Unmarsheling the Body request", http.StatusBadRequest)
		return
	}

	raw := arRequest.Request.Object.Raw
	obj := corev1.Pod{}

	if !isKubeNamespace(arRequest.Request.Namespace) {
		
		if err := json.Unmarshal(raw, &obj); err != nil {
			fmt.Fprintf(os.Stderr, "Error , unable to Deserializing Pod")
			http.Error(w,"Error , unable to Deserializing Pod", http.StatusBadRequest)
			return
		}
	} else {
			fmt.Fprintf(os.Stderr, "Error , unauthorized Namespace")
			http.Error(w,"Error , unauthorized Namespace", http.StatusBadRequest)
			return
	}

	containers := obj.Spec.Containers

	arResponse := v1.AdmissionReview{
		Response: &v1.AdmissionResponse{
				UID: arRequest.Request.UID,
		},
	}

	var patches []patchOperation

	fmt.Fprintf(os.Stdout, "Starting the Loop for containers\n")
	
	for i , container := range containers {
		fmt.Fprintf(os.Stdout, "container[%d] = %s imagePullPolicy = %s\n", i, container.Name , container.ImagePullPolicy)

   		if containers[i].ImagePullPolicy == "Never" || containers[i].ImagePullPolicy == "IfNotPresent"  {
				
			patches = append(patches, patchOperation{
				    	Op: "replace",
						Path: "/spec/containers/"+ strconv.Itoa(i) +"/imagePullPolicy",
						Value: "Always",
			})
		}
	}

	fmt.Fprintf(os.Stdout, "the Json Is : \"%s\"\n", patches)

	patchBytes, err := json.Marshal(patches)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't encode Patches: %v", err)
		http.Error(w, fmt.Sprintf("couldn't encode Patches: %v", err), http.StatusInternalServerError)
		return
	}

	v1JSONPatch := admissionv1.PatchTypeJSONPatch
	arResponse.APIVersion = "admission.k8s.io/v1"
	arResponse.Kind = arRequest.Kind
	arResponse.Response.Allowed = true
	arResponse.Response.Patch = patchBytes
	arResponse.Response.PatchType = &v1JSONPatch
	
	resp, rErr := json.Marshal(arResponse)

	if rErr != nil {
		fmt.Fprintf(os.Stderr, "Can't encode response: %v", rErr)
		http.Error(w, fmt.Sprintf("couldn't encode response: %v", rErr), http.StatusInternalServerError)
	}

	if _ , wErr := w.Write(resp); wErr != nil {
		fmt.Fprintf(os.Stderr, "Can't write response: %v", wErr)
		http.Error(w, fmt.Sprintf("cloud not write response: %v", wErr), http.StatusInternalServerError)
	}

}