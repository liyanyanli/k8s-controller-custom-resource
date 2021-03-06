package main

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	democrdv1 "github.com/liyanyanli/k8s-controller-custom-resource/pkg/apis/democrd/v1"
	clientset "github.com/liyanyanli/k8s-controller-custom-resource/pkg/client/clientset/versioned"
	democrdkscheme "github.com/liyanyanli/k8s-controller-custom-resource/pkg/client/clientset/versioned/scheme"
	informers "github.com/liyanyanli/k8s-controller-custom-resource/pkg/client/informers/externalversions/democrd/v1"
	listers "github.com/liyanyanli/k8s-controller-custom-resource/pkg/client/listers/democrd/v1"
)

const controllerAgentName = "democrd-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Democrd is synced
	SuccessSynced = "Synced"

	// MessageResourceSynced is the message used for an Event fired when a Democrd
	// is synced successfully
	MessageResourceSynced = "Demo synced successfully"
)

// Controller is the controller implementation for Democrd resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// democrdclientset is a clientset for our own API group
	democrdclientset clientset.Interface

	democrdsLister listers.DemocrdLister
	democrdsSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new Democrd controller
func NewController(
	kubeclientset kubernetes.Interface,
	democrdclientset clientset.Interface,
	democrdInformer informers.DemocrdInformer) *Controller {

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(democrdkscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:    kubeclientset,
		democrdclientset: democrdclientset,
		democrdsLister:   democrdInformer.Lister(),
		democrdsSynced:   democrdInformer.Informer().HasSynced,
		workqueue:        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Democrds"),
		recorder:         recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Democrd resources change
	democrdInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueDemocrd,
		UpdateFunc: func(old, new interface{}) {
			oldDemocrd := old.(*democrdv1.Democrd)
			newDemocrd := new.(*democrdv1.Democrd)
			if oldDemocrd.ResourceVersion == newDemocrd.ResourceVersion {
				// Periodic resync will send update events for all known Democrds.
				// Two different versions of the same Democrd will always have different RVs.
				return
			}
			controller.enqueueDemocrd(new)
		},
		DeleteFunc: controller.enqueueDemocrdForDelete,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Democrd control loop")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.democrdsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Democrd resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Democrd resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Democrd resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Democrd resource with this namespace/name
	democrd, err := c.democrdsLister.Democrds(namespace).Get(name)
	if err != nil {
		// The Democrd resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			klog.Warningf("Democrd: %s/%s does not exist in local cache, will delete it from ...",
				namespace, name)

			klog.Infof("Deleting Democrd: %s/%s ...", namespace, name)

			// FIX ME: call API to delete this democrd by name.
			//
			// democrd.Delete(namespace, name)

			return nil
		}

		runtime.HandleError(fmt.Errorf("failed to list Democrd by: %s/%s", namespace, name))

		return err
	}

	klog.Infof("Try to process Democrd: %#v ...", democrd)

	// FIX ME: Do diff().
	//
	// actualDemocrd, exists := democrd.Get(namespace, name)
	//
	// if !exists {
	// 	democrd.Create(namespace, name)
	// } else if !reflect.DeepEqual(actualDemocrd, democrd) {
	// 	democrd.Update(namespace, name)
	// }

	c.recorder.Event(democrd, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

// enqueueDemocrd takes a Democrd resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Democrd.
func (c *Controller) enqueueDemocrd(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// enqueueDemocrdForDelete takes a deleted Democrd resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Democrd.
func (c *Controller) enqueueDemocrdForDelete(obj interface{}) {
	var key string
	var err error
	key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}
