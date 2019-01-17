package main

import (
	"flag"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	clientset "github.com/liyanyanli/k8s-controller-custom-resource/pkg/client/clientset/versioned"
	informers "github.com/liyanyanli/k8s-controller-custom-resource/pkg/client/informers/externalversions"
	"github.com/liyanyanli/k8s-controller-custom-resource/pkg/signals"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	democrdClient, err := clientset.NewForConfig(cfg)
	if err != nil {

		klog.Fatalf("Error building example clientset: %s", err.Error())
	}

	democrdInformerFactory := informers.NewSharedInformerFactory(democrdClient, time.Second*30)

	controller := NewController(kubeClient, democrdClient,
		democrdInformerFactory.Democrd().V1().Democrds())

	go democrdInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
