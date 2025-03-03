// Copyright 2020 NetApp, Inc. All Rights Reserved.

package k8sclient

import (
	"fmt"
	"strconv"
	"strings"

	commonconfig "github.com/netapp/trident/config"
	"github.com/netapp/trident/utils"
)

const (
	TridentAppLabelKey       = "app"
	DefaultContainerLabelKey = "kubectl.kubernetes.io/default-container"
)

func GetNamespaceYAML(namespace string) string {
	return strings.ReplaceAll(namespaceYAMLTemplate, "{NAMESPACE}", namespace)
}

const namespaceYAMLTemplate = `---
apiVersion: v1
kind: Namespace
metadata:
  name: {NAMESPACE}
`

func GetServiceAccountYAML(serviceAccountName string, secrets []string, labels,
	controllingCRDetails map[string]string) string {

	var saYAML string

	if len(secrets) > 0 {
		saYAML = serviceAccountWithSecretYAML
		saYAML = strings.Replace(saYAML, "{SECRETS}", constructServiceAccountSecrets(secrets), 1)
	} else {
		saYAML = serviceAccountYAML
	}

	saYAML = strings.ReplaceAll(saYAML, "{NAME}", serviceAccountName)
	saYAML = replaceMultiline(saYAML, labels, controllingCRDetails, nil)

	return saYAML
}

const serviceAccountYAML = `---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {NAME}
  {LABELS}
  {OWNER_REF}
`

const serviceAccountWithSecretYAML = `---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {NAME}
  {LABELS}
  {OWNER_REF}
{SECRETS}
`

func GetClusterRoleYAML(flavor OrchestratorFlavor, clusterRoleName string, labels,
	controllingCRDetails map[string]string, csi bool) string {

	var clusterRoleYAML string

	if csi {
		clusterRoleYAML = clusterRoleCSIYAMLTemplate
	} else {
		clusterRoleYAML = clusterRoleYAMLTemplate
	}

	// authorization.openshift.io/v1 is applicable to OCP 3.x only
	if flavor == FlavorOpenShift && !csi {
		clusterRoleYAML = strings.ReplaceAll(clusterRoleYAML, "{API_VERSION}", "authorization.openshift.io/v1")
	} else {
		clusterRoleYAML = strings.ReplaceAll(clusterRoleYAML, "{API_VERSION}", "rbac.authorization.k8s.io/v1")
	}

	clusterRoleYAML = strings.ReplaceAll(clusterRoleYAML, "{CLUSTER_ROLE_NAME}", clusterRoleName)
	clusterRoleYAML = replaceMultiline(clusterRoleYAML, labels, controllingCRDetails, nil)

	return clusterRoleYAML
}

const clusterRoleYAMLTemplate = `---
kind: ClusterRole
apiVersion: {API_VERSION}
metadata:
  name: {CLUSTER_ROLE_NAME}
  {LABELS}
  {OWNER_REF}
rules:
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["persistentvolumes", "persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "create", "delete", "update", "patch"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims/status"]
    verbs: ["update", "patch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["watch", "create", "update", "patch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch", "create", "delete", "update", "patch"]
  - apiGroups: ["apiextensions.k8s.io"]
    resources: ["customresourcedefinitions"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["trident.netapp.io"]
    resources: ["tridentversions", "tridentbackends", "tridentstorageclasses", "tridentvolumes","tridentnodes",
"tridenttransactions", "tridentsnapshots", "tridentbackendconfigs", "tridentbackendconfigs/status"]
    verbs: ["get", "list", "watch", "create", "delete", "update", "patch"]
  - apiGroups: ["policy"]
    resources: ["podsecuritypolicies"]
    verbs: ["use"]
    resourceNames:
      - tridentpods
`

const clusterRoleCSIYAMLTemplate = `---
kind: ClusterRole
apiVersion: {API_VERSION}
metadata:
  name: {CLUSTER_ROLE_NAME}
  {LABELS}
  {OWNER_REF}
rules:
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["persistentvolumes", "persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "create", "delete", "update", "patch"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims/status"]
    verbs: ["update", "patch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch", "create", "delete", "update", "patch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch", "create", "delete", "update", "patch"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch", "create", "delete", "update", "patch"]
  - apiGroups: [""]
    resources: ["pods/log"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update", "patch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments/status"]
    verbs: ["update", "patch"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshots", "volumesnapshotclasses"]
    verbs: ["get", "list", "watch", "update", "patch"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshots/status", "volumesnapshotcontents/status"]
    verbs: ["update", "patch"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshotcontents"]
    verbs: ["get", "list", "watch", "create", "delete", "update", "patch"]
  - apiGroups: ["csi.storage.k8s.io"]
    resources: ["csidrivers", "csinodeinfos"]
    verbs: ["get", "list", "watch", "create", "delete", "update", "patch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["csidrivers", "csinodes"]
    verbs: ["get", "list", "watch", "create", "delete", "update", "patch"]
  - apiGroups: ["apiextensions.k8s.io"]
    resources: ["customresourcedefinitions"]
    verbs: ["get", "list", "watch", "create", "delete", "update", "patch"]
  - apiGroups: ["trident.netapp.io"]
    resources: ["tridentversions", "tridentbackends", "tridentstorageclasses", "tridentvolumes","tridentnodes",
"tridenttransactions", "tridentsnapshots", "tridentbackendconfigs", "tridentbackendconfigs/status"]
    verbs: ["get", "list", "watch", "create", "delete", "update", "patch"]
  - apiGroups: ["policy"]
    resources: ["podsecuritypolicies"]
    verbs: ["use"]
    resourceNames:
      - tridentpods
`

func GetClusterRoleBindingYAML(namespace string, flavor OrchestratorFlavor, name string, labels, controllingCRDetails map[string]string, csi bool) string {

	var crbYAML string

	// authorization.openshift.io/v1 is applicable to OCP 3.x only
	if flavor == FlavorOpenShift && !csi {
		crbYAML = clusterRoleBindingOpenShiftYAMLTemplate
	} else {
		crbYAML = clusterRoleBindingKubernetesV1YAMLTemplate
	}

	crbYAML = strings.ReplaceAll(crbYAML, "{NAMESPACE}", namespace)
	crbYAML = strings.ReplaceAll(crbYAML, "{NAME}", name)
	crbYAML = replaceMultiline(crbYAML, labels, controllingCRDetails, nil)
	return crbYAML
}

const clusterRoleBindingOpenShiftYAMLTemplate = `---
kind: ClusterRoleBinding
apiVersion: authorization.openshift.io/v1
metadata:
  name: {NAME}
  {LABELS}
  {OWNER_REF}
subjects:
  - kind: ServiceAccount
    name: {NAME}
    namespace: {NAMESPACE}
roleRef:
  name: {NAME}
`

const clusterRoleBindingKubernetesV1YAMLTemplate = `---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: {NAME}
  {LABELS}
  {OWNER_REF}
subjects:
  - kind: ServiceAccount
    name: {NAME}
    namespace: {NAMESPACE}
roleRef:
  kind: ClusterRole
  name: {NAME}
  apiGroup: rbac.authorization.k8s.io
`

func GetDeploymentYAML(deploymentName, tridentImage, logFormat string, imagePullSecrets []string, labels,
	controllingCRDetails map[string]string, debug bool) string {

	var debugLine string

	if debug {
		debugLine = "- -debug"
	} else {
		debugLine = "#- -debug"
	}

	deploymentYAML := strings.ReplaceAll(deploymentYAMLTemplate, "{TRIDENT_IMAGE}", tridentImage)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{DEPLOYMENT_NAME}", deploymentName)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{DEBUG}", debugLine)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{LABEL_APP}", labels[TridentAppLabelKey])
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{LOG_FORMAT}", logFormat)
	deploymentYAML = replaceMultiline(deploymentYAML, labels, controllingCRDetails, imagePullSecrets)

	return deploymentYAML
}

const deploymentYAMLTemplate = `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {DEPLOYMENT_NAME}
  {LABELS}
  {OWNER_REF}
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: {LABEL_APP}
  template:
    metadata:
      labels:
        app: {LABEL_APP}
    spec:
      serviceAccount: trident
      containers:
      - name: trident-main
        image: {TRIDENT_IMAGE}
        command:
        - /trident_orchestrator
        args:
        - "--crd_persistence"
        - "--k8s_pod"
        - "--log_format={LOG_FORMAT}"
        {DEBUG}
        livenessProbe:
          exec:
            command:
            - tridentctl
            - -s
            - 127.0.0.1:8000
            - version
          failureThreshold: 2
          initialDelaySeconds: 120
          periodSeconds: 120
          timeoutSeconds: 90
      {IMAGE_PULL_SECRETS}
      nodeSelector:
        beta.kubernetes.io/os: linux
        beta.kubernetes.io/arch: amd64
`

func GetCSIServiceYAML(serviceName string, labels, controllingCRDetails map[string]string) string {

	serviceYAML := strings.ReplaceAll(serviceYAMLTemplate, "{LABEL_APP}", labels[TridentAppLabelKey])
	serviceYAML = strings.ReplaceAll(serviceYAML, "{SERVICE_NAME}", serviceName)
	serviceYAML = replaceMultiline(serviceYAML, labels, controllingCRDetails, nil)
	return serviceYAML
}

const serviceYAMLTemplate = `---
apiVersion: v1
kind: Service
metadata:
  name: {SERVICE_NAME}
  {LABELS}
  {OWNER_REF}
spec:
  selector:
    app: {LABEL_APP}
  ports:
  - name: https
    protocol: TCP
    port: 34571
    targetPort: 8443
  - name: metrics
    protocol: TCP
    port: 9220
    targetPort: 8001
`

