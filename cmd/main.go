package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/openshift/api"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("cmd")

func main() {

	logf.SetLogger(logf.ZapLogger(true))

	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		panic(err)
	}

	stopCh := signals.SetupSignalHandler()
	if started := mgr.GetCache().WaitForCacheSync(stopCh); !started {
		panic("not all started")
	}

	// add openshift API

	if err := api.Install(scheme.Scheme); err != nil {
		panic(err)
	}

	// dump registry

	all := scheme.Scheme.AllKnownTypes()
	fmt.Println("Dumping known types ...")
	var keys []schema.GroupVersionKind
	for k := range all {
		keys = append(keys, k)
	}
	sort.Slice(keys[:], func(i, j int) bool {
		k1 := keys[i]
		k2 := keys[j]
		if k1.Group != k2.Group {
			return k1.Group < k2.Group
		}
		if k1.Version != k2.Version {
			return k1.Version < k2.Version
		}
		return k1.Kind < k2.Kind
	})

	for _, k := range keys {
		v := all[k]
		fmt.Printf("  %v -> %v\n", k, v)
	}

	// get entry which is problematic for controller-runtime

	gvk, _, err := scheme.Scheme.ObjectKinds(&corev1.SecretList{})
	if err != nil {
		panic(err)
	}

	fmt.Println("Dumping GVK ...")
	for _, i := range gvk {
		fmt.Printf(" - %v\n", i)
	}

	// get client

	client := mgr.GetClient()

	// get unstructured

	usecrets := unstructured.UnstructuredList{}
	usecrets.SetKind("SecretList")
	usecrets.SetAPIVersion("v1")
	if err := client.List(context.TODO(), nil, &usecrets); err != nil {
		panic(err)
	}

	log.Info("my unstructured secrets", "secrets", len(usecrets.Items))

	// get typed

	var secrets corev1.SecretList
	if err := client.List(context.TODO(), nil, &secrets); err != nil {
		panic(err)
	}

	log.Info("my secrets", "secrets", len(secrets.Items))

}
