package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/rs/zerolog/log"
	admission "k8s.io/api/admission/v1"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecFactory  = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecFactory.UniversalDeserializer()
)

// adding kind admissionReview in scheme
func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admission.AddToScheme(runtimeScheme)
	_ = v1.AddToScheme(runtimeScheme)
}

type admitv1Func func(admission.AdmissionReview) *admission.AdmissionResponse

type admitHandler struct {
	v1 admitv1Func
}

func AdmitHandler(f admitv1Func) admitHandler {
	return admitHandler{
		v1: f,
	}
}

// http handler
func serve(w http.ResponseWriter, r *http.Request, admit admitHandler) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	//verify if the content-type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Error().Msgf("contentType=%s, expect application/json", contentType)
		return
	}

	log.Info().Msgf("handling request: %s", body)
	var responseObj runtime.Object
	if obj, gvk, err := deserializer.Decode(body, nil, nil); err != nil {
		msg := fmt.Sprintf("Request could not be decoded: %v", err)
		log.Error().Msg(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	} else {
		requestedAdmissionReview, ok := obj.(*admission.AdmissionReview)
		if !ok {
			log.Error().Msgf("expected v1.admissionreview but got : %T", obj)
			return
		}
		responseAdmissionReview := &admission.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.v1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview

	}

	log.Info().Msgf("sending response: %v", responseObj)
	respBytes, err := json.Marshal(responseObj)
	if err != nil {
		log.Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "applicaion/json")
	if _, err := w.Write(respBytes); err != nil {
		log.Err(err)
	}
}

func serveValidate(w http.ResponseWriter, r *http.Request) {
	serve(w, r, AdmitHandler(validate))
}

func validate(ar admission.AdmissionReview) *admission.AdmissionResponse {
	log.Info().Msgf("Validating pods")

	podResource := metav1.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	if ar.Request.Resource != podResource {
		log.Error().Msgf("expect resource to be %s", &podResource)
		return nil
	}
	raw := ar.Request.Object.Raw
	pod := corev1.Pod{}

	if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
		log.Err(err)
		return &admission.AdmissionResponse{
			Result: &metav1.Status{
				Message: "got error while decoding and deserializing the pods",
			},
		}
	}

	// labelsmap := map[string]string{
	// 	"app": "nginx",
	// }

	//creating a client set
	// kubeconfig := flag.String("kubeconfig", "/home/kartik/.kube/config", "This is where the kube config file resides")

	// config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	fmt.Printf("Error building config from kubeconfig: %s", err.Error())

	// 	//building from internal cluser
	// 	config, err = rest.InClusterConfig()
	// }

	// //lets build the kubeclient set now
	// clientSet, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	fmt.Printf("Error while creating the kube client set: %s", err.Error())
	// }
	// fmt.Println(clientSet)

	// var mapOfLabelsInNetPol = make(map[string]string)
	// //list all the network policies as well present in the namespace
	// netpolList, err := clientSet.NetworkingV1().NetworkPolicies("production").List(context.Background(), metav1.ListOptions{})
	// if err != nil {
	// 	fmt.Printf("Error while getting the netpolList: %s", err.Error())
	// }
	// for _, netpol := range netpolList.Items {
	// 	fmt.Println(netpol.Name)
	// 	mapOfLabelsInNetPol = netpol.Spec.PodSelector.MatchLabels

	// }

	actual_map := pod.GetLabels()
	// res_map := reflect.DeepEqual(mapOfLabelsInNetPol, actual_map)

	// //

	// b := new(bytes.Buffer)
	// for key, value := range mapOfLabelsInNetPol {
	// 	fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	// }

	// c := new(bytes.Buffer)
	// for key, value := range actual_map {
	// 	fmt.Fprintf(c, "%s=\"%s\"\n", key, value)
	// }
	mapToBeCompared := map[string]string{
		"app": "nginx",
	}
	res_map := reflect.DeepEqual(mapToBeCompared, actual_map)
	//
	if res_map == false {
		return &admission.AdmissionResponse{
			Allowed: false, Result: &metav1.Status{
				Message: "this label is being used in the netpol, so can not modify it",
			},
		}
	}

	return &admission.AdmissionResponse{
		Allowed: true,
	}
}

func main() {
	var tlsKey, tlsCert string
	flag.StringVar(&tlsKey, "tlsKey", "/etc/certs/tls.key", "Path to the tls key")
	flag.StringVar(&tlsCert, "tlsCert", "/etc/certs/tls.crt", "Path to the TLS certificate")
	flag.Parse()
	http.HandleFunc("/validate", serveValidate)
	log.Info().Msg("Server started ...")
	log.Fatal().Err(http.ListenAndServeTLS(":8443", tlsCert, tlsKey, nil)).Msg("webhook server exited")
}