func GetCSIDeploymentYAML(deploymentName, tridentImage,
	autosupportImage, autosupportProxy, autosupportCustomURL, autosupportSerialNumber, autosupportHostname,
	imageRegistry, logFormat string, imagePullSecrets []string, labels, controllingCRDetails map[string]string,
	debug, useIPv6, silenceAutosupport bool, version *utils.Version, topologyEnabled bool) string {

	var debugLine, logLevel, ipLocalhost string

	if debug {
		debugLine = "- -debug"
		logLevel = "9"
	} else {
		debugLine = "#- -debug"
		logLevel = "2"
	}
	if useIPv6 {
		ipLocalhost = "[::1]"
	} else {
		ipLocalhost = "127.0.0.1"
	}

	var deploymentYAML string
	isGCRRegistryVersion := false
	switch version.MinorVersion() {
	case 13:
		deploymentYAML = csiDeployment113YAMLTemplate
	case 14, 15:
		deploymentYAML = csiDeployment114YAMLTemplate
	case 16:
		deploymentYAML = csiDeployment116YAMLTemplate
	case 17, 18, 19:
		deploymentYAML = csiDeployment117YAMLTemplate
		isGCRRegistryVersion = true
	case 20:
		fallthrough
	default:
		deploymentYAML = csiDeployment120YAMLTemplate
		isGCRRegistryVersion = true
	}

	imageRegistry = getRegistryVal(imageRegistry, isGCRRegistryVersion)

	if autosupportImage == "" {
		autosupportImage = commonconfig.DefaultAutosupportImage
	}

	autosupportProxyLine := ""
	if autosupportProxy != "" {
		autosupportProxyLine = fmt.Sprint("- -proxy-url=", autosupportProxy)
	}

	autosupportCustomURLLine := ""
	if autosupportCustomURL != "" {
		autosupportCustomURLLine = fmt.Sprint("- -custom-url=", autosupportCustomURL)
	}

	autosupportSerialNumberLine := ""
	if autosupportSerialNumber != "" {
		autosupportSerialNumberLine = fmt.Sprint("- -serial-number=", autosupportSerialNumber)
	}

	autosupportHostnameLine := ""
	if autosupportHostname != "" {
		autosupportHostnameLine = fmt.Sprint("- -hostname=", autosupportHostname)
	}
	provisionerFeatureGates := ""
	if topologyEnabled {
		provisionerFeatureGates = "- --feature-gates=Topology=True"
	}

	if labels == nil {
		labels = make(map[string]string)
	}
	labels[DefaultContainerLabelKey] = "trident-main"

	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{TRIDENT_IMAGE}", tridentImage)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{DEPLOYMENT_NAME}", deploymentName)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{CSI_SIDECAR_REGISTRY}", imageRegistry)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{DEBUG}", debugLine)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{LABEL_APP}", labels[TridentAppLabelKey])
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{LOG_LEVEL}", logLevel)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{LOG_FORMAT}", logFormat)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{IP_LOCALHOST}", ipLocalhost)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{AUTOSUPPORT_IMAGE}", autosupportImage)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{AUTOSUPPORT_PROXY}", autosupportProxyLine)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{AUTOSUPPORT_CUSTOM_URL}", autosupportCustomURLLine)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{AUTOSUPPORT_SERIAL_NUMBER}", autosupportSerialNumberLine)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{AUTOSUPPORT_HOSTNAME}", autosupportHostnameLine)
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{AUTOSUPPORT_SILENCE}", strconv.FormatBool(silenceAutosupport))
	deploymentYAML = strings.ReplaceAll(deploymentYAML, "{PROVISIONER_FEATURE_GATES}", provisionerFeatureGates)
	deploymentYAML = replaceMultiline(deploymentYAML, labels, controllingCRDetails, imagePullSecrets)

	return deploymentYAML
}

const csiDeployment113YAMLTemplate = `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {DEPLOYMENT_NAME}
  {LABELS}
  {OWNER_REF}
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: {LABEL_APP}
  template:
    metadata:
      labels:
        app: {LABEL_APP}
    spec:
      serviceAccount: trident-csi
      containers:
      - name: trident-main
        image: {TRIDENT_IMAGE}
        ports:
        - containerPort: 8443
        - containerPort: 8001
        command:
        - /trident_orchestrator
        args:
        - "--crd_persistence"
        - "--k8s_pod"
        - "--https_rest"
        - "--https_port=8443"
        - "--csi_node_name=$(KUBE_NODE_NAME)"
        - "--csi_endpoint=$(CSI_ENDPOINT)"
        - "--csi_role=controller"
        - "--log_format={LOG_FORMAT}"
        - "--address={IP_LOCALHOST}"
        - "--metrics"
        {DEBUG}
        livenessProbe:
          exec:
            command:
            - tridentctl
            - -s
            - "{IP_LOCALHOST}:8000"
            - version
          failureThreshold: 2
          initialDelaySeconds: 120
          periodSeconds: 120
          timeoutSeconds: 90
        env:
        - name: KUBE_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CSI_ENDPOINT
          value: unix://plugin/csi.sock
        - name: TRIDENT_SERVER
          value: "{IP_LOCALHOST}:8000"
        volumeMounts:
        - name: socket-dir
          mountPath: /plugin
        - name: certs
          mountPath: /certs
          readOnly: true
      - name: trident-autosupport
        image: {AUTOSUPPORT_IMAGE}
        imagePullPolicy: Always
        command:
        - /usr/local/bin/trident-autosupport
        args:
        - "--k8s-pod"
        - "--log-format={LOG_FORMAT}"
        - "--trident-silence-collector={AUTOSUPPORT_SILENCE}"
        {AUTOSUPPORT_PROXY}
        {AUTOSUPPORT_CUSTOM_URL}
        {AUTOSUPPORT_SERIAL_NUMBER}
        {AUTOSUPPORT_HOSTNAME}
        {DEBUG}
        resources:
          limits:
            memory: 1Gi
        volumeMounts:
        - name: asup-dir
          mountPath: /asup
      - name: csi-provisioner
        image: {CSI_SIDECAR_REGISTRY}/csi-provisioner:v1.0.2
        args:
        - "--v={LOG_LEVEL}"
        - "--connection-timeout=24h"
        - "--csi-address=$(ADDRESS)"
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      - name: csi-attacher
        image: {CSI_SIDECAR_REGISTRY}/csi-attacher:v1.0.1
        args:
        - "--v={LOG_LEVEL}"
        - "--connection-timeout=24h"
        - "--timeout=60s"
        - "--csi-address=$(ADDRESS)"
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      - name: csi-cluster-driver-registrar
        image: {CSI_SIDECAR_REGISTRY}/csi-cluster-driver-registrar:v1.0.1
        args:
        - "--v={LOG_LEVEL}"
        - "--connection-timeout=24h"
        - "--csi-address=$(ADDRESS)"
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      {IMAGE_PULL_SECRETS}
      nodeSelector:
        beta.kubernetes.io/os: linux
        beta.kubernetes.io/arch: amd64
      volumes:
      - name: socket-dir
        emptyDir:
      - name: certs
        secret:
          secretName: trident-csi
      - name: asup-dir
        emptyDir:
          medium: ""
          sizeLimit: 1Gi
`

const csiDeployment114YAMLTemplate = `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {DEPLOYMENT_NAME}
  {LABELS}
  {OWNER_REF}
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: {LABEL_APP}
  template:
    metadata:
      labels:
        app: {LABEL_APP}
    spec:
      serviceAccount: trident-csi
      containers:
      - name: trident-main
        image: {TRIDENT_IMAGE}
        ports:
        - containerPort: 8443
        - containerPort: 8001
        command:
        - /trident_orchestrator
        args:
        - "--crd_persistence"
        - "--k8s_pod"
        - "--https_rest"
        - "--https_port=8443"
        - "--csi_node_name=$(KUBE_NODE_NAME)"
        - "--csi_endpoint=$(CSI_ENDPOINT)"
        - "--csi_role=controller"
        - "--log_format={LOG_FORMAT}"
        - "--address={IP_LOCALHOST}"
        - "--metrics"
        {DEBUG}
        livenessProbe:
          exec:
            command:
            - tridentctl
            - -s
            - "{IP_LOCALHOST}:8000"
            - version
          failureThreshold: 2
          initialDelaySeconds: 120
          periodSeconds: 120
          timeoutSeconds: 90
        env:
        - name: KUBE_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CSI_ENDPOINT
          value: unix://plugin/csi.sock
        - name: TRIDENT_SERVER
          value: "{IP_LOCALHOST}:8000"
        volumeMounts:
        - name: socket-dir
          mountPath: /plugin
        - name: certs
          mountPath: /certs
          readOnly: true
      - name: trident-autosupport
        image: {AUTOSUPPORT_IMAGE}
        imagePullPolicy: Always
        command:
        - /usr/local/bin/trident-autosupport
        args:
        - "--k8s-pod"
        - "--log-format={LOG_FORMAT}"
        - "--trident-silence-collector={AUTOSUPPORT_SILENCE}"
        {AUTOSUPPORT_PROXY}
        {AUTOSUPPORT_CUSTOM_URL}
        {AUTOSUPPORT_SERIAL_NUMBER}
        {AUTOSUPPORT_HOSTNAME}
        {DEBUG}
        resources:
          limits:
            memory: 1Gi
        volumeMounts:
        - name: asup-dir
          mountPath: /asup
      - name: csi-provisioner
        image: {CSI_SIDECAR_REGISTRY}/csi-provisioner:v1.6.1
        args:
        - "--v={LOG_LEVEL}"
        - "--timeout=600s"
        - "--csi-address=$(ADDRESS)"
        - "--retry-interval-start=8s"
        - "--retry-interval-max=30s"
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      - name: csi-attacher
        image: {CSI_SIDECAR_REGISTRY}/csi-attacher:v2.2.1
        args:
        - "--v={LOG_LEVEL}"
        - "--timeout=60s"
        - "--retry-interval-start=10s"
        - "--csi-address=$(ADDRESS)"
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      {IMAGE_PULL_SECRETS}
      nodeSelector:
        kubernetes.io/os: linux
        kubernetes.io/arch: amd64
      volumes:
      - name: socket-dir
        emptyDir:
      - name: certs
        secret:
          secretName: trident-csi
      - name: asup-dir
        emptyDir:
          medium: ""
          sizeLimit: 1Gi
`

