package main

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
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
	port := ":8080"
	log.Printf("Starting the server on port %v", port)
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %+v", err)
	}
	chartPath = os.Getenv("CHART_PATH")
	namespace = os.Getenv("KUBE_NAMESPACE")
	authName := os.Getenv("AUTH_NAME")
	authPass := os.Getenv("AUTH_PASSWORD")
	realm := "Please login first"
	log.Printf("chartPath: %s", chartPath)
	log.Printf("namespace: %s", namespace)
	http.HandleFunc("/start-chart", BasicAuth(startChart, authName, authPass, realm))
	http.HandleFunc("/uninstall-chart", BasicAuth(stopChart, authName, authPass, realm))
	log.Fatal(http.ListenAndServe(port, logRequest(http.DefaultServeMux)))
}

type StartChartRequest struct {
	ReleaseName string                 `json:"releaseName"`
	Values      map[string]interface{} `json:"values"`
}

func startChart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024) // max 2 MB
		dec := json.NewDecoder(r.Body)
		// dec.DisallowUnknownFields()
		var body StartChartRequest
		err := dec.Decode(&body)
		if err != nil {
			msg := "Failed to parse request body"
			log.Printf("Request failed: %s (%s)", err.Error(), msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		upgradeOrInstallChart(body.ReleaseName, body.Values)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}
}

// Some help here https://stackoverflow.com/questions/45692719/samples-on-kubernetes-helm-golang-client
func upgradeOrInstallChart(releaseName string, values map[string]interface{}) {
	log.Printf("New release: %s with values %v", releaseName, values)
	chart, err := loader.Load(chartPath)
	if err != nil {
		log.Fatalf("Failed to load chart %v", err)
	}
	actionConf := new(action.Configuration)
	clientGetter := cli.New().RESTClientGetter()
	if err := actionConf.Init(clientGetter, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Printf("%+v", err)
	}

	var rel *release.Release

	// Install if does not exist
	histClient := action.NewHistory(actionConf)
	if _, err := histClient.Run(releaseName); err == driver.ErrReleaseNotFound {
		log.Printf("%s does not exist. Installing it now...", releaseName)
		instClient := action.NewInstall(actionConf)
		instClient.ReleaseName = releaseName
		instClient.Namespace = namespace
		instRel, err := instClient.Run(chart, values)
		rel = instRel
		if err != nil {
			log.Fatalf("Failed to install chart %v", err)
		}
	} else {
		upgrClient := action.NewUpgrade(actionConf)
		upgrClient.Namespace = namespace
		// upgrClient.DryRun = true // - for testing

		upgrRel, err := upgrClient.Run(releaseName, chart, values)
		rel = upgrRel
		if err != nil {
			log.Fatalf("Failed to upgrade chart %v", err)
		}
	}
	log.Printf("Installed/Upgraded Chart '%s' from path: '%s' in namespace: '%s'\n", releaseName, rel.Name, rel.Namespace)
}

func stopChart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		releaseName := r.URL.Query().Get("release")
		uninstallChart(releaseName)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}
}

func uninstallChart(releaseName string) {
	actionConf := new(action.Configuration)
	clientGetter := cli.New().RESTClientGetter()
	if err := actionConf.Init(clientGetter, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Printf("%+v", err)
	}
	client := action.NewUninstall(actionConf)
	rel, err := client.Run(releaseName)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Uninstalled Chart '%s' from path: '%s' in namespace: '%s'\n", releaseName, rel.Release.Name, rel.Release.Namespace)
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
