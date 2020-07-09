package nginx

import (
	"context"
	"fmt"
	"os"

	appv1alpha1 "github.com/chenzhiwei/nginx-operator/pkg/apis/app/v1alpha1"
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

	// Initialize the default Nginx CR
	initInstance(mgr.GetClient())

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

	// Define a new Deployment object
	deployment := newDeploymentForCR(instance)

	// Set Nginx deployment as the owner and controller
	if err := controllerutil.SetControllerReference(instance, deployment, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Deployment already exists
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, found)
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

	// Deployment already exists - don't requeue
	reqLogger.Info("Skip reconcile: Deployment already exists", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
	return reconcile.Result{}, nil
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
							Name:    "init",
							Image:   "quay.io/siji/nginx:fake",
							Command: []string{"sleep", "5"},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "quay.io/siji/nginx:fake",
						},
					},
				},
			},
		},
	}
}

func initInstance(client client.Client) error {
	name := "default"
	namespace := os.Getenv("WATCH_NAMESPACE")

	log.Info(fmt.Sprintf("Trying to create Nginx CR with name %s and namespace %s", name, namespace))

	nginx := newNginxCR(name, namespace)
	err := client.Create(context.TODO(), nginx)
	if errors.IsAlreadyExists(err) {
		log.Info(fmt.Sprintf("Nginx CR with name %s and namespace %s already exist", name, namespace))
		return nil
	}

	if err == nil {
		log.Info("Nginx CR was created successfully")
		return nil
	}

	log.Error(err, "Error creating Nginx CR, please create it manually")
	return err
}

func newNginxCR(name, namespace string) *appv1alpha1.Nginx {
	return &appv1alpha1.Nginx{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appv1alpha1.NginxSpec{
			Replicas: 1,
		},
	}
}

func int32Ptr(i int32) *int32 { return &i }
