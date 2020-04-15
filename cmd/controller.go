package main

import (
	agonesv1 "agones.dev/agones/pkg/client/clientset/versioned"
	"flag"
	"fmt"
	"github.com/danieloliveira079/agones-controller-sample/internal/version"
	"github.com/danieloliveira079/agones-controller-sample/pkg/controllers"
	"github.com/danieloliveira079/agones-controller-sample/pkg/log"
	"github.com/danieloliveira079/agones-controller-sample/pkg/signals"
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

	logger := log.NewLoggerWithLevel(loglevel)

	logger.Debug("Starting GameServer Controller")
	logger.Debugf(fmt.Sprintf("%#v", (version.Info())))

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
	gameServerController, err := controllers.NewGameServerController(logger, agonesClientSet)
	if err != nil {
		logger.Fatal(err)
	}

	stop := signals.SetupSignalHandler()
	gameServerController.Run(stop)

	logger.Info("GameServer Controller Terminated")
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster")
	flag.BoolVar(&verbose, "verbose", false, "Set loglevel to verbose. Use for debugging purpose")
}
