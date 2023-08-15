package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func UpdateNode(nodeName string, dryRun bool) {
	logEntry := logrus.WithFields(logrus.Fields{"node": nodeName, "dry-run": dryRun})
	logEntry.Info("")

	clientset, err := GetKubeClientset()
	if err != nil {
		logEntry.Error("Failed to get Kube Client Set")
		panic(err.Error())
	}

	node, err := clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		logEntry.Errorf("Failed to get Node %s", nodeName)
		panic(err.Error())
	}

	sess, err := session.NewSession()
	if err != nil {
		logEntry.Error("Failed to create AWS session")
		panic(err.Error())
	}

	ec2metadataSvc := ec2metadata.New(sess)
	if !ec2metadataSvc.Available() {
		panic("EC2 metadata service not available")
	}

	instanceID, err := ec2metadataSvc.GetMetadata("instance-id")
	if err != nil {
		logEntry.Errorf("Error fetching instance-id")
		panic(err.Error())
	}
	logEntry.Infof("instance-id: %s", instanceID)

	availabilityZone, err := ec2metadataSvc.GetMetadata("placement/availability-zone")
	if err != nil {
		logEntry.Errorf("Error fetching availability-zone")
		panic(err.Error())
	}
	logEntry.Infof("availability-zone: %s", availabilityZone)

	node.Spec.ProviderID = fmt.Sprintf("aws:///%s/%s", availabilityZone, instanceID)

	if !dryRun {
		_, err = clientset.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
		if err != nil {
			panic(err.Error())
		}
	}
}
