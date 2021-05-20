package main

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"log"
	"net/http"
	"os"
)

var (
	chartPath string
	namespace string
)

func main() {
	log.Println("Starting server...")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
	chartPath = os.Getenv("CHART_PATH")
	namespace = os.Getenv("KUBE_NAMESPACE")
	log.Printf("chartPath: %s", chartPath)
	log.Printf("namespace: %s", namespace)
	http.HandleFunc("/start-chart", startChart)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type RequestBody struct {
	releaseName string
	values      map[string]interface{}
}

func startChart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024) // max 2 MB
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		var body RequestBody
		err := dec.Decode(&body)
		if err != nil {
			msg := "Failed to parse request body"
			log.Printf("Request failed: %s (%s)", err.Error(), msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		upgradeOrInstallChart(body.releaseName, body.values)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}
}

// Some help here https://stackoverflow.com/questions/45692719/samples-on-kubernetes-helm-golang-client
func upgradeOrInstallChart(releaseName string, values map[string]interface{}) {
	// Load the chart
	chart, err := loader.Load(chartPath)
	if err != nil {
		panic(err)
	}
	actionConf := new(action.Configuration)
	clientGetter := cli.New().RESTClientGetter()
	if err := actionConf.Init(clientGetter, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Printf("%+v", err)
	}

	client := action.NewUpgrade(actionConf)
	client.Namespace = namespace
	// client.DryRun = true // - for testing

	rel, err := client.Run(releaseName, chart, values)
	if err != nil {
		panic(err)
	}
	log.Printf("Installed/Upgraded Chart '%s' from path: '%s' in namespace: '%s'\n", releaseName, rel.Name, rel.Namespace)
}
