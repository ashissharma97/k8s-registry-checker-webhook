package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	v1 "k8s.io/api/admission/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	universalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
	port                  = LookupEnvOrString("PORT", "8443")
	allowRegistryName     = LookupEnvOrString("ALLOW_REGISTRY_NAME", "gcr")
	tlsCertFile           = LookupEnvOrString("TLS_CRT_PATH", "/etc/webhook/certs/tls.crt")
	tlsKeyFile            = LookupEnvOrString("TLS_KEY_PATH", "/etc/webhook/certs/tls.key")
)

func main() {

	http.HandleFunc("/validate", Validate)
	log.Fatal(http.ListenAndServeTLS(":"+port, tlsCertFile, tlsKeyFile, nil))
}

func Validate(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return
	}

	var admissionReviewReq v1.AdmissionReview

	if _, _, err := universalDeserializer.Decode(body, nil, &admissionReviewReq); err != nil {
		http.Error(w, fmt.Sprintf("could not deserialize request: %v", err), http.StatusBadRequest)
		return
	} else if admissionReviewReq.Request == nil {
		http.Error(w, "malformed admission review: request is nil", http.StatusBadRequest)
		return
	}

	var pod apiv1.Pod

	err = json.Unmarshal(admissionReviewReq.Request.Object.Raw, &pod)

	if err != nil {
		http.Error(w, fmt.Sprintf("could not unmarshal pod on admission request: %v", err), http.StatusBadRequest)
		return
	}

	var allow bool = true

	for _, pod := range pod.Spec.Containers {

		if !strings.HasPrefix(pod.Image, allowRegistryName) {
			allow = false
			break
		}

	}

	if allow {

		response := v1.AdmissionReview{
			TypeMeta: admissionReviewReq.TypeMeta,
			Response: &v1.AdmissionResponse{
				UID:     admissionReviewReq.Request.UID,
				Allowed: true,
			},
		}

		bytes, err := json.Marshal(response)

		if err != nil {
			fmt.Println(err)
			http.Error(w, fmt.Sprintf("could not marshal response: %v", err), http.StatusInternalServerError)
			return

		}

		w.Header().Set("Content-Type", "application/json")

		w.Write(bytes)

	} else {

		response := v1.AdmissionReview{
			TypeMeta: admissionReviewReq.TypeMeta,
			Response: &v1.AdmissionResponse{
				UID:     admissionReviewReq.Request.UID,
				Allowed: false,
				Result: &metav1.Status{
					Message: "One of the container image name doesn't start with allowed registry",
				},
			},
		}

		bytes, err := json.Marshal(response)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not marshal response: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		w.Write(bytes)
	}

}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
