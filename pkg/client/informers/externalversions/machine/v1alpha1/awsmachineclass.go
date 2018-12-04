// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	machine_v1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	versioned "github.com/gardener/machine-controller-manager/pkg/client/clientset/versioned"
	internalinterfaces "github.com/gardener/machine-controller-manager/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/gardener/machine-controller-manager/pkg/client/listers/machine/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// AWSMachineClassInformer provides access to a shared informer and lister for
// AWSMachineClasses.
type AWSMachineClassInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.AWSMachineClassLister
}

type aWSMachineClassInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewAWSMachineClassInformer constructs a new informer for AWSMachineClass type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewAWSMachineClassInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredAWSMachineClassInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredAWSMachineClassInformer constructs a new informer for AWSMachineClass type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredAWSMachineClassInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.MachineV1alpha1().AWSMachineClasses(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.MachineV1alpha1().AWSMachineClasses(namespace).Watch(options)
			},
		},
		&machine_v1alpha1.AWSMachineClass{},
		resyncPeriod,
		indexers,
	)
}

func (f *aWSMachineClassInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredAWSMachineClassInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *aWSMachineClassInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&machine_v1alpha1.AWSMachineClass{}, f.defaultInformer)
}

func (f *aWSMachineClassInformer) Lister() v1alpha1.AWSMachineClassLister {
	return v1alpha1.NewAWSMachineClassLister(f.Informer().GetIndexer())
}
