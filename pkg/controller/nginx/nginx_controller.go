package nginx

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appv1alpha1 "github.com/chenzhiwei/nginx-operator/pkg/apis/app/v1alpha1"
)

var log = logf.Log.WithName("controller_nginx")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Nginx Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNginx{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("nginx-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Nginx
	err = c.Watch(&source.Kind{Type: &appv1alpha1.Nginx{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Deployments and requeue the owner Nginx
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.Nginx{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileNginx implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNginx{}

// ReconcileNginx reconciles a Nginx object
type ReconcileNginx struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Nginx object and makes changes based on the state read
// and what is in the Nginx.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Deployment as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNginx) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Nginx")

	// Fetch the Nginx instance
	instance := &appv1alpha1.Nginx{}
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

	// Delete event triggered
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// add the deletion logic here
	}

	// Define a new Deployment object
	deployment := newDeploymentForCR(instance)

	// Set Nginx deployment as the owner
	if err := controllerutil.SetControllerReference(instance, deployment, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Deployment already exists
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, found)
	// The Deployment does not exist, create it
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.client.Create(context.TODO(), deployment)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Deployment created successfully - don't requeue
		return reconcile.Result{}, nil

	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Check if the Deployment is desired state
	if isDeploymentChanged(deployment, found) {
		// Update the Deployment into desired state
		err = r.client.Update(context.TODO(), deployment)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func isDeploymentChanged(d1, d2 *appsv1.Deployment) bool {
	if *d1.Spec.Replicas != *d2.Spec.Replicas {
		return true
	}

	if len(d1.Spec.Template.Spec.InitContainers) != len(d2.Spec.Template.Spec.InitContainers) {
		return true
	} else {
		for i, container := range d1.Spec.Template.Spec.InitContainers {
			if !reflect.DeepEqual(container.Resources, d2.Spec.Template.Spec.InitContainers[i].Resources) {
				return true
			}
		}
	}

	if len(d1.Spec.Template.Spec.Containers) != len(d2.Spec.Template.Spec.Containers) {
		return true
	} else {
		for i, container := range d1.Spec.Template.Spec.Containers {
			if !reflect.DeepEqual(container.Resources, d2.Spec.Template.Spec.Containers[i].Resources) {
				return true
			}
		}
	}

	// add your other comparisons here

	return false
}

// newDeploymentForCR returns an Nginx deployment with the same name/namespace as the cr
func newDeploymentForCR(cr *appv1alpha1.Nginx) *appsv1.Deployment {
	labels := map[string]string{
		"app": cr.Name,
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(cr.Spec.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:      "init",
							Image:     "quay.io/siji/nginx:fake",
							Command:   []string{"sleep", "5"},
							Resources: cr.Spec.InitContainer.Resources,
						},
					},
					Containers: []corev1.Container{
						{
							Name:      "nginx",
							Image:     "quay.io/siji/nginx:fake",
							Resources: cr.Spec.AppContainer.Resources,
						},
					},
				},
			},
		},
	}
}

func int32Ptr(i int32) *int32 { return &i }