const csiDeployment116YAMLTemplate = `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {DEPLOYMENT_NAME}
  {LABELS}
  {OWNER_REF}
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: {LABEL_APP}
  template:
    metadata:
      labels:
        app: {LABEL_APP}
    spec:
      serviceAccount: trident-csi
      containers:
      - name: trident-main
        image: {TRIDENT_IMAGE}
        ports:
        - containerPort: 8443
        - containerPort: 8001
        command:
        - /trident_orchestrator
        args:
        - "--crd_persistence"
        - "--k8s_pod"
        - "--https_rest"
        - "--https_port=8443"
        - "--csi_node_name=$(KUBE_NODE_NAME)"
        - "--csi_endpoint=$(CSI_ENDPOINT)"
        - "--csi_role=controller"
        - "--log_format={LOG_FORMAT}"
        - "--address={IP_LOCALHOST}"
        - "--metrics"
        {DEBUG}
        livenessProbe:
          exec:
            command:
            - tridentctl
            - -s
            - "{IP_LOCALHOST}:8000"
            - version
          failureThreshold: 2
          initialDelaySeconds: 120
          periodSeconds: 120
          timeoutSeconds: 90
        env:
        - name: KUBE_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CSI_ENDPOINT
          value: unix://plugin/csi.sock
        - name: TRIDENT_SERVER
          value: "{IP_LOCALHOST}:8000"
        volumeMounts:
        - name: socket-dir
          mountPath: /plugin
        - name: certs
          mountPath: /certs
          readOnly: true
      - name: trident-autosupport
        image: {AUTOSUPPORT_IMAGE}
        imagePullPolicy: Always
        command:
        - /usr/local/bin/trident-autosupport
        args:
        - "--k8s-pod"
        - "--log-format={LOG_FORMAT}"
        - "--trident-silence-collector={AUTOSUPPORT_SILENCE}"
        {AUTOSUPPORT_PROXY}
        {AUTOSUPPORT_CUSTOM_URL}
        {AUTOSUPPORT_SERIAL_NUMBER}
        {AUTOSUPPORT_HOSTNAME}
        {DEBUG}
        resources:
          limits:
            memory: 1Gi
        volumeMounts:
        - name: asup-dir
          mountPath: /asup
      - name: csi-provisioner
        image: {CSI_SIDECAR_REGISTRY}/csi-provisioner:v1.6.1
        args:
        - "--v={LOG_LEVEL}"
        - "--timeout=600s"
        - "--csi-address=$(ADDRESS)"
        - "--retry-interval-start=8s"
        - "--retry-interval-max=30s"
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      - name: csi-attacher
        image: {CSI_SIDECAR_REGISTRY}/csi-attacher:v2.2.1
        args:
        - "--v={LOG_LEVEL}"
        - "--timeout=60s"
        - "--retry-interval-start=10s"
        - "--csi-address=$(ADDRESS)"
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      - name: csi-resizer
        image: {CSI_SIDECAR_REGISTRY}/csi-resizer:v1.1.0
        args:
        - "--v={LOG_LEVEL}"
        - "--timeout=300s"
        - "--csi-address=$(ADDRESS)"
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      {IMAGE_PULL_SECRETS}
      nodeSelector:
        kubernetes.io/os: linux
        kubernetes.io/arch: amd64
      volumes:
      - name: socket-dir
        emptyDir:
      - name: certs
        secret:
          secretName: trident-csi
      - name: asup-dir
        emptyDir:
          medium: ""
          sizeLimit: 1Gi
`

const csiDeployment117YAMLTemplate = `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {DEPLOYMENT_NAME}
  {LABELS}
  {OWNER_REF}
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: {LABEL_APP}
  template:
    metadata:
      labels:
        app: {LABEL_APP}
    spec:
      serviceAccount: trident-csi
      containers:
      - name: trident-main
        image: {TRIDENT_IMAGE}
        ports:
        - containerPort: 8443
        - containerPort: 8001
        command:
        - /trident_orchestrator
        args:
        - "--crd_persistence"
        - "--k8s_pod"
        - "--https_rest"
        - "--https_port=8443"
        - "--csi_node_name=$(KUBE_NODE_NAME)"
        - "--csi_endpoint=$(CSI_ENDPOINT)"
        - "--csi_role=controller"
        - "--log_format={LOG_FORMAT}"
        - "--address={IP_LOCALHOST}"
        - "--metrics"
        {DEBUG}
        livenessProbe:
          exec:
            command:
            - tridentctl
            - -s
            - "{IP_LOCALHOST}:8000"
            - version
          failureThreshold: 2
          initialDelaySeconds: 120
          periodSeconds: 120
          timeoutSeconds: 90
        env:
        - name: KUBE_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CSI_ENDPOINT
          value: unix://plugin/csi.sock
        - name: TRIDENT_SERVER
          value: "{IP_LOCALHOST}:8000"
        volumeMounts:
        - name: socket-dir
          mountPath: /plugin
        - name: certs
          mountPath: /certs
          readOnly: true
      - name: trident-autosupport
        image: {AUTOSUPPORT_IMAGE}
        imagePullPolicy: Always
        command:
        - /usr/local/bin/trident-autosupport
        args:
        - "--k8s-pod"
        - "--log-format={LOG_FORMAT}"
        - "--trident-silence-collector={AUTOSUPPORT_SILENCE}"
        {AUTOSUPPORT_PROXY}
        {AUTOSUPPORT_CUSTOM_URL}
        {AUTOSUPPORT_SERIAL_NUMBER}
        {AUTOSUPPORT_HOSTNAME}
        {DEBUG}
        resources:
          limits:
            memory: 1Gi
        volumeMounts:
        - name: asup-dir
          mountPath: /asup
      - name: csi-provisioner
        image: {CSI_SIDECAR_REGISTRY}/csi-provisioner:v2.1.1
        args:
        - "--v={LOG_LEVEL}"
        - "--timeout=600s"
        - "--csi-address=$(ADDRESS)"
        - "--retry-interval-start=8s"
        - "--retry-interval-max=30s"
        {PROVISIONER_FEATURE_GATES}
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      - name: csi-attacher
        image: {CSI_SIDECAR_REGISTRY}/csi-attacher:v3.1.0
        args:
        - "--v={LOG_LEVEL}"
        - "--timeout=60s"
        - "--retry-interval-start=10s"
        - "--csi-address=$(ADDRESS)"
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      - name: csi-resizer
        image: {CSI_SIDECAR_REGISTRY}/csi-resizer:v1.1.0
        args:
        - "--v={LOG_LEVEL}"
        - "--timeout=300s"
        - "--csi-address=$(ADDRESS)"
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      - name: csi-snapshotter
        image: {CSI_SIDECAR_REGISTRY}/csi-snapshotter:v3.0.3
        args:
        - "--v={LOG_LEVEL}"
        - "--timeout=300s"
        - "--csi-address=$(ADDRESS)"
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      {IMAGE_PULL_SECRETS}
      nodeSelector:
        kubernetes.io/os: linux
        kubernetes.io/arch: amd64
      volumes:
      - name: socket-dir
        emptyDir:
      - name: certs
        secret:
          secretName: trident-csi
      - name: asup-dir
        emptyDir:
          medium: ""
          sizeLimit: 1Gi
`

const csiDeployment120YAMLTemplate = `---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {DEPLOYMENT_NAME}
  {LABELS}
  {OWNER_REF}
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: {LABEL_APP}
  template:
    metadata:
      labels:
        app: {LABEL_APP}
    spec:
      serviceAccount: trident-csi
      containers:
      - name: trident-main
        image: {TRIDENT_IMAGE}
        ports:
        - containerPort: 8443
        - containerPort: 8001
        command:
        - /trident_orchestrator
        args:
        - "--crd_persistence"
        - "--k8s_pod"
        - "--https_rest"
        - "--https_port=8443"
        - "--csi_node_name=$(KUBE_NODE_NAME)"
        - "--csi_endpoint=$(CSI_ENDPOINT)"
        - "--csi_role=controller"
        - "--log_format={LOG_FORMAT}"
        - "--address={IP_LOCALHOST}"
        - "--metrics"
        {DEBUG}
        livenessProbe:
          exec:
            command:
            - tridentctl
            - -s
            - "{IP_LOCALHOST}:8000"
            - version
          failureThreshold: 2
          initialDelaySeconds: 120
          periodSeconds: 120
          timeoutSeconds: 90
        env:
        - name: KUBE_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CSI_ENDPOINT
          value: unix://plugin/csi.sock
        - name: TRIDENT_SERVER
          value: "{IP_LOCALHOST}:8000"
        volumeMounts:
        - name: socket-dir
          mountPath: /plugin
        - name: certs
          mountPath: /certs
          readOnly: true
      - name: trident-autosupport
        image: {AUTOSUPPORT_IMAGE}
        imagePullPolicy: Always
        command:
        - /usr/local/bin/trident-autosupport
        args:
        - "--k8s-pod"
        - "--log-format={LOG_FORMAT}"
        - "--trident-silence-collector={AUTOSUPPORT_SILENCE}"
        {AUTOSUPPORT_PROXY}
        {AUTOSUPPORT_CUSTOM_URL}
        {AUTOSUPPORT_SERIAL_NUMBER}
        {AUTOSUPPORT_HOSTNAME}
        {DEBUG}
        resources:
          limits:
            memory: 1Gi
        volumeMounts:
        - name: asup-dir
          mountPath: /asup
      - name: csi-provisioner
        image: {CSI_SIDECAR_REGISTRY}/csi-provisioner:v2.1.1
        args:
        - "--v={LOG_LEVEL}"
        - "--timeout=600s"
        - "--csi-address=$(ADDRESS)"
        - "--retry-interval-start=8s"
        - "--retry-interval-max=30s"
        {PROVISIONER_FEATURE_GATES}
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      - name: csi-attacher
        image: {CSI_SIDECAR_REGISTRY}/csi-attacher:v3.1.0
        args:
        - "--v={LOG_LEVEL}"
        - "--timeout=60s"
        - "--retry-interval-start=10s"
        - "--csi-address=$(ADDRESS)"
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      - name: csi-resizer
        image: {CSI_SIDECAR_REGISTRY}/csi-resizer:v1.1.0
        args:
        - "--v={LOG_LEVEL}"
        - "--timeout=300s"
        - "--csi-address=$(ADDRESS)"
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      - name: csi-snapshotter
        image: {CSI_SIDECAR_REGISTRY}/csi-snapshotter:v3.0.3
        args:
        - "--v={LOG_LEVEL}"
        - "--timeout=300s"
        - "--csi-address=$(ADDRESS)"
        env:
        - name: ADDRESS
          value: /var/lib/csi/sockets/pluginproxy/csi.sock
        volumeMounts:
        - name: socket-dir
          mountPath: /var/lib/csi/sockets/pluginproxy/
      {IMAGE_PULL_SECRETS}
      nodeSelector:
        kubernetes.io/os: linux
        kubernetes.io/arch: amd64
      volumes:
      - name: socket-dir
        emptyDir:
      - name: certs
        secret:
          secretName: trident-csi
      - name: asup-dir
        emptyDir:
          medium: ""
          sizeLimit: 1Gi
`

