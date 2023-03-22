#Upgrade workflow
Suppose that we want to upgrade from 1.17 to 1.18
1. Pull the latest cost from the open source repository
1. Create a new branch `release-1.18-eks` based on the open source branch `cluster-autoscaler-release-1.18`
1. Find the patch file `patch/1.17.patch` in the branch `release-1.17-eks`, try to use it apply to the current branch
1. After modifications, push the code to our repository. Then it will trigger the build https://jenkins.visenze.com/job/kubernetes-cluster-autoscaler/
1. Then we can test it in staging environment.


#How to test the cluster autoscaler work
1. Test if gpu related resources can trigger the scaling up and scaling down with this pod definition.

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gpu-pod
  labels:
    app: gpu-pod
spec:
  replicas: 0
  selector:
    matchLabels:
      app: gpu-pod
  template:
    metadata:
      labels:
        app: gpu-pod
    spec:
      nodeSelector:
        visenze.component: search
        visenze.gpu: "true"
      containers:
        - name: digits-container
          image: nvcr.io/nvidia/digits:20.12-tensorflow-py3
          #image: banst/awscli
          resources:
            limits:
              visenze.com/nvidia-gpu-memory: 8988051968
              # visenze.com/nvidia-mps-context: 20
              # nvidia.com/gpu: 1
```
2. Or you can use the the files `test-ca.sh` and `gpu-deploy-tmpl.yaml` in [scripts](cluster-autoscaler/scripts) folder to test it automatically

#Note
* If it can work, then generate and commit a new patch for the next version upgrade. The command to generate the patch:
 `git diff [commit that before applying the patch] ':(exclude)cluster-autoscaler/go.mod' ':(exclude)cluster-autoscaler/go.sum' > patch/1.18.patch`
