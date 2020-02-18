# Code

## Real time client

The `client.Reader` is the real time client.

```
client.Reader reader: mgr.GetAPIReader()
client.Client client: mgr.GetClient()
```


## Watch multiple namespaces

Update the `cmd/manager/main.go` code with following:

```
// This example creates a new Manager that has a cache scoped to a list of namespaces.
func ExampleNew_multinamespaceCache() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "unable to get kubeconfig")
		os.Exit(1)
	}

	mgr, err := manager.New(cfg, manager.Options{
		NewCache: cache.MultiNamespacedCacheBuilder([]string{"namespace1", "namespace2"}),
	})
	if err != nil {
		log.Error(err, "unable to set up manager")
		os.Exit(1)
	}
	log.Info("created manager", "manager", mgr)
}
```

Refer: https://github.com/kubernetes-sigs/controller-runtime/pull/267/files