func GetCSIDaemonSetYAML(daemonsetName, tridentImage, imageRegistry, kubeletDir, logFormat string,
	imagePullSecrets []string, labels, controllingCRDetails map[string]string, debug, nodePrep bool,
	version *utils.Version) string {

	var debugLine, logLevel string

	if debug {
		debugLine = "- -debug"
		logLevel = "9"
	} else {
		debugLine = "#- -debug"
		logLevel = "2"
	}

	isGCRRegistryVersion := true
	daemonSetYAML := daemonSet118YAMLTemplate
	if version.MajorVersion() == 1 {
		switch version.MinorVersion() {
		case 13:
			daemonSetYAML = daemonSet113YAMLTemplate
			isGCRRegistryVersion = false
		case 14, 15, 16:
			daemonSetYAML = daemonSet114YAMLTemplate
			isGCRRegistryVersion = false
		case 17:
			daemonSetYAML = daemonSet114YAMLTemplate
		}
	}

	imageRegistry = getRegistryVal(imageRegistry, isGCRRegistryVersion)

	labels[DefaultContainerLabelKey] = "trident-main"

	kubeletDir = strings.TrimRight(kubeletDir, "/")
	daemonSetYAML = strings.ReplaceAll(daemonSetYAML, "{TRIDENT_IMAGE}", tridentImage)
	daemonSetYAML = strings.ReplaceAll(daemonSetYAML, "{DAEMONSET_NAME}", daemonsetName)
	daemonSetYAML = strings.ReplaceAll(daemonSetYAML, "{CSI_SIDECAR_REGISTRY}", imageRegistry)
	daemonSetYAML = strings.ReplaceAll(daemonSetYAML, "{KUBELET_DIR}", kubeletDir)
	daemonSetYAML = strings.ReplaceAll(daemonSetYAML, "{LABEL_APP}", labels[TridentAppLabelKey])
	daemonSetYAML = strings.ReplaceAll(daemonSetYAML, "{DEBUG}", debugLine)
	daemonSetYAML = strings.ReplaceAll(daemonSetYAML, "{LOG_LEVEL}", logLevel)
	daemonSetYAML = strings.ReplaceAll(daemonSetYAML, "{LOG_FORMAT}", logFormat)
	daemonSetYAML = strings.ReplaceAll(daemonSetYAML, "{NODE_PREP}", strconv.FormatBool(nodePrep))
	daemonSetYAML = replaceMultiline(daemonSetYAML, labels, controllingCRDetails, imagePullSecrets)

	return daemonSetYAML
}

const daemonSet113YAMLTemplate = `---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: {DAEMONSET_NAME}
  {LABELS}
  {OWNER_REF}
spec:
  selector:
    matchLabels:
      app: {LABEL_APP}
  template:
    metadata:
      labels:
        app: {LABEL_APP}
    spec:
      serviceAccount: trident-csi
      hostNetwork: true
      hostIPC: true
      hostPID: true
      dnsPolicy: ClusterFirstWithHostNet
      containers:
      - name: trident-main
        securityContext:
          privileged: true
          capabilities:
            add: ["SYS_ADMIN"]
          allowPrivilegeEscalation: true
        image: {TRIDENT_IMAGE}
        command:
        - /trident_orchestrator
        args:
        - "--no_persistence"
        - "--rest=false"
        - "--csi_node_name=$(KUBE_NODE_NAME)"
        - "--csi_endpoint=$(CSI_ENDPOINT)"
        - "--csi_role=node"
        - "--log_format={LOG_FORMAT}"
        - "--node_prep={NODE_PREP}"
        - "--https_rest"
        - "--https_port=34572"
        {DEBUG}
        livenessProbe:
          httpGet:
            path: /liveness
            scheme: HTTPS
            port: 34572
          failureThreshold: 3
          timeoutSeconds: 1
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readiness
            scheme: HTTPS
            port: 34572
          failureThreshold: 5
          initialDelaySeconds: 10
          periodSeconds: 10
        env:
        - name: KUBE_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CSI_ENDPOINT
          value: unix://plugin/csi.sock
        - name: PATH
          value: /netapp:/bin
        volumeMounts:
        - name: plugin-dir
          mountPath: /plugin
        - name: plugins-mount-dir
          mountPath: {KUBELET_DIR}/plugins
        - name: pods-mount-dir
          mountPath: {KUBELET_DIR}/pods
          mountPropagation: "Bidirectional"
        - name: dev-dir
          mountPath: /dev
        - name: sys-dir
          mountPath: /sys
        - name: host-dir
          mountPath: /host
          mountPropagation: "Bidirectional"
        - name: trident-tracking-dir
          mountPath: /var/lib/trident/tracking
        - name: certs
          mountPath: /certs
          readOnly: true
      - name: driver-registrar
        image: {CSI_SIDECAR_REGISTRY}/csi-node-driver-registrar:v1.0.2
        args:
        - "--v={LOG_LEVEL}"
        - "--connection-timeout=24h"
        - "--csi-address=$(ADDRESS)"
        - "--kubelet-registration-path=$(REGISTRATION_PATH)"
        env:
        - name: ADDRESS
          value: /plugin/csi.sock
        - name: REGISTRATION_PATH
          value: "{KUBELET_DIR}/plugins/csi.trident.netapp.io/csi.sock"
        - name: KUBE_NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        volumeMounts:
        - name: plugin-dir
          mountPath: /plugin
        - name: registration-dir
          mountPath: /registration
      {IMAGE_PULL_SECRETS}
      nodeSelector:
        beta.kubernetes.io/os: linux
        beta.kubernetes.io/arch: amd64
      tolerations:
      - effect: NoExecute
        operator: Exists
      - effect: NoSchedule
        operator: Exists
      volumes:
      - name: plugin-dir
        hostPath:
          path: {KUBELET_DIR}/plugins/csi.trident.netapp.io/
          type: DirectoryOrCreate
      - name: registration-dir
        hostPath:
          path: {KUBELET_DIR}/plugins_registry/
          type: Directory
      - name: plugins-mount-dir
        hostPath:
          path: {KUBELET_DIR}/plugins
          type: DirectoryOrCreate
      - name: pods-mount-dir
        hostPath:
          path: {KUBELET_DIR}/pods
          type: DirectoryOrCreate
      - name: dev-dir
        hostPath:
          path: /dev
          type: Directory
      - name: sys-dir
        hostPath:
          path: /sys
          type: Directory
      - name: host-dir
        hostPath:
          path: /
          type: Directory
      - name: trident-tracking-dir
        hostPath:
          path: /var/lib/trident/tracking
          type: DirectoryOrCreate
      - name: certs
        secret:
          secretName: trident-csi
`

const daemonSet114YAMLTemplate = `---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: {DAEMONSET_NAME}
  {LABELS}
  {OWNER_REF}
spec:
  selector:
    matchLabels:
      app: {LABEL_APP}
  template:
    metadata:
      labels:
        app: {LABEL_APP}
    spec:
      serviceAccount: trident-csi
      hostNetwork: true
      hostIPC: true
      hostPID: true
      dnsPolicy: ClusterFirstWithHostNet
      containers:
      - name: trident-main
        securityContext:
          privileged: true
          capabilities:
            add: ["SYS_ADMIN"]
          allowPrivilegeEscalation: true
        image: {TRIDENT_IMAGE}
        command:
        - /trident_orchestrator
        args:
        - "--no_persistence"
        - "--rest=false"
        - "--csi_node_name=$(KUBE_NODE_NAME)"
        - "--csi_endpoint=$(CSI_ENDPOINT)"
        - "--csi_role=node"
        - "--log_format={LOG_FORMAT}"
        - "--node_prep={NODE_PREP}"
        - "--https_rest"
        - "--https_port=34572"
        {DEBUG}
        livenessProbe:
          httpGet:
            path: /liveness
            scheme: HTTPS
            port: 34572
          failureThreshold: 3
          timeoutSeconds: 1
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readiness
            scheme: HTTPS
            port: 34572
          failureThreshold: 5
          initialDelaySeconds: 10
          periodSeconds: 10
        env:
        - name: KUBE_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CSI_ENDPOINT
          value: unix://plugin/csi.sock
        - name: PATH
          value: /netapp:/bin
        volumeMounts:
        - name: plugin-dir
          mountPath: /plugin
        - name: plugins-mount-dir
          mountPath: {KUBELET_DIR}/plugins
        - name: pods-mount-dir
          mountPath: {KUBELET_DIR}/pods
          mountPropagation: "Bidirectional"
        - name: dev-dir
          mountPath: /dev
        - name: sys-dir
          mountPath: /sys
        - name: host-dir
          mountPath: /host
          mountPropagation: "Bidirectional"
        - name: trident-tracking-dir
          mountPath: /var/lib/trident/tracking
        - name: certs
          mountPath: /certs
          readOnly: true
      - name: driver-registrar
        image: {CSI_SIDECAR_REGISTRY}/csi-node-driver-registrar:v2.1.0
        args:
        - "--v={LOG_LEVEL}"
        - "--csi-address=$(ADDRESS)"
        - "--kubelet-registration-path=$(REGISTRATION_PATH)"
        env:
        - name: ADDRESS
          value: /plugin/csi.sock
        - name: REGISTRATION_PATH
          value: "{KUBELET_DIR}/plugins/csi.trident.netapp.io/csi.sock"
        - name: KUBE_NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        volumeMounts:
        - name: plugin-dir
          mountPath: /plugin
        - name: registration-dir
          mountPath: /registration
      {IMAGE_PULL_SECRETS}
      nodeSelector:
        kubernetes.io/os: linux
        kubernetes.io/arch: amd64
      tolerations:
      - effect: NoExecute
        operator: Exists
      - effect: NoSchedule
        operator: Exists
      volumes:
      - name: plugin-dir
        hostPath:
          path: {KUBELET_DIR}/plugins/csi.trident.netapp.io/
          type: DirectoryOrCreate
      - name: registration-dir
        hostPath:
          path: {KUBELET_DIR}/plugins_registry/
          type: Directory
      - name: plugins-mount-dir
        hostPath:
          path: {KUBELET_DIR}/plugins
          type: DirectoryOrCreate
      - name: pods-mount-dir
        hostPath:
          path: {KUBELET_DIR}/pods
          type: DirectoryOrCreate
      - name: dev-dir
        hostPath:
          path: /dev
          type: Directory
      - name: sys-dir
        hostPath:
          path: /sys
          type: Directory
      - name: host-dir
        hostPath:
          path: /
          type: Directory
      - name: trident-tracking-dir
        hostPath:
          path: /var/lib/trident/tracking
          type: DirectoryOrCreate
      - name: certs
        secret:
          secretName: trident-csi
`

