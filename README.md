# Auto Provider ID

The motivation for this app was to be able to creae an autoscaling cluster in AWS using Talos Linux and Terraform. With Talos' immutable nature, there's no way to query AWS metadata on startup to dynamically set the `providerID` of the node on startup through a script or similar.

Without the `providerID`, the AWS cloud controller manager and kubernetes cluster autoscaler will fail to run.

The operator runs in the cluster and watches for new nodes. When a node new enters the cluster, the operator creates a job on the new node. This job uses the AWS SDK to query the metadata for the node / EC2 instance and update the node's spec with `providerID`: `aws:///<availability-zone>/<instance-id>`.