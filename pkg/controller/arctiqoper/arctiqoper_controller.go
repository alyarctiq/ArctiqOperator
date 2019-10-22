package arctiqoper

import (
	"context"
	"log"
	"reflect"

	appv1alpha1 "github.com/arctiq-operator/ArcitqOper/pkg/apis/app/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	//logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

//var log = logf.Log.WithName("controller_arctiqoper")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ArctiqOper Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileArctiqOper{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("arctiqoper-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ArctiqOper
	err = c.Watch(&source.Kind{Type: &appv1alpha1.ArctiqOper{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ArctiqOper
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.ArctiqOper{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileArctiqOper implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileArctiqOper{}

// ReconcileArctiqOper reconciles a ArctiqOper object
type ReconcileArctiqOper struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ArctiqOper object and makes changes based on the state read
// and what is in the ArctiqOper.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileArctiqOper) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling ArctiqOper %s/%s\n", request.Namespace, request.Name)

	// Fetch the ArctiqOper instance
	instance := &appv1alpha1.ArctiqOper{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	//found := &appsv1.Deployment{}
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {

		// Define a new deployment
		dep := r.newDeploymentForCR(instance)
		log.Printf("Creating a new Deployment %s/%s\n", dep.Namespace, dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			log.Printf("Failed to create new Deployment: %v\n", err)
			return reconcile.Result{}, err
		}

		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		log.Printf("Failed to get Deployment: %v\n", err)
		return reconcile.Result{}, err
	}

	// Ensure the deployment Count is the same as the spec
	count := instance.Spec.Count
	if *found.Spec.Replicas != count {
		found.Spec.Replicas = &count
		err = r.client.Update(context.TODO(), found)
		if err != nil {
			log.Printf("Failed to update Deployment: %v\n", err)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	// List the pods for this deployment
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForArctiqOper(instance.Name))
	listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
	err = r.client.List(context.TODO(), listOps, podList)
	if err != nil {
		log.Printf("Failed to list pods: %v", err)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.PodNames if needed
	if !reflect.DeepEqual(podNames, instance.Status.PodNames) {
		instance.Status.PodNames = podNames
		err := r.client.Update(context.TODO(), instance)
		if err != nil {
			log.Printf("failed to update node status: %v", err)
			return reconcile.Result{}, err
		}
	}

	// Update AppGroup status
	if instance.Spec.Group != instance.Status.AppGroup {
		instance.Status.AppGroup = instance.Spec.Group
		err := r.client.Update(context.TODO(), instance)
		if err != nil {
			log.Printf("failed to update group status: %v", err)
			return reconcile.Result{}, err
		}
	}

	//reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

// getPodNames returns the pod names of the array of pods passed in.
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

//Set labels in a map.
func labelsForArctiqOper(name string) map[string]string {
	return map[string]string{"app": "Arctiq-Operator", "arctiqoperator_cr": name}
}

// Create newDeploymentForCR method to create a deployment.
func (r *ReconcileArctiqOper) newDeploymentForCR(m *appv1alpha1.ArctiqOper) *appsv1.Deployment {
	labels := labelsForArctiqOper(m.Name)
	replicas := m.Spec.Count
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: m.Spec.Image,
						Name:  m.Name,
						Ports: []corev1.ContainerPort{{
							ContainerPort: m.Spec.Port,
							Name:          m.Name,
						}},
					}},
				},
			},
		},
	}
	// Set ArctiqOper instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.scheme)
	return dep

}
