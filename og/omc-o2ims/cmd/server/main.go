package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"os"
	"sync"

	res "github.com/enrayga/omc-o2ims/internal/operator/resource"
	"github.com/enrayga/omc-o2ims/internal/operator/watcher"

	"github.com/enrayga/omc-o2ims/internal/config"
	"github.com/enrayga/omc-o2ims/internal/operator/store"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {

}

var (
	InfoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarnLogger  = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func getInClusterConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
	}
	return config, nil
}

func main() {

	InfoLogger.Printf("Starting O-Cloud server\n")
	InfoLogger.Printf("docker run -d --name omc-o2ims  --network=host   -v /path/to/host/config:/app/config \n ")
	//log.Printf("/path to host should containe  config.yaml , kubeconfig.yaml and client.crts \n ")

	configFile := flag.String("config", "config/config.yaml", "Path to configuration file")

	flag.Parse()
	_ = configFile
	InfoLogger.Print("config file found at ", *configFile, "\n")

	// // Load configuration
	// // if no argument is provided, default paths are used
	config, err := config.LoadConfig(*configFile)
	if err != nil {
		ErrorLogger.Println("Failed to load configuration:", err)
		os.Exit(1)
	}

	dumpConfig, _ := json.MarshalIndent(config, "", "    ")
	dumpConfig = bytes.Replace(dumpConfig, []byte(config.Omc.Password), []byte("******"), -1)
	InfoLogger.Println(string(dumpConfig))

	// Export config to environment variables
	os.Setenv("OMC_BACKEND", config.BackendType)
	os.Setenv("OMC_BACKEND_URL", config.Omc.URL)
	os.Setenv("OMC_BACKEND_USERNAME", config.Omc.Username)
	os.Setenv("OMC_BACKEND_PASSWORD", config.Omc.Password)

	InfoLogger.Println("OMC_BACKEND:", config.BackendType)
	InfoLogger.Println("OMC_BACKEND_URL:", config.Omc.URL)
	InfoLogger.Println("OMC_BACKEND_USERNAME:", config.Omc.Username)

	var kubeconfig string
	var kube *rest.Config

	kubeconfig = config.Kubernetes.KubeConfig
	namespace := config.Kubernetes.Namespace
	crdFile := config.CRD.Files[0]

	crdDefinitions, err := os.ReadFile(crdFile)
	if err != nil {
		ErrorLogger.Println("failed to load CRD definition:", err)
		os.Exit(1)
	}
	crdDefinition := string(crdDefinitions)

	//_ = crdDefinition

	//_ = namespace
	//_ = crdFile
	//_ = kubeconfig

	kube, err = getInClusterConfig()
	if err != nil {
		fmt.Printf("Error getting in-cluster config: %v\n", err)
		kube, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			ErrorLogger.Println("failed to load cluster  config:", err)
			os.Exit(1)
			return
		}
	}

	ctx := context.Background()
	dynamicClient, err := dynamic.NewForConfig(kube)

	if err != nil {
		ErrorLogger.Printf("failed to create dynamic client: %v", err)
	} else {
		InfoLogger.Println("Successfully created dynamic client")
	}

	apiextensionsClient, err := apiextensionsclient.NewForConfig(kube)
	if err != nil {
		ErrorLogger.Printf("failed to create apiextensions client: %v", err)
	} else {
		InfoLogger.Println("Successfully created apiextensions client")
	}

	var lstore *store.K8sStore[*res.ProvisioningRequest]

	if config.DataStore == "" || config.DataStore == "k8s" {
		k8sStore, err := store.NewK8sStore[*res.ProvisioningRequest](
			apiextensionsClient,
			dynamicClient,
			ctx, crdDefinition, namespace)
		if err != nil {
			ErrorLogger.Printf("failed to create k8sStore: %v\n", err)
			os.Exit(1)
		}
		lstore = k8sStore
	}

	if lstore == nil {
		ErrorLogger.Println("failed to initialize data store. Exiting.")
		os.Exit(1)
	}

	watcher := watcher.NewWatcherImpl[*res.ProvisioningRequest](lstore)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := watcher.StartWatching(ctx); err != nil {
			ErrorLogger.Printf("failed to start watching: %v", err)
			os.Exit(1)
		}
	}()
	wg.Wait()
	select {} // wait forever

}
