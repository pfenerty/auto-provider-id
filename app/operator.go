package main

import (
	"context"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/google/uuid"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func StartOperator(dryRunMode bool) {
	if dryRunMode {
		logrus.Info("Starting Auto Provider ID Operator in dry run mdoe")
	} else {
		logrus.Info("Starting Auto Provider ID Operator")
	}

	clientset, err := GetKubeClientset()
	if err != nil {
		panic(err.Error())
	}

	imageRef := getImageRef(clientset, os.Getenv("POD_NAMESPACE"), os.Getenv("POD_NAME"))

	// Watch for new nodes
	watchlist := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"nodes",
		metav1.NamespaceAll,
		fields.Everything(),
	)

	_, controller := cache.NewInformer(
		watchlist,
		&corev1.Node{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				node := obj.(*corev1.Node)

				logEntry := logrus.WithFields(logrus.Fields{"node": node.Name})
				logEntry.Infof("New node detected %s", node.Name)

				err := updateNode(clientset, imageRef, node, dryRunMode)
				if err != nil && !errors.IsConflict(err) {
					logEntry.Errorf("Failed start update job on node: %v", err)
				} else {
					logEntry.Info("Update job applied to node.")
				}
			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)

	// Keep the program running
	select {}
}

func updateNode(clientset *kubernetes.Clientset, imageRef string, node *corev1.Node, dryRunMode bool) error {
	ctx := context.TODO()

	logEntry := logrus.WithFields(logrus.Fields{"node": node.Name})

	u, err := uuid.NewUUID()
	if err != nil {
		logEntry.Error("Error Creating UUID")
		return err
	}

	args := []string{"--node", node.Name}
	if dryRunMode {
		args = append(args, "--dry-run")
	}

	namespace := os.Getenv("POD_NAMESPACE")

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "node-updater-" + u.String(),
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "auto-provider-id",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "node-updater",
					Tolerations: []corev1.Toleration{
						{
							Key:      "node-role.kubernetes.io/control-plane",
							Operator: "Exists",
							Effect:   "NoSchedule",
						},
						{
							Key:      "node.cloudprovider.kubernetes.io/uninitialized",
							Operator: "Equal",
							Value:    "true",
							Effect:   "NoSchedule",
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "updater",
							Image:           imageRef,
							Command:         []string{"/ko-app/auto-provider-id", "update-node"},
							Args:            args,
							ImagePullPolicy: "Always",
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: &[]bool{false}[0],
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
								RunAsNonRoot: &[]bool{true}[0],
								RunAsUser:    &[]int64{1000}[0],
								SeccompProfile: &corev1.SeccompProfile{
									Type: corev1.SeccompProfileTypeRuntimeDefault,
								},
							},
						},
					},
					NodeSelector: map[string]string{
						"kubernetes.io/hostname": node.Name,
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
		},
	}

	logEntry.WithFields(logrus.Fields{"job": job}).Debug()

	_, err = clientset.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	return err
}

func getImageRef(clientset *kubernetes.Clientset, namespace string, podname string) string {
	for {
		pod, err := clientset.CoreV1().Pods(namespace).Get(
			context.TODO(), podname, metav1.GetOptions{},
		)
		if err != nil {
			panic(err.Error())
		}

		if pod.Status.ContainerStatuses[0].ImageID != "" {
			logrus.Infof("Using Image: %s", pod.Status.ContainerStatuses[0].ImageID)
			return pod.Status.ContainerStatuses[0].ImageID
		}
		logrus.Infof("Blank ImageID, trying again in 10 seconds")
		time.Sleep(10 * time.Second)
	}
}
