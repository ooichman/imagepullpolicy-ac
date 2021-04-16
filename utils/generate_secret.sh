#!/bin/bash

if [ ! -x "$(command -v openssl)" ]; then
    echo "openssl not found"
    exit 1
fi

usage() {
    cat <<EOF
Generate Certificate for our Validation admission controller

This script will use openssl and geneate the certificate and key for the API to use.
More so , this script will generate a new CA for signing the certificate , because we are 
using it for admission controller , the Kubernetes Cluster must know and approve the 
certificate CA , please read the OpenShit/Kubernetes Documentation about how to add the
Admission Controller CA to your Cluster.

usage: ${0} [OPTIONS]
The following flags are required.
       --service          Service name of webhook.
       --namespace        Namespace where webhook service and secret reside.
EOF
	exit 1
}

while [[ $# -gt 0 ]]; do
    case ${1} in
        --service)
            service="$2"
            shift
            ;;
        --namespace)
            namespace="$2"
            shift
            ;;
        *)
            usage
            ;;
    esac
    shift
done

[ -z "${service}" ] && echo "service Name Not defined" && exit 1
[ -z "${namespace}" ] && echo "Namespace Name Not defined" && exit 1

# Generate the Secret
kubectl create secret generic "${service}"-tls \
--from-file=key.pem=`pwd`/${service}.key \
--from-file=cert.pem=`pwd`/${service}.crt \
-n "${namespace}"
