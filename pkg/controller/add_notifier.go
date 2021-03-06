package controller

import (
	"github.com/housinganywhere/k8s-notify/pkg/controller/notifier"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, notifier.Add)
}
