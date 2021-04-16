#!/bin/bash

CA_BASE64=$(/bin/cat ca_base64.txt)

echo 'apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: "imagepullpolicy.il.redhat.io"
webhooks:
- name: "imagepullpolicy.il.redhat.io"
  reinvocationPolicy: IfNeeded
  namespaceSelector:
    matchExpressions:
    - key: admission.il.redhat.io/imagePullPolicy
      operator: In
      values: ["True"]
  rules:
  - apiGroups:   [""]
    apiVersions: ["v1"]
    operations:  ["CREATE","UPDATE"]
    resources:   ["pods"]
    scope:       "Namespaced"
  clientConfig:
    service:
      namespace: "kube-ippac"
      name: "ippac"
      path: /mutate
      port: 8443
    caBundle: <CA_BASE64>
  admissionReviewVersions: ["v1", "v1beta1"]
  sideEffects: None' | sed "s/<CA_BASE64>/${CA_BASE64}/g"
