package controllers

import (
	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	"agones.dev/agones/pkg/client/clientset/versioned"
	"agones.dev/agones/pkg/client/informers/externalversions"
	informersv1 "agones.dev/agones/pkg/client/informers/externalversions/agones/v1"
	listersv1 "agones.dev/agones/pkg/client/listers/agones/v1"
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/cache"
	"reflect"
	"time"
)

type Controller struct {
	logger              *logrus.Entry
	informerFactory     externalversions.SharedInformerFactory
	gameServersInformer informersv1.GameServerInformer
	gameServersLister   listersv1.GameServerLister
}

// NewGameServerController returns a new Controller with informer and lister set. It requires a not nil Agones clientset.
func NewGameServerController(logger *logrus.Entry, clientSet versioned.Interface) (*Controller, error) {
	if clientSet == nil {
		logger.Fatal("controller can't be created with a nil clientSet")
	}

	// Create a new SharedInformerFactory with a re-sync period of 15 seconds.
	agonesInformerFactory := externalversions.NewSharedInformerFactory(clientSet, time.Second*15)

	// Same approach can be used for other types of informers like: GameServerSets and Fleets
	gameServersInformer := agonesInformerFactory.Agones().V1().GameServers()

	controller := &Controller{
		logger:              logger,
		informerFactory:     agonesInformerFactory,
		gameServersInformer: gameServersInformer,
		gameServersLister:   gameServersInformer.Lister(),
	}

	return controller, nil
}

// Run starts the GameServer controller and attach event handlers which will receive notifications from the informer.
func (c *Controller) Run(stop <-chan struct{}) {
	c.gameServersInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// Alternatively, you could set any method that contains the right signature `func(obj interface{})`
		// AddFunc: c.EventHandlerAdd,
		AddFunc: func(obj interface{}) {
			if err := c.EventHandlerGameServerAdd(obj); err != nil {
				c.logger.WithError(err).Error("add event error")
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if err := c.EventHandlerGameServerUpdate(oldObj, newObj); err != nil {
				c.logger.WithError(err).Error("update event error")
			}
		},
		DeleteFunc: func(obj interface{}) {
			if err := c.EventHandlerGameServerDelete(obj); err != nil {
				c.logger.WithError(err).Error("delete event error")
			}
		},
	})

	go c.informerFactory.Start(stop)

	<-stop
	c.logger.Info("Stopping GameServer Controller")
}

// EventHandlerGameServerAdd handles events triggered in two occasions:
// 1. When the controller is first run and has to sync with the cache
// 2. When a new resource is added via Kubernetes API requests: kubectl, client-go, http requests, etc.
func (c *Controller) EventHandlerGameServerAdd(obj interface{}) error {
	key, addedGameServer, err := IsGameServerKind(obj)
	if err != nil {
		return err
	}

	// Implement your business logic here.
	// I.e: Send a http request to the external world, modify the GameServer status or labels or even
	// communicate with your GameServer backend

	// This is just an example of how to check general changes. Generally, checks will look for differences within the
	// resource status
	c.logger.Debugf("Handled Add GameServer Event: %s - State: %s", key, addedGameServer.Status.State)

	return nil
}

// EventHandlerGameServerUpdate handles events triggered due to a resource being updated.
// That includes chances caused by either the Kubernetes controller manager or any other external actor modifying the
// resource. I.e.: Another GameServer controller.
func (c *Controller) EventHandlerGameServerUpdate(oldObj, newObj interface{}) error {
	oldKey, oldGameServer, err := IsGameServerKind(oldObj)
	if err != nil {
		return err
	}

	newKey, newGameServer, err := IsGameServerKind(newObj)
	if err != nil {
		return err
	}

	// Implement your business logic here.
	// I.e: Send a http request to the external world, modify the GameServer status or labels or even
	// communicate with your GameServer backend

	// This is just an example of how to check general changes. Generally, checks will look for differences within the
	// resource status
	if reflect.DeepEqual(oldGameServer, newGameServer) == false {
		c.logger.Debugf("Handled Update GameServer Event: %s (%s) - version %s to %s", oldKey, newGameServer.Status.State, oldGameServer.ResourceVersion, newGameServer.ResourceVersion)

		// Both properties from the old and the new GameServer can be accessed. Not only Status.
		if newGameServer.Status.State == agonesv1.GameServerStateReady && newGameServer.DeletionTimestamp.IsZero() {
			c.logger.Infof("GameServer Ready %s - %s:%d", newKey, newGameServer.Status.Address, newGameServer.Status.Ports[0].Port)
		}

		return nil
	}

	c.logger.Debugf("Handled Update GameServer Event: %s - nothing changed", newKey)

	return nil
}

// EventHandlerGameServerDelete handles events triggered due to a resource being deleted.
func (c *Controller) EventHandlerGameServerDelete(obj interface{}) error {
	key, deletedGameServer, err := IsGameServerKind(obj)
	if err != nil {
		return err
	}

	// Implement your business logic here.
	// I.e: Send a http request to the external world, modify the GameServer status or labels or even
	// communicate with your GameServer backend

	// This is just an example of how to check general changes. Generally, checks will look for differences within the
	// resource status
	c.logger.Debugf("Handled Delete GameServer Event: %s - %s", key, deletedGameServer.DeletionTimestamp.String())

	return nil
}

// IsGameServerKind checks if the passed object is of type GameServer and returns the resource key (namespace/name),
// the GameServer reference. An error is returned if neither a key can be extracted from the object nor the object can
// be casted to a GameServer type.
func IsGameServerKind(obj interface{}) (string, *agonesv1.GameServer, error) {
	var key string
	var err error

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		return key, nil, err
	}

	if _, ok := obj.(*agonesv1.GameServer); !ok {
		return key, nil, fmt.Errorf("object is not of type %T", &agonesv1.GameServer{})
	}

	return key, obj.(*agonesv1.GameServer), nil
}
