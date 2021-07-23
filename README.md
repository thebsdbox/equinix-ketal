# Equinix Ketal

Equinix Metal API objects accessible in Kubernetes

## Initialise the project

This requires kubebuilder

Generate Environment:
```
$ mkdir $HOME/go/src/github.com/$USER/equinix-ketal; cd $HOME/go/src/github.com/$USER/equinix-ketal
$ # Create project
$ kubebuilder init --domain equinix.metal --owner "Dan Finneran"
$ # Create CRD/Controllers
$ kubebuilder create api --group ketal --kind Eip --version v1
$ kubebuilder create api --group ketal --kind Device --version v1
```

## 