package controller

import (
	"github.com/arctiq-operator/ArcitqOper/pkg/controller/arctiqoper"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, arctiqoper.Add)
}
