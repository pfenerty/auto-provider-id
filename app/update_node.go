package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func UpdatNode(nodeName string, dryRun bool) {
	logEntry := logrus.WithFields(logrus.Fields{"node": nodeName, "dry-run": dryRun})

	clientset, err := GetKubeClientset()
	if err != nil {
		panic(err.Error())
	}

	node, err := clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	sess, err := session.NewSession()
	if err != nil {
		panic(err.Error())
	}

	ec2metadataSvc := ec2metadata.New(sess)
	if !ec2metadataSvc.Available() {
		panic(err.Error())
	}

	instanceID, err := ec2metadataSvc.GetMetadata("instance-id")
	if err != nil {
		logEntry.Errorf("Error fetching instance-id")
		panic(err.Error())
	}

	availabilityZone, err := ec2metadataSvc.GetMetadata("placement/availability-zone")
	if err != nil {
		logEntry.Errorf("Error fetching availability-zone")
		panic(err.Error())
	}

	node.Spec.ProviderID = fmt.Sprintf("aws:///%s/%s", availabilityZone, instanceID)

	if !dryRun {
		_, err = clientset.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
		if err != nil {
			panic(err.Error())
		}
	}
}
