package main

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"log"
	"net/http"
	"os"
)

type server struct{}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "hello world"}`))
}

func main() {
	s := &server{}
	http.Handle("/", s)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Some help here https://stackoverflow.com/questions/45692719/samples-on-kubernetes-helm-golang-client
func installChart() {
	chartPath := "/some/path"
	namespace := "default"
	releaseName := "my-release"

	// Test/example values (TODO: from request)
	values := map[string]interface{}{
		"example": map[string]interface{}{
			"addr": "localhost",
			"port": "26379",
		},
	}

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

	iCli := action.NewInstall(actionConf)
	iCli.Namespace = namespace
	iCli.ReleaseName = releaseName
	// client.DryRun = true // - for testing

	rel, err := iCli.Run(chart, values)
	if err != nil {
		panic(err)
	}
	log.Printf("Installed Chart from path: %s in namespace: %s\n", rel.Name, rel.Namespace)
}
