package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"os"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	examplev1 "secretswatcher-operator/api/v1"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(examplev1.AddToScheme(scheme))
}

type SecretWatcherReconciler struct {
	client.Client
	Log logr.Logger
}

func randomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// Reconcile rotates the referenced secret's data when the CR changes
func (r *SecretWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("secretwatcher", req.NamespacedName)

	var sw examplev1.SecretWatcher
	if err := r.Get(ctx, req.NamespacedName, &sw); err != nil {
		log.Error(err, "unable to fetch SecretWatcher")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Fetch the referenced Secret
	var secret corev1.Secret
	if err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: sw.Spec.SecretName}, &secret); err != nil {
		log.Error(err, "unable to fetch referenced Secret")
		return ctrl.Result{}, nil // Don't requeue if secret is missing
	}

	// Rotate all data fields
	for k := range secret.Data {
		secret.Data[k] = []byte(randomString(16))
	}

	if err := r.Update(ctx, &secret); err != nil {
		log.Error(err, "failed to update Secret")
		return ctrl.Result{}, err
	}

	// Update status
	// sw.Status.LastRotated = metav1.NewTime(time.Now())
	// if err := r.Status().Update(ctx, &sw); err != nil {
	// 	log.Error(err, "failed to update SecretWatcher status")
	// }

	log.Info("Rotated secret", "secret", secret.Name)
	return ctrl.Result{RequeueAfter: 15 * time.Second}, nil // Requeue for periodic rotation
}

func main() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	namespace := os.Getenv("WATCH_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&SecretWatcherReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("SecretWatcher"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecretWatcher")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// SetupWithManager wires up the controller to watch SecretWatcher resources
func (r *SecretWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplev1.SecretWatcher{}).
		Complete(r)
}
