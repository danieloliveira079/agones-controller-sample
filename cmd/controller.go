package main

import (
	agonesv1 "agones.dev/agones/pkg/client/clientset/versioned"
	"flag"
	"github.com/danieloliveira079/howto-agones-informers/pkg/controller"
	"github.com/danieloliveira079/howto-agones-informers/pkg/signals"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig string
	verbose    bool
)

func main() {
	flag.Parse()

	loglevel := logrus.InfoLevel
	if verbose {
		loglevel = logrus.DebugLevel
	}

	logger := NewLoggerWithLevel(loglevel)

	logger.Debug("Starting GameServer Controller")

	// The account from the Kubeconfig must have the right RBAC configurations.
	// TODO provide examples of RBAC or point to Agones Docs
	clientConf, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logger.WithError(err).Fatal("error building kubeconfig")
	}

	// Create new AgonesClientSet using the previously created clientConf
	agonesClientSet, err := agonesv1.NewForConfig(clientConf)
	if err != nil {
		logger.Fatal(err)
	}

	// Instantiates the new agones controller passing the logger and the clientSet to be used by the informer
	agonesController, err := controller.NewAgonesController(logger, agonesClientSet)
	if err != nil {
		logger.Fatal(err)
	}

	stop := signals.SetupSignalHandler()

	agonesController.Run(stop)

	logger.Info("GameServer Controller Terminated")
}

func NewLoggerWithLevel(level logrus.Level) *logrus.Entry {
	log := logrus.New()
	if verbose {
		log.SetLevel(level)
	}
	return logrus.NewEntry(log)
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster")
	flag.BoolVar(&verbose, "verbose", false, "Set loglevel to verbose. Use for debugging purpose")
}
