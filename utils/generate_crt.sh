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

# Generate the RSA key for the CA 
openssl genrsa -out ca.key 4096

cat > ca_answer.txt << EOF
[req]
default_bits = 4096
prompt = no
default_md = sha256
distinguished_name = dn 
x509_extensions = usr_cert

[ dn ]
C=US
ST=New York
L=New York
O=MyOrg
OU=MyOU
emailAddress=me@working.me
CN = server.example.com

[ usr_cert ]
basicConstraints=CA:TRUE
subjectKeyIdentifier=hash
authorityKeyIdentifier=keyid,issuer
EOF

# Generate the Custom CA
openssl req -new -x509 -key ca.key -days 730 -out ca.crt -config <( cat ca_answer.txt )

# Generate service Key
openssl genrsa -out ${service}.key 4096 # (key.pem)

# Generate the Certificate Request Answer file

cat > csr_answer.txt << EOF
[req]
default_bits = 4096
prompt = no
default_md = sha256
x509_extensions = req_ext
req_extensions = req_ext
distinguished_name = dn

[ dn ]
C=US
ST=New York
L=New York
O=MyOrg
OU=MyOrgUnit
emailAddress=me@working.me
CN = ${service}

[ req_ext ]
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = ${service}
DNS.2 = ${service}.${namespace}
DNS.3 = ${service}.${namespace}.svc
EOF

# Generate Service CSR
openssl req -new -key ${service}.key -out ${service}.csr -config <( cat csr_answer.txt )

# Test the CSR
openssl req -in ${service}.csr -noout -text | grep DNS

# Sign the CSR :
 openssl x509 -req -in ${service}.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out ${service}.crt -days 730 -extensions 'req_ext' -extfile <(cat csr_answer.txt)

 # Bundle the Certificate
 mv ${service}.crt ${service}-certonly.crt
 cat ${service}-certonly.crt ca.crt > ${service}.crt

 # Testing the Certificate
 openssl x509 -in ${service}.crt -noout -text | grep DNS

 openssl verify -CAfile ca.crt ${service}.crt

# Generating the CA_BUNDLE base64
echo "Generating your CA with BASE64 :"
AC_CA_BUNDLE=`cat ca.crt | base64 -w0`
echo "ca_base64: $AC_CA_BUNDLE"
echo "$AC_CA_BUNDLE" > ca_base64.txt
mkdir certs
cp ${service}.key certs/key.pem
cp ${service}.crt certs/cert.pem
chmod a+r certs/*
