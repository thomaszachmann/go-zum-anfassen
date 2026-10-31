package pods

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// Watch meldet jedes neue Finding über report, bis ctx endet.
func Watch(ctx context.Context, cs kubernetes.Interface, ns string,
	minRestarts int32, report func(Finding)) error {
	factory := informers.NewSharedInformerFactoryWithOptions(cs, 0,
		informers.WithNamespace(ns))
	inf := factory.Core().V1().Pods().Informer()
	seen := map[Finding]bool{} // Handler laufen seriell: kein Mutex nötig
	check := func(obj any) {
		p, ok := obj.(*corev1.Pod)
		if !ok {
			return
		}
		for _, f := range Inspect(p, minRestarts) {
			if !seen[f] {
				seen[f] = true
				report(f)
			}
		}
	}
	_, err := inf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    check,
		UpdateFunc: func(_, obj any) { check(obj) },
	})
	if err != nil {
		return err
	}
	factory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), inf.HasSynced) {
		return errors.New("informer-cache nicht synchron")
	}
	<-ctx.Done()
	factory.Shutdown()
	return nil
}
