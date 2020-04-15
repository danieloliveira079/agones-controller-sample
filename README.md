## Agones Controller Sample

This repository implements a simple controller for watching GameServers resources which have been deployed using https://agones.dev.

The project has an educational purpose and its intent is purely to demonstrate and briefly explain some of the components which are required when building Kubernetes controllers.
 
What differentiates this project from others found on the internet is the ability to interact with non-native Kubernetes resources. In this case those resources are [Agones GameServers](https://agones.dev/site/docs/reference/gameserver).

### What is Agones?
[Agones](https://github.com/googleforgames/agones) is a library for hosting, running and scaling [dedicated game](https://en.wikipedia.org/wiki/Game_server#Dedicated_server) servers on [Kubernetes](https://kubernetes.io/).

The Agones project is open source and can be found on https://github.com/googleforgames/agones.

Additionally, there is a ton of documentation on the Agones blog https://agones.dev/site and a Slack community.

### What is a Kubernetes controller?
> In Kubernetes, controllers are control loops that watch the state of your cluster, then make or request changes where needed. Each controller tries to move the current cluster state closer to the desired state.
>
> -- <cite>[kubernetes.io](https://kubernetes.io/docs/concepts/architecture/controller)</cite>

There is a vast source of material if you are interested on the topic. Some are listed below:
- https://github.com/kubernetes/sample-controller
- https://book.kubebuilder.io/
- https://github.com/operator-framework/operator-sdk
- [Programming Kubernetes Book](https://www.amazon.com/Programming-Kubernetes-Developing-Cloud-Native-Applications-dp-1492047104/dp/1492047104/ref=mt_paperback?_encoding=UTF8&me=&qid=1586961333)

## GameServer Controller

Requirements:
- A Kubernetes v1.14.x cluster running Agones. Instructions can be found on https://agones.dev/site/docs/installation/creating-cluster/
- If you are running the GameServer controller out of the cluster, make sure you are passing a valid `--kubeconfig` path as argument. Usually this file can be found at `~/.kube/config`. 
- Go 1.14+ (possible compatible with previus versions, not tested though)

Limitations
- Not built or tested on Windows machines 
 
### Controller Core Components

**Kubernetes:**

- **ClientConfig**: Holds the configs parsed from the kubeconfig file. Used when creating Kubernetes clientsets.
- **ClientSet**: Gives access to clients for multiple API groups and resources. The GameServer Controller uses it to access Agones GameServers resources. 
- **SharedInformerFactory**: Allows informers to be shared for the same resource in an application. 
- **Informer**: In memory caching that can react to changes of objects in nearly real-time. 
- **Lister**: Perform Create, Get, Update and Delete operations for an specific type of resource. 
- **Workqueue** [optional]: This is a data structure that implements a priority queue.

Details about all these componentes can be found on https://github.com/kubernetes/client-go

**Controller:**

- EventHandlers: Methods that will be called by the informer when a notification happens. Possible events are: Add, Update and Delete. These are the places where the business logic of your controller can be implemented. 

### Project Structure

Below you can find some highlights of the GameServer controller code base which are crucial for a good understanding.   

`cmd/controller.go`: Initiates the application and sets the config, the agones clientset and creates the GameServer controller.

`pkg/controllers/gameserver.go`: All the GameServer controller logic and required objects. That includes event handlers, informer factory, informers and lister.

Detailed description of the most important blocks of code can be found below:

- cmd/controller.go
    - Create the client config based on the `--kubeconfig` flag
        ```go   
        // kubeconfig must be a path to a valid Kubeconfig file. 
        // I.e: /Users/foo/.kube/config
        // The master URL argument can be omitted.
        clientConf, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
        ```
    - Create new AgonesClientSet using the previously created clientConf
        ```go
        //Make sure you have imported the required Agones packages
        import (
            agonesv1 "agones.dev/agones/pkg/client/clientset/versioned"
            ...
        )
      
        // clientConf, logger, ...
      
        agonesClientSet, err := agonesv1.NewForConfig(clientConf)
        ```
- pkg/controller/gameserver.go
    - Create the new SharedInformerFactory
        ```go
        // Make sure you have imported the required Agones packages
        import (
            "agones.dev/agones/pkg/client/informers/externalversions"
              ...
        )
        // Create a new SharedInformerFactory with a re-sync period of 15 seconds.
        agonesInformerFactory := externalversions.NewSharedInformerFactory(clientSet, time.Second*15)
        ```
    - Get the GameServer informer from the SharedInformerFactory
        ```go
        // Same approach can be used for other types of informers like: GameServerSets and Fleets
        gameServersInformer := agonesInformerFactory.Agones().V1().GameServers()
        ```
    - Get the GameServer lister from the SharedInformer
        ```go
        controller := &Controller{
                logger:              logger,
                informerFactory:     agonesInformerFactory,
                gameServersInformer: gameServersInformer,
                // the lister is used for Create, Update and Delete operations
                gameServersLister:   gameServersInformer.Lister(),
            }
        ```
    - Add EventHandlers and Start the informer
    ```go
    // Alternatively, you could set any method that contains the right signature for the event
    // I.e.: AddFunc: c.EventHandlerAdd,
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
			if err := c.EventHandlerGameServerDelete(obj); err != nil {
				c.logger.WithError(err).Error("delete event error")
			}
		},
	})

    // start the informer to receive events notifications 	 
    c.informerFactory.Start(wait.NeverStop)
    ```
    - Example using Update event handler
    ```go
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
    ```

### How to build and run this project?

You can use the `Makefile` that provides:
- `make build`: build the controller targeting Linux platform and output the binary to `bin/agones-controller`
- `make test`: run all the project's tests
- `make dist`: build the controller for multiple platforms, including: Linux and Darwin. Binaries will be output to the `bin/` folder

Feel free to explore other options available on the `Makefile`.

Using the output binary:
```bash
$ ./bin/agones-controller-sample --verbose --kubeconfig /Users/foo/.kube/config
```

Using go tools:
```bash
$ go run cmd/controller.go --verbose --kubeconfig /Users/foo/.kube/config
```

The expected output should be something similar to:
```bash
DEBU[0000] Starting GameServer Controller               
DEBU[0000] version.BuildInfo{Version:"0f73df9", BuildTS:"2020-04-15 06:10:41", GitBranch:"master", GitCommit:"0f73df96c1fa7fb5c4bf4ca14a50ad9850cf43c3", GoVersion:"go1.14.1"} 
```

### Triggering Events

Using a second shell session, deploy the `simple-udp-agones.yaml` running:
```bash
$ kubectl apply -f examples/simple-udp-agones.yaml
```

The expected output should be something similar to:
```bash
 DEBU[0041] Handled Add GameServer Event: default/simple-udp-agones - State: PortAllocation 
 DEBU[0041] Handled Update GameServer Event: default/simple-udp-agones (Creating) - version 334893 to 334896 
 DEBU[0041] Handled Update GameServer Event: default/simple-udp-agones (Starting) - version 334896 to 334899 
 DEBU[0041] Handled Update GameServer Event: default/simple-udp-agones (Scheduled) - version 334899 to 334904 
 DEBU[0045] Handled Update GameServer Event: default/simple-udp-agones - nothing changed 
 DEBU[0050] Handled Update GameServer Event: default/simple-udp-agones (RequestReady) - version 334904 to 334922 
 DEBU[0050] Handled Update GameServer Event: default/simple-udp-agones (Ready) - version 334922 to 334924 
 INFO[0050] GameServer Ready default/simple-udp-agones - 172.17.0.2:7300 
```

**Update the GameServer label:**
```bash
$ kubectl label --overwrite gameservers.agones.dev simple-udp-agones app=simple-udp-server-v2
```

The expected output should be something similar to:
```bash
DEBU[0569] Handled Update GameServer Event: default/simple-udp-agones (Ready) - version 335572 to 335592 
INFO[0569] GameServer Ready default/simple-udp-agones - 172.17.0.2:7104 
```

**Delete the GameServer running:**
```bash
$ kubectl delete -f examples/simple-udp-agones.yaml
```

The expected output should be something similar to:
```bash
DEBU[0112] Handled Update GameServer Event: default/simple-udp-agones (Ready) - version 334924 to 335004 
DEBU[0116] Handled Delete GameServer Event: default/simple-udp-agones - 2020-04-15 20:14:31 +0200 CEST
```

### TODO
- [ ] Add CI
- [ ] Add Dockerfile
- [ ] Push to a Docker Hub repo
- [ ] Add Deployment manifests
- [ ] Add RBAC example required to run the GameServer controller
- [ ] More test coverage
- [ ] Use fake client for testing
- [ ] Add an example using workqueue and lister
- [ ] Upgrade to Agones 1.5
- [ ] Add examples using filtered informers
- [ ] Add examples using listers with filtered getOptions
- [ ] Add communication with the external world. Request a remote endpoint.  