const daemonSet118YAMLTemplate = `---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: {DAEMONSET_NAME}
  {LABELS}
  {OWNER_REF}
spec:
  selector:
    matchLabels:
      app: {LABEL_APP}
  template:
    metadata:
      labels:
        app: {LABEL_APP}
    spec:
      serviceAccount: trident-csi
      hostNetwork: true
      hostIPC: true
      hostPID: true
      dnsPolicy: ClusterFirstWithHostNet
      containers:
      - name: trident-main
        securityContext:
          privileged: true
          capabilities:
            add: ["SYS_ADMIN"]
          allowPrivilegeEscalation: true
        image: {TRIDENT_IMAGE}
        command:
        - /trident_orchestrator
        args:
        - "--no_persistence"
        - "--rest=false"
        - "--csi_node_name=$(KUBE_NODE_NAME)"
        - "--csi_endpoint=$(CSI_ENDPOINT)"
        - "--csi_role=node"
        - "--log_format={LOG_FORMAT}"
        - "--node_prep={NODE_PREP}"
        - "--https_rest"
        - "--https_port=34572"
        {DEBUG}
        startupProbe:
          httpGet:
            path: /liveness
            scheme: HTTPS
            port: 34572
          failureThreshold: 5
          timeoutSeconds: 1
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /liveness
            scheme: HTTPS
            port: 34572
          failureThreshold: 3
          timeoutSeconds: 1
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readiness
            scheme: HTTPS
            port: 34572
          failureThreshold: 5
          initialDelaySeconds: 10
          periodSeconds: 10
        env:
        - name: KUBE_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CSI_ENDPOINT
          value: unix://plugin/csi.sock
        - name: PATH
          value: /netapp:/bin
        volumeMounts:
        - name: plugin-dir
          mountPath: /plugin
        - name: plugins-mount-dir
          mountPath: {KUBELET_DIR}/plugins
        - name: pods-mount-dir
          mountPath: {KUBELET_DIR}/pods
          mountPropagation: "Bidirectional"
        - name: dev-dir
          mountPath: /dev
        - name: sys-dir
          mountPath: /sys
        - name: host-dir
          mountPath: /host
          mountPropagation: "Bidirectional"
        - name: trident-tracking-dir
          mountPath: /var/lib/trident/tracking
        - name: certs
          mountPath: /certs
          readOnly: true
      - name: driver-registrar
        image: {CSI_SIDECAR_REGISTRY}/csi-node-driver-registrar:v2.1.0
        args:
        - "--v={LOG_LEVEL}"
        - "--csi-address=$(ADDRESS)"
        - "--kubelet-registration-path=$(REGISTRATION_PATH)"
        env:
        - name: ADDRESS
          value: /plugin/csi.sock
        - name: REGISTRATION_PATH
          value: "{KUBELET_DIR}/plugins/csi.trident.netapp.io/csi.sock"
        - name: KUBE_NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        volumeMounts:
        - name: plugin-dir
          mountPath: /plugin
        - name: registration-dir
          mountPath: /registration
      {IMAGE_PULL_SECRETS}
      nodeSelector:
        kubernetes.io/os: linux
        kubernetes.io/arch: amd64
      tolerations:
      - effect: NoExecute
        operator: Exists
      - effect: NoSchedule
        operator: Exists
      volumes:
      - name: plugin-dir
        hostPath:
          path: {KUBELET_DIR}/plugins/csi.trident.netapp.io/
          type: DirectoryOrCreate
      - name: registration-dir
        hostPath:
          path: {KUBELET_DIR}/plugins_registry/
          type: Directory
      - name: plugins-mount-dir
        hostPath:
          path: {KUBELET_DIR}/plugins
          type: DirectoryOrCreate
      - name: pods-mount-dir
        hostPath:
          path: {KUBELET_DIR}/pods
          type: DirectoryOrCreate
      - name: dev-dir
        hostPath:
          path: /dev
          type: Directory
      - name: sys-dir
        hostPath:
          path: /sys
          type: Directory
      - name: host-dir
        hostPath:
          path: /
          type: Directory
      - name: trident-tracking-dir
        hostPath:
          path: /var/lib/trident/tracking
          type: DirectoryOrCreate
      - name: certs
        secret:
          secretName: trident-csi
`

func GetInstallerServiceAccountYAML() string {

	return installerServiceAccountYAML
}

const installerServiceAccountYAML = `---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: trident-installer
`

func GetInstallerClusterRoleYAML(flavor OrchestratorFlavor) string {
	switch flavor {
	case FlavorOpenShift:
		return installerClusterRoleOpenShiftYAML
	default:
		fallthrough
	case FlavorKubernetes:
		return installerClusterRoleKubernetesYAMLTemplate
	}
}

const installerClusterRoleOpenShiftYAML = `---
kind: ClusterRole
apiVersion: "authorization.openshift.io/v1"
metadata:
  name: trident-installer
rules:
  - apiGroups: [""]
    resources: ["namespaces", "pods", "pods/exec", "pods/log", "persistentvolumes", "persistentvolumeclaims", "persistentvolumeclaims/status", "secrets", "serviceaccounts", "services", "events", "nodes", "configmaps"]
    verbs: ["*"]
  - apiGroups: ["extensions"]
    resources: ["deployments", "daemonsets"]
    verbs: ["*"]
  - apiGroups: ["apps"]
    resources: ["statefulsets", daemonsets", "deployments"]
    verbs: ["*"]
  - apiGroups: ["authorization.openshift.io", "rbac.authorization.k8s.io"]
    resources: ["clusterroles", "clusterrolebindings"]
    verbs: ["*"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses", "volumeattachments", "volumeattachments/status"]
    verbs: ["*"]
  - apiGroups: ["metrics.k8s.io"]
    resources: ["*"]
    verbs: ["*"]
  - apiGroups: ["security.openshift.io"]
    resources: ["securitycontextconstraints"]
    verbs: ["*"]
  - apiGroups: ["apiextensions.k8s.io"]
    resources: ["customresourcedefinitions"]
    verbs: ["*"]
  - apiGroups: ["trident.netapp.io"]
    resources: ["tridentversions", "tridentbackends", "tridentstorageclasses", "tridentvolumes","tridentnodes",
"tridenttransactions", "tridentsnapshots", "tridentbackendconfigs", "tridentbackendconfigs/status"]
    verbs: ["*"]
  - apiGroups: ["policy"]
    resources: ["podsecuritypolicies"]
    verbs: ["*"]
`

const installerClusterRoleKubernetesYAMLTemplate = `---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: trident-installer
rules:
  - apiGroups: [""]
    resources: ["namespaces", "pods", "pods/exec", "pods/log", "persistentvolumes", "persistentvolumeclaims", "persistentvolumeclaims/status", "secrets", "serviceaccounts", "services", "events", "nodes", "configmaps"]
    verbs: ["*"]
  - apiGroups: ["extensions"]
    resources: ["deployments", "daemonsets"]
    verbs: ["*"]
  - apiGroups: ["apps"]
    resources: ["statefulsets", "daemonsets", "deployments"]
    verbs: ["*"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["clusterroles", "clusterrolebindings"]
    verbs: ["*"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses", "volumeattachments", "volumeattachments/status", csidrivers", "csinodes"]
    verbs: ["*"]
  - apiGroups: ["metrics.k8s.io"]
    resources: ["*"]
    verbs: ["*"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshots", "volumesnapshotclasses", "volumesnapshotcontents"]
    verbs: ["*"]
  - apiGroups: ["apiextensions.k8s.io"]
    resources: ["customresourcedefinitions"]
    verbs: ["*"]
  - apiGroups: ["csi.storage.k8s.io"]
    resources: ["csidrivers", "csinodeinfos"]
    verbs: ["*"]
  - apiGroups: ["trident.netapp.io"]
    resources: ["tridentversions", "tridentbackends", "tridentstorageclasses", "tridentvolumes","tridentnodes",
"tridenttransactions", "tridentsnapshots", "tridentbackendconfigs", "tridentbackendconfigs/status"]
    verbs: ["*"]
  - apiGroups: ["policy"]
    resources: ["podsecuritypolicies"]
    verbs: ["*"]
`

func GetInstallerClusterRoleBindingYAML(namespace string, flavor OrchestratorFlavor) string {

	var crbYAML string

	switch flavor {
	case FlavorOpenShift:
		crbYAML = installerClusterRoleBindingOpenShiftYAMLTemplate
	default:
		fallthrough
	case FlavorKubernetes:
		crbYAML = installerClusterRoleBindingKubernetesV1YAMLTemplate
	}

	crbYAML = strings.ReplaceAll(crbYAML, "{NAMESPACE}", namespace)
	return crbYAML
}

const installerClusterRoleBindingOpenShiftYAMLTemplate = `---
kind: ClusterRoleBinding
apiVersion: authorization.openshift.io/v1
metadata:
  name: trident-installer
subjects:
  - kind: ServiceAccount
    name: trident-installer
    namespace: {NAMESPACE}
roleRef:
  name: trident-installer
`

const installerClusterRoleBindingKubernetesV1YAMLTemplate = `---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: trident-installer
subjects:
  - kind: ServiceAccount
    name: trident-installer
    namespace: {NAMESPACE}
roleRef:
  kind: ClusterRole
  name: trident-installer
  apiGroup: rbac.authorization.k8s.io
`

func GetInstallerPodYAML(label, tridentImage string, commandArgs []string) string {

	command := `["` + strings.Join(commandArgs, `", "`) + `"]`

	jobYAML := strings.ReplaceAll(installerPodTemplate, "{LABEL_APP}", label)
	jobYAML = strings.ReplaceAll(jobYAML, "{TRIDENT_IMAGE}", tridentImage)
	jobYAML = strings.ReplaceAll(jobYAML, "{COMMAND}", command)
	return jobYAML
}

const installerPodTemplate = `---
apiVersion: v1
kind: Pod
metadata:
  name: trident-installer
  labels:
    app: {LABEL_APP}
spec:
  serviceAccount: trident-installer
  containers:
  - name: trident-installer
    image: {TRIDENT_IMAGE}
    workingDir: /
    command: {COMMAND}
    volumeMounts:
    - name: setup-dir
      mountPath: /setup
  restartPolicy: Never
  nodeSelector:
    beta.kubernetes.io/os: linux
    beta.kubernetes.io/arch: amd64
  volumes:
  - name: setup-dir
    configMap:
      name: trident-installer
`

func GetUninstallerPodYAML(label, tridentImage string, commandArgs []string) string {

	command := `["` + strings.Join(commandArgs, `", "`) + `"]`

	jobYAML := strings.ReplaceAll(uninstallerPodTemplate, "{LABEL_APP}", label)
	jobYAML = strings.ReplaceAll(jobYAML, "{TRIDENT_IMAGE}", tridentImage)
	jobYAML = strings.ReplaceAll(jobYAML, "{COMMAND}", command)
	return jobYAML
}

const uninstallerPodTemplate = `---
apiVersion: v1
kind: Pod
metadata:
  name: trident-installer
  labels:
    app: {LABEL_APP}
spec:
  serviceAccount: trident-installer
  containers:
  - name: trident-installer
    image: {TRIDENT_IMAGE}
    workingDir: /
    command: {COMMAND}
  nodeSelector:
    beta.kubernetes.io/os: linux
    beta.kubernetes.io/arch: amd64
  restartPolicy: Never
`

