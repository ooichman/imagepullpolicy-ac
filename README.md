# imagepullpolicy-ac

an Admission controller for Image Pull Policy

This is a very Simple and Small tutorial about an effective way to use Mutate Admission Controller
and validating admission controller at your own cluster.

In Order to use it (and Learn from it) all you need to do is clone the repository :

## Clone and Generate
```bash
$ git clone https://github.com/ooichman/imagepullpolicy-ac.git
```

Now we need to run a couple of tasks before all of our setup is complete.

```bash
$ cd imagepullpolicy-ac
```

### Generate Certificate

Our communication between the Server API and the Web Hook admission controller is done TLS (you can checkout the main.go file).
In order for that to work we need to generate a CA , a certificate request and then sign the certificate with the CA.
Our webhook will run in a service under a namespace so we need to generate the certificate accordingly.

Once we clone the repository our tools are in the utils directory.
```bash
$ cd utils
```
Now Let's generate the certificate: 

```bash
$ ./generate_crt.sh --service ippac --namespace kube-ippac
```

### Deploy the Pod infrastructure :

Now we need to go back to the deployment and create the namespace
```bash
$ cd ..
```
```bash
$ kubectl create -f deploy/namespace.yaml
```
and create the service :
```bash
$ kubectl create -f deploy/service.yaml
```

Now we need to create the secret before we create the deployment :
```bash
$ cd utils/
```
And Generate
```bash
$ ./generate_secret.sh --service ippac --namespace kube-ippac
```

Go Back to the parent directory:
```bash
$ cd ..
```

Now we can go ahead and create the deployment
( NOTE that is you are working in a disconnected environment then save the image and then change the image path
to your internal directory)  

```bash
$ kubectl create -f deploy/deployment.yaml
```

### create the admission controllers

For deploying the admission controller I have create a simple script that reads the CA base64 which was generated , 
push it to our YAML configuration which leave us to apply it :

Change to the utils directory
```bash
$ cd utils/
```
Now run the generate command to see how the admission controller setting should look like with our newly created CA :
```bash
$ ./generate_validatingwebhook.sh 
```

and create it :
```bash
$ ./generate_validatingwebhook.sh | kubectl create -f -
validatingwebhookconfiguration.admissionregistration.k8s.io/imagepullpolicy.il.redhat.io created
```

Now we will run the same way for the mutating admission controller :  
```bash
$ ./generate_mutate_yaml.sh | kubectl create -f -
mutatingwebhookconfiguration.admissionregistration.k8s.io/imagepullpolicy.il.redhat.io created
```

### testing
To test you rconfiguration I have created a tests directory so you can check your deployment.

Go back to the main directory :
```bash
$ cd ..
```

First deploy our test namespace:
```bash
$ oc create -f tests/namespace.yaml
```
Now we can deploy both the deployment and the signal pod :
```bash
$ kubectl create -f tests/deployment-working.yaml -f tests/pod-testing.yaml
```

Once you deployed both of them you can run a JSON query and see that you are getting the Always results for the imagePullPolicy key :
```bash
$ kubectl get pod -o name | \
     grep ippac-example | \
     xargs kubectl get -o jsonpath='{.spec.containers[0].imagePullPolicy}'  ; echo
Always
```

### Clean UP

In Order to clean up just remove the ns 
```bash
$ kubectl delete -f tests/namespace.yaml
```

You should see the "Always" output at the end of the command

### your own namespace
Now if you want your namespace to automatically be updated for the imagePullPolicy of Always in your namespace just add the label for it :

```bash
$ kubectl label ns <your namespace> admission.il.redhat.io/imagePullPolicy=True
```
And you are done

If you have any question feel free to responed/leave a comment.