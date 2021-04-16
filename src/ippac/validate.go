package main

import (
	"net/http"
	"os"
	"io/ioutil"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"encoding/json"
	"fmt"

)

func (gs *myServerHandler) valserve(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/validate" {
		fmt.Fprintf(os.Stderr, "Not a Valid URL Path\n")
		http.Error(w, "Not A Valid URL Path", http.StatusBadRequest)
	}

	var Body []byte
	if r.Body != nil {
		if data , err := ioutil.ReadAll(r.Body); err == nil {
			Body = data
		} else {
			fmt.Fprintf(os.Stderr, "Unable to Copy the Body\n")
		}
	}

	if len(Body) == 0 {
		fmt.Fprintf(os.Stderr, "Unable to retrieve Body from the WebHook\n")
		http.Error(w, "Unable to retrieve Body from the API" , http.StatusBadRequest )
		return
	} else {
		fmt.Fprintf(os.Stdout, "Body retrieved\n")
	}

	arRequest := &admissionv1.AdmissionReview{}

	if err := json.Unmarshal(Body, arRequest); err != nil {
		fmt.Fprintf(os.Stderr, "unable to Unmarshal the Body Request\n")
		http.Error(w, "unable to Unmarshal the Body Request" , http.StatusBadRequest)
		return
	}
	
	// Making Sure we are not running on a system Namespace
	if isKubeNamespace(arRequest.Request.Namespace) {
		fmt.Fprintf(os.Stderr, "Unauthorized Namespace\n")
		http.Error(w, "Unauthorized Namespace", http.StatusBadRequest)
	}

	// initial the POD values from the request

	row := arRequest.Request.Object.Raw
	pod := corev1.Pod{}

	if err := json.Unmarshal(row, &pod); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to Unmarshal the Pod Information\n")
		http.Error(w, "Unable to Unmarshal the Pod Information", http.StatusBadRequest)
	}


	// Now we Are going to Start and build the Response 
	arResponse := admissionv1.AdmissionReview {
		Response: &admissionv1.AdmissionResponse{
			Result: &metav1.Status{Status: "Failure", Message: "Not All Images are set to \"Always pull image\" policy", Code: 401},
			UID: arRequest.Request.UID,
			Allowed: false,
		},
	}

	// Let's take an array of all the containers 
	containers := pod.Spec.Containers

	var pullPolicyFlag bool
	pullPolicyFlag = true

	for i , container := range containers {
		fmt.Fprintf(os.Stdout, "container[%d]= %s imagePullPolicy=%s", i, container.Name , container.ImagePullPolicy)
		if container.Name != "recycler-container" && containers[i].ImagePullPolicy != "Always" {
			pullPolicyFlag = false
		}
	}

	arResponse.APIVersion = "admission.k8s.io/v1"
	arResponse.Kind = arRequest.Kind
	

	if pullPolicyFlag == true {
		arResponse.Response.Allowed = true
		arResponse.Response.Result = &metav1.Status{Status: "Success", 
			Message: "All Images are set to \"Always pull image\" policy", 
			Code: 201}
	}

	resp , resp_err := json.Marshal(arResponse)

	if resp_err != nil {
		fmt.Fprintf(os.Stderr, "Unable to Marshal the Request\n")
		http.Error(w, "Unable to Marshal the Request", http.StatusBadRequest)
	}

	if _ , werr := w.Write(resp); werr != nil {
		fmt.Fprintf(os.Stderr, "Unable to Write the Response\n")
		http.Error(w, "Unable to Write Response", http.StatusBadRequest)
	}
}