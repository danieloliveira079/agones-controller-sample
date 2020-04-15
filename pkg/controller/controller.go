package controller

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

func NewAgonesController(logger *logrus.Entry, clientSet versioned.Interface) (*Controller, error) {
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

func (c *Controller) Run(stop <-chan struct{}) {
	c.gameServersInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
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

		},
	})

	go c.informerFactory.Start(stop)

	<-stop
	c.logger.Info("Stopping Agones Controller")
}

func (c *Controller) EventHandlerGameServerAdd(obj interface{}) error {
	var key string
	var err error

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		return err
	}

	if _, ok := obj.(*agonesv1.GameServer); !ok {
		return fmt.Errorf("object is not of type %T", &agonesv1.GameServer{})
	}

	gameServer := obj.(*agonesv1.GameServer)

	// Implement your business logic here.
	// I.e: Send a http request to the external world, modify the gameserver status or labels, etc
	c.logger.Debugf("Handled Add GameServer Event: %s - State: %s", key, gameServer.Status.State)

	return nil
}

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
		c.logger.Debugf("Handled Update GameServer Event: %s - version %s to %s", oldKey, oldGameServer.ResourceVersion, newGameServer.ResourceVersion)
		return nil
	}

	c.logger.Debugf("Handled Update GameServer Event: %s - nothing changed", newKey)

	return nil
}

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