func GetTridentVersionPodYAML(name, tridentImage, serviceAccountName string, imagePullSecrets []string, labels,
	controllingCRDetails map[string]string) string {

	versionPodYAML := strings.ReplaceAll(tridentVersionPodYAML, "{NAME}", name)
	versionPodYAML = strings.ReplaceAll(versionPodYAML, "{TRIDENT_IMAGE}", tridentImage)
	versionPodYAML = strings.ReplaceAll(versionPodYAML, "{SERVICE_ACCOUNT}", serviceAccountName)
	versionPodYAML = replaceMultiline(versionPodYAML, labels, controllingCRDetails, imagePullSecrets)

	return versionPodYAML
}

const tridentVersionPodYAML = `---
apiVersion: v1
kind: Pod
metadata:
  name: {NAME}
  {LABELS}
  {OWNER_REF}
spec:
  serviceAccount: {SERVICE_ACCOUNT}
  restartPolicy: Never
  containers:
  - name: trident-main
    imagePullPolicy: IfNotPresent
    image: {TRIDENT_IMAGE}
    command: ["tridentctl"]
    args: ["pause"]
  {IMAGE_PULL_SECRETS}
  nodeSelector:
    beta.kubernetes.io/os: linux
    beta.kubernetes.io/arch: amd64
`

func GetEmptyConfigMapYAML(label, name, namespace string) string {

	configMapYAML := emptyConfigMapTemplate

	configMapYAML = strings.ReplaceAll(configMapYAML, "{LABEL_APP}", label)
	configMapYAML = strings.ReplaceAll(configMapYAML, "{NAMESPACE}", namespace)
	configMapYAML = strings.ReplaceAll(configMapYAML, "{NAME}", name)
	return configMapYAML
}

const emptyConfigMapTemplate = `---
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app: {LABEL_APP}
  name: {NAME}
  namespace: {NAMESPACE}
`

func GetOpenShiftSCCYAML(sccName, user, namespace string, labels, controllingCRDetails map[string]string) string {
	sccYAML := openShiftPrivilegedSCCYAML
	if !strings.Contains(labels[TridentAppLabelKey], "csi") && user != "trident-installer" {
		sccYAML = openShiftUnprivilegedSCCYAML
	}
	sccYAML = strings.ReplaceAll(sccYAML, "{SCC}", sccName)
	sccYAML = strings.ReplaceAll(sccYAML, "{NAMESPACE}", namespace)
	sccYAML = strings.ReplaceAll(sccYAML, "{USER}", user)
	sccYAML = replaceMultiline(sccYAML, labels, controllingCRDetails, nil)
	return sccYAML
}

const openShiftPrivilegedSCCYAML = `
apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  annotations:
    kubernetes.io/description: '{SCC} is a clone of the privileged built-in, and is meant just for use with trident.'
  name: {SCC}
  {LABELS}
  {OWNER_REF}
allowHostDirVolumePlugin: true
allowHostIPC: true
allowHostNetwork: true
allowHostPID: true
allowHostPorts: true
allowPrivilegeEscalation: true
allowPrivilegedContainer: true
allowedCapabilities:
- '*'
allowedUnsafeSysctls:
- '*'
defaultAddCapabilities: null
fsGroup:
  type: RunAsAny
groups: []
priority: null
readOnlyRootFilesystem: false
requiredDropCapabilities: null
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: RunAsAny
seccompProfiles:
- '*'
supplementalGroups:
  type: RunAsAny
users:
- system:serviceaccount:{NAMESPACE}:{USER}
volumes:
- '*'
`

const openShiftUnprivilegedSCCYAML = `
apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  annotations:
    kubernetes.io/description: '{SCC} is a clone of the anyuid built-in, and is meant just for use with trident.'
  name: {SCC}
  {LABELS}
  {OWNER_REF}
allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowPrivilegeEscalation: true
allowPrivilegedContainer: false
allowedCapabilities: null
apiVersion: security.openshift.io/v1
defaultAddCapabilities: null
fsGroup:
  type: RunAsAny
groups: []
priority: 10
readOnlyRootFilesystem: false
requiredDropCapabilities:
- MKNOD
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: RunAsAny
users:
- system:serviceaccount:{NAMESPACE}:{USER}
volumes:
- configMap
- downwardAPI
- emptyDir
- persistentVolumeClaim
- projected
- secret
`

func GetOpenShiftSCCQueryYAML(scc string) string {
	return strings.ReplaceAll(openShiftSCCQueryYAMLTemplate, "{SCC}", scc)
}

const openShiftSCCQueryYAMLTemplate = `
kind: SecurityContextConstraints
apiVersion: security.openshift.io/v1
metadata:
  name: {SCC}
`

func GetSecretYAML(secretName, namespace string, labels, controllingCRDetails, data, stringData map[string]string) string {

	secretYAML := strings.ReplaceAll(secretYAMLTemplate, "{SECRET_NAME}", secretName)
	secretYAML = strings.ReplaceAll(secretYAML, "{NAMESPACE}", namespace)
	secretYAML = replaceMultiline(secretYAML, labels, controllingCRDetails, nil)

	if data != nil {
		secretYAML += "data:\n"
		for key, value := range data {
			secretYAML += fmt.Sprintf("  %s: %s\n", key, value)
		}
	}

	if stringData != nil {
		secretYAML += "stringData:\n"
		for key, value := range stringData {
			secretYAML += fmt.Sprintf("  %s: %s\n", key, value)
		}
	}

	return secretYAML
}

const secretYAMLTemplate = `
apiVersion: v1
kind: Secret
metadata:
  name: {SECRET_NAME}
  namespace: {NAMESPACE}
  {LABELS}
  {OWNER_REF}
`

func GetCRDsYAML(useCRDv1 bool) string {
	if useCRDv1 {
		return customResourceDefinitionYAML_v1
	} else {
		return customResourceDefinitionYAML_v1beta1
	}
}

func GetVersionCRDYAML(useCRDv1 bool) string {
	if useCRDv1 {
		return tridentVersionCRDYAML_v1
	} else {
		return tridentVersionCRDYAML_v1beta1
	}
}

func GetBackendCRDYAML(useCRDv1 bool) string {
	if useCRDv1 {
		return tridentBackendCRDYAML_v1
	} else {
		return tridentBackendCRDYAML_v1beta1
	}
}

func GetBackendConfigCRDYAML(useCRDv1 bool) string {
	if useCRDv1 {
		return tridentBackendConfigCRDYAML_v1
	} else {
		return tridentBackendConfigCRDYAML_v1beta1
	}
}

func GetStorageClassCRDYAML(useCRDv1 bool) string {
	if useCRDv1 {
		return tridentStorageClassCRDYAML_v1
	} else {
		return tridentStorageClassCRDYAML_v1beta1
	}
}

func GetVolumeCRDYAML(useCRDv1 bool) string {
	if useCRDv1 {
		return tridentVolumeCRDYAML_v1
	} else {
		return tridentVolumeCRDYAML_v1beta1
	}
}

func GetNodeCRDYAML(useCRDv1 bool) string {
	if useCRDv1 {
		return tridentNodeCRDYAML_v1
	} else {
		return tridentNodeCRDYAML_v1beta1
	}
}

func GetTransactionCRDYAML(useCRDv1 bool) string {
	if useCRDv1 {
		return tridentTransactionCRDYAML_v1
	} else {
		return tridentTransactionCRDYAML_v1beta1
	}
}

func GetSnapshotCRDYAML(useCRDv1 bool) string {
	if useCRDv1 {
		return tridentSnapshotCRDYAML_v1
	} else {
		return tridentSnapshotCRDYAML_v1beta1
	}
}

func GetOrchestratorCRDYAML(useCRDv1 bool) string {
	if useCRDv1 {
		return tridentOrchestratorCRDYAML_v1
	} else {
		return tridentOrchestratorCRDYAML_v1beta1
	}
}

/*
kubectl delete crd tridentversions.trident.netapp.io --wait=false
kubectl delete crd tridentbackends.trident.netapp.io --wait=false
kubectl delete crd tridentbackendconfigs.trident.netapp.io --wait=false
kubectl delete crd tridentstorageclasses.trident.netapp.io --wait=false
kubectl delete crd tridentvolumes.trident.netapp.io --wait=false
kubectl delete crd tridentnodes.trident.netapp.io --wait=false
kubectl delete crd tridenttransactions.trident.netapp.io --wait=false
kubectl delete crd tridentsnapshots.trident.netapp.io --wait=false

kubectl patch crd tridentversions.trident.netapp.io -p '{"metadata":{"finalizers": []}}' --type=merge
kubectl patch crd tridentbackends.trident.netapp.io -p '{"metadata":{"finalizers": []}}' --type=merge
kubectl patch crd tridentbackendconfigs.trident.netapp.io -p '{"metadata":{"finalizers": []}}' --type=merge
kubectl patch crd tridentstorageclasses.trident.netapp.io -p '{"metadata":{"finalizers": []}}' --type=merge
kubectl patch crd tridentvolumes.trident.netapp.io -p '{"metadata":{"finalizers": []}}' --type=merge
kubectl patch crd tridentnodes.trident.netapp.io -p '{"metadata":{"finalizers": []}}' --type=merge
kubectl patch crd tridenttransactions.trident.netapp.io -p '{"metadata":{"finalizers": []}}' --type=merge
kubectl patch crd tridentsnapshots.trident.netapp.io -p '{"metadata":{"finalizers": []}}' --type=merge

kubectl delete crd tridentversions.trident.netapp.io
kubectl delete crd tridentbackends.trident.netapp.io
kubectl delete crd tridentbackendconfigs.trident.netapp.io
kubectl delete crd tridentstorageclasses.trident.netapp.io
kubectl delete crd tridentvolumes.trident.netapp.io
kubectl delete crd tridentnodes.trident.netapp.io
kubectl delete crd tridenttransactions.trident.netapp.io
kubectl delete crd tridentsnapshots.trident.netapp.io
*/

