package main

import (
	agonesv1 "agones.dev/agones/pkg/client/clientset/versioned"
	"flag"
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

	logger.Debug("Starting Agones Controller")

	clientConf, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logger.Fatal(err)
	}

	agonesClientSet, err := agonesv1.NewForConfig(clientConf)
	if err != nil {
		logger.Fatal(err)
	}

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