const tridentVersionCRDYAML_v1beta1 = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: tridentversions.trident.netapp.io
spec:
  group: trident.netapp.io
  version: v1
  versions:
    - name: v1
      served: true
      storage: true
  scope: Namespaced
  names:
    plural: tridentversions
    singular: tridentversion
    kind: TridentVersion
    shortNames:
    - tver
    - tversion
    categories:
    - trident
    - trident-internal
  additionalPrinterColumns:
    - name: Version
      type: string
      description: The Trident version
      priority: 0
      JSONPath: .trident_version`

const tridentBackendCRDYAML_v1beta1 = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: tridentbackends.trident.netapp.io
spec:
  group: trident.netapp.io
  version: v1
  versions:
    - name: v1
      served: true
      storage: true
  scope: Namespaced
  names:
    plural: tridentbackends
    singular: tridentbackend
    kind: TridentBackend
    shortNames:
    - tbe
    - tbackend
    categories:
    - trident
    - trident-internal
  additionalPrinterColumns:
    - name: Backend
      type: string
      description: The backend name
      priority: 0
      JSONPath: .backendName
    - name: Backend UUID
      type: string
      description: The backend UUID
      priority: 0
      JSONPath: .backendUUID`

const tridentBackendConfigCRDYAML_v1beta1 = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: tridentbackendconfigs.trident.netapp.io
spec:
  group: trident.netapp.io
  version: v1
  versions:
    - name: v1
      served: true
      storage: true
  scope: Namespaced
  subresources:
    status: {}
  names:
    plural: tridentbackendconfigs
    singular: tridentbackendconfig
    kind: TridentBackendConfig
    shortNames:
    - tbc
    - tbconfig
    - tbackendconfig
    categories:
    - trident
    - trident-internal
    - trident-external
  additionalPrinterColumns:
    - name: Backend Name
      type: string
      description: The backend name
      priority: 0
      JSONPath: .status.backendInfo.backendName
    - name: Backend UUID
      type: string
      description: The backend UUID
      priority: 0
      JSONPath: .status.backendInfo.backendUUID
    - name: Phase
      type: string
      description: The backend config phase
      priority: 0
      JSONPath: .status.phase
    - name: Status
      type: string
      description: The result of the last operation
      priority: 0
      JSONPath: .status.lastOperationStatus
    - name: Storage Driver
      type: string
      description: The storage driver type
      priority: 1
      JSONPath: .spec.storageDriverName
    - name: Deletion Policy
      type: string
      description: The deletion policy
      priority: 1
      JSONPath: .status.deletionPolicy`

const tridentStorageClassCRDYAML_v1beta1 = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: tridentstorageclasses.trident.netapp.io
spec:
  group: trident.netapp.io
  version: v1
  versions:
    - name: v1
      served: true
      storage: true
  scope: Namespaced
  names:
    plural: tridentstorageclasses
    singular: tridentstorageclass
    kind: TridentStorageClass
    shortNames:
    - tsc
    - tstorageclass
    categories:
    - trident
    - trident-internal`

const tridentVolumeCRDYAML_v1beta1 = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: tridentvolumes.trident.netapp.io
spec:
  group: trident.netapp.io
  version: v1
  versions:
    - name: v1
      served: true
      storage: true
  scope: Namespaced
  names:
    plural: tridentvolumes
    singular: tridentvolume
    kind: TridentVolume
    shortNames:
    - tvol
    - tvolume
    categories:
    - trident
    - trident-internal
  additionalPrinterColumns:
    - name: Age
      type: date
      priority: 0
      JSONPath: .metadata.creationTimestamp
    - name: Size
      type: string
      description: The volume's size
      priority: 1
      JSONPath: .config.size
    - name: Storage Class
      type: string
      description: The volume's storage class
      priority: 1
      JSONPath: .config.storageClass
    - name: State
      type: string
      description: The volume's state
      priority: 1
      JSONPath: .state
    - name: Protocol
      type: string
      description: The volume's protocol
      priority: 1
      JSONPath: .config.protocol
    - name: Backend UUID
      type: string
      description: The volume's backend UUID
      priority: 1
      JSONPath: .backendUUID
    - name: Pool
      type: string
      description: The volume's pool
      priority: 1
      JSONPath: .pool`

const tridentNodeCRDYAML_v1beta1 = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: tridentnodes.trident.netapp.io
spec:
  group: trident.netapp.io
  version: v1
  versions:
    - name: v1
      served: true
      storage: true
  scope: Namespaced
  names:
    plural: tridentnodes
    singular: tridentnode
    kind: TridentNode
    shortNames:
    - tnode
    categories:
    - trident
    - trident-internal`

const tridentTransactionCRDYAML_v1beta1 = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: tridenttransactions.trident.netapp.io
spec:
  group: trident.netapp.io
  version: v1
  versions:
    - name: v1
      served: true
      storage: true
  scope: Namespaced
  names:
    plural: tridenttransactions
    singular: tridenttransaction
    kind: TridentTransaction
    shortNames:
    - ttx
    - ttransaction
    categories:
    - trident-internal`

const tridentSnapshotCRDYAML_v1beta1 = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: tridentsnapshots.trident.netapp.io
spec:
  group: trident.netapp.io
  version: v1
  versions:
    - name: v1
      served: true
      storage: true
  scope: Namespaced
  names:
    plural: tridentsnapshots
    singular: tridentsnapshot
    kind: TridentSnapshot
    shortNames:
    - tss
    - tsnap
    - tsnapshot
    categories:
    - trident
    - trident-internal
  additionalPrinterColumns:
    - name: State
      type: string
      description: The snapshot's state
      priority: 1
      JSONPath: .state`

const tridentOrchestratorCRDYAML_v1beta1 = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: tridentorchestrators.trident.netapp.io
spec:
  group: trident.netapp.io
  names:
    kind: TridentOrchestrator
    listKind: TridentOrchestratorList
    plural: tridentorchestrators
    singular: tridentorchestrator
    shortNames:
    - torc
    - torchestrator
  scope: Cluster
  subresources:
    status: {}
  version: v1
  versions:
  - name: v1
    served: true
    storage: true`

const customResourceDefinitionYAML_v1beta1 = tridentVersionCRDYAML_v1beta1 +
	"\n---" + tridentBackendCRDYAML_v1beta1 +
	"\n---" + tridentBackendConfigCRDYAML_v1beta1 +
	"\n---" + tridentStorageClassCRDYAML_v1beta1 +
	"\n---" + tridentVolumeCRDYAML_v1beta1 +
	"\n---" + tridentNodeCRDYAML_v1beta1 +
	"\n---" + tridentTransactionCRDYAML_v1beta1 +
	"\n---" + tridentSnapshotCRDYAML_v1beta1

const tridentVersionCRDYAML_v1 = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tridentversions.trident.netapp.io
spec:
  group: trident.netapp.io
  versions:
    - name: v1
      served: true
      storage: true
      schema:
          openAPIV3Schema:
              type: object
              x-kubernetes-preserve-unknown-fields: true
      additionalPrinterColumns:
      - name: Version
        type: string
        description: The Trident version
        priority: 0
        jsonPath: .trident_version
  scope: Namespaced
  names:
    plural: tridentversions
    singular: tridentversion
    kind: TridentVersion
    shortNames:
    - tver
    - tversion
    categories:
    - trident
    - trident-internal`

const tridentBackendCRDYAML_v1 = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tridentbackends.trident.netapp.io
spec:
  group: trident.netapp.io
  versions:
    - name: v1
      served: true
      storage: true
      schema:
          openAPIV3Schema:
              type: object
              x-kubernetes-preserve-unknown-fields: true
      additionalPrinterColumns:
      - name: Backend
        type: string
        description: The backend name
        priority: 0
        jsonPath: .backendName
      - name: Backend UUID
        type: string
        description: The backend UUID
        priority: 0
        jsonPath: .backendUUID
  scope: Namespaced
  names:
    plural: tridentbackends
    singular: tridentbackend
    kind: TridentBackend
    shortNames:
    - tbe
    - tbackend
    categories:
    - trident
    - trident-internal`

const tridentBackendConfigCRDYAML_v1 = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tridentbackendconfigs.trident.netapp.io
spec:
  group: trident.netapp.io
  versions:
    - name: v1
      served: true
      storage: true
      schema:
          openAPIV3Schema:
              type: object
              x-kubernetes-preserve-unknown-fields: true
      subresources:
        status: {}
      additionalPrinterColumns:
      - name: Backend Name
        type: string
        description: The backend name
        priority: 0
        jsonPath: .status.backendInfo.backendName
      - name: Backend UUID
        type: string
        description: The backend UUID
        priority: 0
        jsonPath: .status.backendInfo.backendUUID
      - name: Phase
        type: string
        description: The backend config phase
        priority: 0
        jsonPath: .status.phase
      - name: Status
        type: string
        description: The result of the last operation
        priority: 0
        jsonPath: .status.lastOperationStatus
      - name: Storage Driver
        type: string
        description: The storage driver type
        priority: 1
        jsonPath: .spec.storageDriverName
      - name: Deletion Policy
        type: string
        description: The deletion policy
        priority: 1
        jsonPath: .status.deletionPolicy
  scope: Namespaced
  names:
    plural: tridentbackendconfigs
    singular: tridentbackendconfig
    kind: TridentBackendConfig
    shortNames:
    - tbc
    - tbconfig
    - tbackendconfig
    categories:
    - trident
    - trident-internal
    - trident-external`

const tridentStorageClassCRDYAML_v1 = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tridentstorageclasses.trident.netapp.io
spec:
  group: trident.netapp.io
  versions:
    - name: v1
      served: true
      storage: true
      schema:
          openAPIV3Schema:
              type: object
              x-kubernetes-preserve-unknown-fields: true
  scope: Namespaced
  names:
    plural: tridentstorageclasses
    singular: tridentstorageclass
    kind: TridentStorageClass
    shortNames:
    - tsc
    - tstorageclass
    categories:
    - trident
    - trident-internal`

const tridentVolumeCRDYAML_v1 = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tridentvolumes.trident.netapp.io
spec:
  group: trident.netapp.io
  versions:
    - name: v1
      served: true
      storage: true
      schema:
          openAPIV3Schema:
              type: object
              x-kubernetes-preserve-unknown-fields: true
      additionalPrinterColumns:
      - name: Age
        type: date
        priority: 0
        jsonPath: .metadata.creationTimestamp
      - name: Size
        type: string
        description: The volume's size
        priority: 1
        jsonPath: .config.size
      - name: Storage Class
        type: string
        description: The volume's storage class
        priority: 1
        jsonPath: .config.storageClass
      - name: State
        type: string
        description: The volume's state
        priority: 1
        jsonPath: .state
      - name: Protocol
        type: string
        description: The volume's protocol
        priority: 1
        jsonPath: .config.protocol
      - name: Backend UUID
        type: string
        description: The volume's backend UUID
        priority: 1
        jsonPath: .backendUUID
      - name: Pool
        type: string
        description: The volume's pool
        priority: 1
        jsonPath: .pool
  scope: Namespaced
  names:
    plural: tridentvolumes
    singular: tridentvolume
    kind: TridentVolume
    shortNames:
    - tvol
    - tvolume
    categories:
    - trident
    - trident-internal`

const tridentNodeCRDYAML_v1 = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tridentnodes.trident.netapp.io
spec:
  group: trident.netapp.io
  versions:
    - name: v1
      served: true
      storage: true
      schema:
          openAPIV3Schema:
              type: object
              x-kubernetes-preserve-unknown-fields: true
  scope: Namespaced
  names:
    plural: tridentnodes
    singular: tridentnode
    kind: TridentNode
    shortNames:
    - tnode
    categories:
    - trident
    - trident-internal`

const tridentTransactionCRDYAML_v1 = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tridenttransactions.trident.netapp.io
spec:
  group: trident.netapp.io
  versions:
    - name: v1
      served: true
      storage: true
      schema:
          openAPIV3Schema:
              type: object
              x-kubernetes-preserve-unknown-fields: true
  scope: Namespaced
  names:
    plural: tridenttransactions
    singular: tridenttransaction
    kind: TridentTransaction
    shortNames:
    - ttx
    - ttransaction
    categories:
    - trident-internal`

const tridentSnapshotCRDYAML_v1 = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tridentsnapshots.trident.netapp.io
spec:
  group: trident.netapp.io
  versions:
    - name: v1
      served: true
      storage: true
      schema:
          openAPIV3Schema:
              type: object
              x-kubernetes-preserve-unknown-fields: true
      additionalPrinterColumns:
      - name: State
        type: string
        description: The snapshot's state
        priority: 1
        jsonPath: .state
  scope: Namespaced
  names:
    plural: tridentsnapshots
    singular: tridentsnapshot
    kind: TridentSnapshot
    shortNames:
    - tss
    - tsnap
    - tsnapshot
    categories:
    - trident
    - trident-internal`

const tridentOrchestratorCRDYAML_v1 = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tridentorchestrators.trident.netapp.io
spec:
  group: trident.netapp.io
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          x-kubernetes-preserve-unknown-fields: true
      subresources:
        status: {}
  names:
    kind: TridentOrchestrator
    listKind: TridentOrchestratorList
    plural: tridentorchestrators
    singular: tridentorchestrator
    shortNames:
    - torc
    - torchestrator
  scope: Cluster`

const customResourceDefinitionYAML_v1 = tridentVersionCRDYAML_v1 +
	"\n---" + tridentBackendCRDYAML_v1 +
	"\n---" + tridentBackendConfigCRDYAML_v1 +
	"\n---" + tridentStorageClassCRDYAML_v1 +
	"\n---" + tridentVolumeCRDYAML_v1 +
	"\n---" + tridentNodeCRDYAML_v1 +
	"\n---" + tridentTransactionCRDYAML_v1 +
	"\n---" + tridentSnapshotCRDYAML_v1 + "\n"

func GetCSIDriverCRDYAML() string {
	return CSIDriverCRDYAML
}

const CSIDriverCRDYAML = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: csidrivers.csi.storage.k8s.io
  labels:
    addonmanager.kubernetes.io/mode: Reconcile
spec:
  group: csi.storage.k8s.io
  names:
    kind: CSIDriver
    plural: csidrivers
  scope: Cluster
  validation:
    openAPIV3Schema:
      properties:
        spec:
          description: Specification of the CSI Driver.
          properties:
            attachRequired:
              description: Indicates this CSI volume driver requires an attach operation,
                and that Kubernetes should call attach and wait for any attach operation
                to complete before proceeding to mount.
              type: boolean
            podInfoOnMountVersion:
              description: Indicates this CSI volume driver requires additional pod
                information (like podName, podUID, etc.) during mount operations.
              type: string
  version: v1alpha1
`

func GetCSINodeInfoCRDYAML() string {
	return CSINodeInfoCRDYAML
}

const CSINodeInfoCRDYAML = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: csinodeinfos.csi.storage.k8s.io
  labels:
    addonmanager.kubernetes.io/mode: Reconcile
spec:
  group: csi.storage.k8s.io
  names:
    kind: CSINodeInfo
    plural: csinodeinfos
  scope: Cluster
  validation:
    openAPIV3Schema:
      properties:
        spec:
          description: Specification of CSINodeInfo
          properties:
            drivers:
              description: List of CSI drivers running on the node and their specs.
              type: array
              items:
                properties:
                  name:
                    description: The CSI driver that this object refers to.
                    type: string
                  nodeID:
                    description: The node from the driver point of view.
                    type: string
                  topologyKeys:
                    description: List of keys supported by the driver.
                    items:
                      type: string
                    type: array
        status:
          description: Status of CSINodeInfo
          properties:
            drivers:
              description: List of CSI drivers running on the node and their statuses.
              type: array
              items:
                properties:
                  name:
                    description: The CSI driver that this object refers to.
                    type: string
                  available:
                    description: Whether the CSI driver is installed.
                    type: boolean
                  volumePluginMechanism:
                    description: Indicates to external components the required mechanism
                      to use for any in-tree plugins replaced by this driver.
                    pattern: in-tree|csi
                    type: string
  version: v1alpha1
`

func GetCSIDriverCRYAML(name string, labels, controllingCRDetails map[string]string) string {

	CSIDriverCR := strings.ReplaceAll(CSIDriverCRYAML, "{NAME}", name)
	CSIDriverCR = replaceMultiline(CSIDriverCR, labels, controllingCRDetails, nil)
	return CSIDriverCR
}

const CSIDriverCRYAML = `
apiVersion: storage.k8s.io/v1beta1
kind: CSIDriver
metadata:
  name: {NAME}
  {LABELS}
  {OWNER_REF}
spec:
  attachRequired: true
`

func GetPrivilegedPodSecurityPolicyYAML(pspName string, labels, controllingCRDetails map[string]string) string {

	pspYAML := strings.ReplaceAll(PrivilegedPodSecurityPolicyYAML, "{PSP_NAME}", pspName)
	pspYAML = replaceMultiline(pspYAML, labels, controllingCRDetails, nil)
	return pspYAML
}

const PrivilegedPodSecurityPolicyYAML = `
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: {PSP_NAME}
  {LABELS}
  {OWNER_REF}
spec:
  privileged: true
  allowPrivilegeEscalation: true
  allowedCapabilities:
  - "SYS_ADMIN"
  hostIPC: true
  hostPID: true
  hostNetwork: true
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  runAsUser:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
  volumes:
  - '*'
`

func GetUnprivilegedPodSecurityPolicyYAML(pspName string, labels, controllingCRDetails map[string]string) string {

	pspYAML := strings.ReplaceAll(UnprivilegedPodSecurityPolicyYAML, "{PSP_NAME}", pspName)
	pspYAML = replaceMultiline(pspYAML, labels, controllingCRDetails, nil)
	return pspYAML
}

const UnprivilegedPodSecurityPolicyYAML = `
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: {PSP_NAME}
  {LABELS}
  {OWNER_REF}
spec:
  privileged: false
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  runAsUser:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
  volumes:
    - '*'
`

func GetInstallerSecurityPolicyYAML() string {
	return InstallerSecurityPolicyYAML
}

const InstallerSecurityPolicyYAML = `
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: tridentinstaller
spec:
  privileged: false
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  runAsUser:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
  volumes:
    - '*'
`

// replaceMultiline replaces tags with multiline indented YAML, to make sure it works properly:
// 1. It should be called last after all single line replacements have been made.
// 2. Use only spaces before the tag
// 3. No space(s) or any other special character (other than newline) should be there after the tag
func replaceMultiline(originalYAML string, labels, ownerRef map[string]string, imagePullSecrets []string) string {
	for {
		tagWithSpaces, tag, spaceCount := utils.GetYAMLTagWithSpaceCount(originalYAML)

		if tagWithSpaces == "" {
			break
		}

		switch tag {
		case "LABELS":
			originalYAML = strings.Replace(originalYAML, tagWithSpaces, contructLabels(labels, createSpaces(spaceCount)), 1)
		case "OWNER_REF":
			originalYAML = strings.Replace(originalYAML, tagWithSpaces, constructOwnerRef(ownerRef, createSpaces(spaceCount)), 1)
		case "IMAGE_PULL_SECRETS":
			originalYAML = strings.Replace(originalYAML, tagWithSpaces, constructImagePullSecrets(imagePullSecrets, createSpaces(spaceCount)), 1)
		default:
			fmt.Errorf("found an unsupported tag %s in the YAML", tag)
			return ""
		}
	}

	return originalYAML
}

func createSpaces(spaceCount int) string {
	return strings.Repeat(" ", spaceCount)
}

func contructLabels(labels map[string]string, spaces string) string {

	var labelData string

	if labels != nil {
		labelData += spaces + "labels:\n"
		for key, value := range labels {
			labelData += fmt.Sprintf(spaces+"  %s: %s\n", key, value)
		}
	}

	return labelData
}

func constructOwnerRef(ownerRef map[string]string, spaces string) string {

	var ownerRefData string
	if ownerRef != nil {
		isFirst := true
		ownerRefData += spaces + "ownerReferences:\n"
		for key, value := range ownerRef {
			if isFirst {
				ownerRefData += fmt.Sprintf(spaces+"- %s: %s\n", key, value)
				isFirst = false
			} else {
				ownerRefData += fmt.Sprintf(spaces+"  %s: %s\n", key, value)
			}
		}
	}

	return ownerRefData
}

func constructImagePullSecrets(imagePullSecrets []string, spaces string) string {

	var imagePullSecretsData string
	if len(imagePullSecrets) > 0 {
		imagePullSecretsData += spaces + "imagePullSecrets:\n"
		for _, value := range imagePullSecrets {
			imagePullSecretsData += fmt.Sprintf(spaces+"- name: %s\n", value)
		}
	}

	return imagePullSecretsData
}

func constructServiceAccountSecrets(serviceAccountSecrets []string) string {

	var serviceAccountSecretsData string
	if len(serviceAccountSecrets) > 0 {
		serviceAccountSecretsData += "secrets:\n"
		for _, value := range serviceAccountSecrets {
			serviceAccountSecretsData += fmt.Sprintf("- name: %s\n", value)
		}
	}

	return serviceAccountSecretsData
}

// getRegistryVal uses a custom registry as is, if one is provided, otherwise, depending on k8s
// version < 1.17 or 1.17+ use quay.io/k8scsi or k8s.gcr.io/sig-storage, respectively
func getRegistryVal(registry string, isGCRRegistryVersion bool) string {
	if registry != "" {
		return strings.TrimSuffix(registry, "/")
	} else if isGCRRegistryVersion {
		return commonconfig.KubernetesCSISidecarRegistry117Plus
	}
	return commonconfig.KubernetesCSISidecarRegistryPre117
}
