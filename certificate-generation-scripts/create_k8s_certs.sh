#!/bin/bash
# Kubernetes API Client Certificate and Access Script

# Last Modified: 10/31/2017
# Author manux.ullas@intel.com

# The script can used by a client of the Kubernetes API to:
# 1) Generate X509 Client Certs signed by the Root CA of the Kubernetes Master for authentication
# 2) Provision the desired access to resources in the Kubernetes configuration
# 3) Package the Kubernetes API's Server Certificate into a Java KeyStore (JKS) aka Trust Store
# 4) Package the certs from step 1 into another JKS aka Cert Store
# The JKS files generated are protected by a randomly generated passphrase. This is provided in a JKS_NextSteps file along with the output.

# Pre-requisites:
# Script should be executed:
# 1) On the Kubernetes Master Node
# 2) As root user
# 3) After the Kuberenetes cluster is up and running

# Dependencies:
# 1) openssl - for operations on PKCS12 bundles
# 2) cfssl and cfssljson - for cert generation and signing
# 3) kubectl for interacting with the Kuberenetes API
# 4) keytool for packaging certs into Java Keystore
# 5) ansible for automated copy of generated files

# Input Parameters:
# 1) -u* userName 
# 2) -s* "hostname1.mydomain.net,hostname2" - these entries will populated the Subject Alternative Names (SAN) field in the client cert
# 3) -k* /path/to/K8S_SERVER_CA_KEY - path to the Kubernetes Root CA Key
# 4) -c* /path/to/K8S_SERVER_CA_CERT - path to the Kubernetes Root CA Certificate
# 5) -a* "path/to/K8S_APIServer_CERT/" - path to the Kuberenetes API Server Certificate
# 6) -r* "resource1,resource2,resource3" - list of resources for which access is to be provisioned
# 7) -v* "verb1,verb2,verb3 - verbs denoting operations that are accessible (eg.get,list,delete,create)
# * - indicates mandatory arguments

# Output:
# Two keystore files: 
# 1) The Kubernetes API client certificate keypair is packaged into a <username>_k8s_client.jks 
# 2) Server certificate keypair is packaged into a <username>_k8s_trust.jks
# 3) Tenant Configuration JSON template for the keystores is also provided in a <username>_JKS_NextSteps file.
# 4) Nova Plugin Tenant Configuration JSON template.

#Example:
#./create_k8s_certs.sh -u ahub_tenant_1 -s "10.105.168.157" -c /etc/kubernetes/pki/ca.crt -k /etc/kubernetes/pki/ca.key \
# -a /etc/kubernetes/pki/apiserver.crt -r "geolocationcrds.isecl.intel.com,platformcrds.isecl.intel.com" \
# -v "get,list,delete,patch,deletecollection,create,update" -d "/home/" -f "true" -i "attestation-hub" \
# -p "/opt/attestation-hub/configuration" -e "true" -y "3" -t "default" -b "project:admin" -o "admin" \
# -g "admin" -l "http://10.105.167.206:35357/v3"

echo -e "Starting Kubernetes API Server Client Access Provision Workflow\n-------------------------------\n\n"

# Constants
RESPONSE_OK=0
RESPONSE_USERNAME_MISSING=-255
RESPONSE_CA_CERT_MISSING=-254
RESPONSE_CA_KEY_MISSING=-253
RESPONSE_SERVER_CERT_MISSING=-238
RESPONSE_OPENSSL_MISSING=-252
RESPONSE_KUBECTL_MISSING=-251
RESPONSE_RW_ACCESS_PKI=-250
RESPONSE_GENRSA_FAIL=-249
RESPONSE_GENCERTSIGNREQUEST_FAIL=-248
RESPONSE_SIGNCERT_FAIL=-247
RESPONSE_REGCREDK8S_FAIL=-246
RESPONSE_CREATECLUSTERROLE_FAIL=-245
RESPONSE_CREATEROLEBINDING_FAIL=-244
RESPONSE_WGET_FAIL=-243
RESPONSE_WGET_MISSING=-242
RESPONSE_SANS_MISSING=-239
ERR_KEYSTORE_EXIST=-241
ERR_TRUST_STORE_EXIST=-240
ERR_KEYTOOL_MISSING=-237
ERR_CFSSL_MISSING=-236
ERR_CFSSLJSON_MISSING=-235
RESPONSE_DIRECTORY_PATH_MISSING=-256
RESPONSE_DOCKER_FLAG_OTHER_VALUES=-258
RESPONSE_DOCKER_FIELDS_MISSING=-259
RESPONSE_ANSIBLE_MISSING=-260
RESPONSE_VM_WORKER_FLAG_OTHER_VALUES=-262
RESPONSE_OPENSTACK_FIELDS_MISSING=-263

RESPONSE_MESSAGE_OPENSSL_MISSING="Error: OpenSSL is missing! Please install and try again."
RESPONSE_MESSAGE_KUBECTL_MISSING="Error: kubectl is missing! Please install and try again."
RESPONSE_MESSAGE_KUBEAPI_ERROR="Error: kubeserver api connection error! Check if KUBECONFIG is set."
RESPONSE_MESSAGE_RW_ACCESS_PKI="Error: No write access to the kubernetes PKI! Login as superuser and try again."
RESPONSE_MESSAGE_GENRSA_FAIL="Error: Generating RSA failed."
RESPONSE_MESSAGE_GENCERTSIGNREQUEST_FAIL="Error: Generating Cert Signing Request failed."
RESPONSE_MESSAGE_SIGNCERT_FAIL="Error: Signing Certificate failed."
RESPONSE_MESSAGE_REGCREDK8S_FAIL="Error: Credential Registration with K8S failed"
RESPONSE_MESSAGE_CREATECLUSTERROLE_FAIL="Error: Creation of Cluster Role failed"
RESPONSE_MESSAGE_CREATEROLEBINDING_FAIL="Error: Creation of Cluster Role Binding failed"
RESPONSE_MESSAGE_WGET_MISSING="Error: wget is missing! Please install and try again."
RESPONSE_MESSAGE_KEYTOOL_MISSING="Error: keytool is missing! Please install and try again."
RESPONSE_MESSAGE_CFSSL_MISSING="Error: cfssl is missing! Please install and try again."
RESPONSE_MESSAGE_CFSSL_MISSING="Error: cfssljson is missing! Please install and try again."
RESPONSE_MESSAGE_ANSIBLE_MISSING="Error: ansible is missing! Please install and try again."


USAGE_INFO='Usage:  -u userName -s "hostname1.mydomain.net,hostname2" -c /path/to/K8S_SERVER_CA_CERT -k /path/to/K8S_SERVER_CA_KEY -a "path/to/K8S_APIServer_CERT/" -r "resource1,resource2,resource3" -v "verb1,verb2,verb3 -d "/path/to/copy/keystores" -f "true/if/using/container/else/false" -i "containerId_or_name/required/when/using/containers" -p "container/path/to/copy/keystores" -e "true/if/using/vm/worker/else/false" -y "keystone/version/required/when/using/vms" -t "openstack/tenant/name" -b "openstack/scope" -o "openstack/username/" -g "password/openstack/username" -l "openstack/uri"'


# function to run a command and check if executed OK
function check_exec {
        exec_to_check=$1
        resp_if_fail=$2
        resp_err_code=$3
        $exec_to_check > /dev/null
        if [ $? -ne 0 ]
        then
                echo $resp_if_fail
                exit $resp_err_code
        fi
}

# check if openssl is present
check_exec 'whereis openssl' "$RESPONSE_MESSAGE_OPENSSL_MISSING" $RESPONSE_OPENSSL_MISSING

#check if wget is present
check_exec 'whereis wget' "$RESPONSE_MESSAGE_WGET_MISSING" $RESPONSE_WGET_MISSING

# we need kubectl and api server to add the user
# check if kubectl command is present
check_exec 'whereis kubectl' "$RESPONSE_MESSAGE_KUBECTL_MISSING" $RESPONSE_KUBECTL_MISSING

# check if API server is up and running
check_exec 'kubectl get no' "$RESPONSE_MESSAGE_KUBEAPI_ERROR" $RESPONSE_KUBECTL_ERROR

#check if keytool is installed
check_exec 'keytool' "$RESPONSE_MESSAGE_KEYTOOL_MISSING" $ERR_KEYTOOL_MISSING

# check if cfssl and cfssl json are installed
check_exec 'cfssl version' "$RESPONSE_MESSAGE_CFSSL_MISSING" $ERR_CFSSL_MISSING
check_exec 'whereis cfssljson' "$RESPONSE_MESSAGE_CFSSLJSON_MISSING" $ERR_CFSSLJSON_MISSING

#check if ansible is present
check_exec 'whereis ansible' "$RESPONSE_MESSAGE_ANSIBLE_MISSING" $RESPONSE_ANSIBLE_MISSING

# check if we have access to PKI
#if [ ! -w /etc/kubernetes/pki ]
#then
#        echo $RESPONSE_MESSAGE_RW_ACCESS_PKI
#        exit $RESPONSE_RW_ACCESS_PKI
#fi

while getopts :u:c:k:r:v:s:a:d:f:i:p:e:y:t:b:o:g:l:h opt
do
case "$opt" in
	u)	user="${OPTARG}";;
	k)	ca_key="${OPTARG}";;
	s)	sans="${OPTARG}";;
	c)	ca_cert="${OPTARG}";;
	r)	resources="${OPTARG}";;
	v)	verbs="${OPTARG}";;
	a)	k8s_server_cert="${OPTARG}";;
	d)	baremetal_path="${OPTARG}";;
	f)	docker_flag="${OPTARG}";;
	i)	container="${OPTARG}";;
	p)	container_path="${OPTARG}";;
	e)	vm_worker_enabled="${OPTARG}";;
	y)	keystone_version="${OPTARG}";;
	t)	openstack_tenant="${OPTARG}";;
	b)	openstack_scope="${OPTARG}";;
	o)	openstack_username="${OPTARG}";;
	g)	openstack_pass="${OPTARG}";;
	l)	openstack_uri="${OPTARG}";;
	h)	echo "$USAGE_INFO"; exit;;
esac
done

# We take the user name as input.
# If user has not been provided, then respond with error message.
#if [ -z "$user" -o ! -r "$ca_key" -o ! -r "$ca_cert" -o -z "$sans" -o ! -r "$k8s_server_cert" -o -z "$baremetal_path" -o "$docker_flag" != "true" -a "$docker_flag" != "false" -o "$vm_worker_enabled" !="true" -a "$vm_worker_enabled"!="false" ]
if [ -z "$user" -o ! -r "$ca_key" -o ! -r "$ca_cert" -o -z "$sans" -o ! -r "$k8s_server_cert" -o -z "$baremetal_path" ] || [ "$docker_flag" != "true" -a "$docker_flag" != "false" ] || [ "$vm_worker_enabled" != "true" -a "$vm_worker_enabled" != "false" ]
then
	echo "ERROR: mandatory arguments missing!"
        echo $USAGE_INFO
	exit

        if [ -z "$user" ]
        then
                exit $RESPONSE_USERNAME_MISSING
        elif [ ! -r "$ca_cert" ]
        then
                exit  $RESPONSE_CA_CERT_MISSING
        elif [ ! -r "$ca_key" ]
        then
                exit $RESPONSE_CA_KEY_MISSING
	elif [ -z "$sans" ]
	then
		exit $RESPONSE_SANS_MISSING
	elif [ ! -r "$k8s_server_cert" ]
	then
		exit $RESPONSE_SERVER_CERT_MISSING
	elif [ -z "$baremetal_path" ]
	then
		exit $RESPONSE_DIRECTORY_PATH_MISSING
	elif [ "$docker_flag" != "true" ] && [ "$docker_flag" != "false" ]
	then
		exit $RESPONSE_DOCKER_FLAG_OTHER_VALUES
	elif [ "$vm_worker_enabled" != "true" ] && [ "$vm_worker_enabled" != "false" ]
	then
		exit $RESPONSE_VM_WORKER_FLAG_OTHER_VALUES
	fi
fi

#Checking container Id and path
if [ "$docker_flag" == "true" ]
then
	if [ -z "$container" -o -z "$container_path" ] 
	then
		echo "ERROR: mandatory arguments missing!"
        	echo $USAGE_INFO
		exit $RESPONSE_DOCKER_FIELDS_MISSING
	fi
fi

#Checking openstack config
if [ "$vm_worker_enabled" == "true" ]
then
        if [ -z "$keystone_version" -o -z "$openstack_tenant" -o -z "$openstack_scope" -o -z "$openstack_username" -o -z "$openstack_pass" -o -z "$openstack_uri" ]
        then
                echo "ERROR: mandatory arguments missing!"
                echo $USAGE_INFO
                exit $RESPONSE_OPENSTACK_FIELDS_MISSING
        fi
fi

echo -n "Creating Cert Signing Request for the client certificate"
cat > ${user}-csr.json<<EOF
{
  "hosts": [
        `echo $sans | tr -s "[:space:]"| sed 's/,/\",\"/g' | sed 's/^/\"/' | sed 's/$/\"/' | sed 's/,/,\n/g'`
  ],
  "CN": "$user",
  "key": {
        "algo": "rsa",
        "size": 2048
  }
}
EOF

cfssl genkey ${user}-csr.json | cfssljson -bare ${user}
if [ $? -ne 0 ]
then
	echo "$RESPONSE_MESSAGE_GENCERTSIGNREQUEST_FAIL"
	exit $RESPONSE_GENCERTSIGNREQUEST_FAIL
fi
echo "Done"

echo -n "Signing the client certificate"
cfssl sign -csr=${user}.csr -ca-key=${ca_key} -ca=${ca_cert} | cfssljson -bare ${user}
if [ $? -ne 0 ]
then
        echo "$RESPONSE_MESSAGE_SIGNCERT_FAIL"
        exit $RESPONSE_SIGNCERT_FAIL
fi
echo "Done"

mv ${user}-key.pem ${user}.key
mv ${user}.pem ${user}.crt

echo -n "Registering the client credentials with K8S"
check_exec "kubectl config set-credentials ${user} --client-certificate=${user}.crt --client-key=${user}.key" "$RESPONSE_MESSAGE_REGCREDK8S_FAIL" $RESPONSE_REGCREDK8S_FAIL
echo "done"

# Create a K8S role that specifies the level of access:
echo -n "Provisioning access in K8S..."
clusterrolename=cr_${user}_`echo $verbs | tr -d ','`_`echo $resources | tr -d ',.'`
#check_exec "kubectl create clusterrole $clusterrolename --verb=${verbs} --resource=${resources}" "$RESPONSE_MESSAGE_CREATECLUSTERROLE_FAIL : ${clusterrolename}" $RESPONSE_CREATECLUSTERROLE_FAIL
kubectl create clusterrole $clusterrolename --verb=${verbs} --resource=${resources} -o json
if [ $? -ne 0 ]
then
	# we rollback the operation
	echo -e "ClusterRole ${clusterrolename} already exists for the user ${user} for the resources ${resources} and operations ${verbs}.\nDelete existing role using the command \nkubectl delete clusterrole $clusterrolename\nand try again."
	#kubectl delete clusterrole $clusterrolename
	#echo "$RESPONSE_MESSAGE_CREATECLUSTERROLE_FAIL : ${clusterrolename}"
	exit $RESPONSE_CREATECLUSTERROLE_FAIL
fi

# Create a RoleBinding or a ClusterRoleBinding to assign the Role to the account:
#check_exec "kubectl create clusterrolebinding crb_${clusterrolename} --clusterrole=${clusterrolename} --user=${user} -o json" "$RESPONSE_MESSAGE_CREATEROLEBINDING_FAIL : crb_${clusterrolename}" $RESPONSE_CREATEROLEBINDING_FAIL
kubectl create clusterrolebinding crb_${clusterrolename} --clusterrole=${clusterrolename} --user=${user}
if [ $? -ne 0 ]
then
        # we rollback the operation
        #echo "ROLLBACK: deleting clusterrolebinding crb_${clusterrolename}"
        #kubectl delete clusterrolebinding crb_$clusterrolename
	echo -e "ClusterRoleBinding crb_${clusterrolename} already exists for the user ${user} for the resources ${resources} and operations ${verbs}.\nDelete existing ClusterRoleBinding and associated ClusterRole using the command \nkubectl delete clusterrolebinding crb_$clusterrolename"
	echo -e "\nkubectl delete clusterrole $clusterrolename\nand try again."

        #echo "$RESPONSE_MESSAGE_CREATECLUSTERROLEBINDING_FAIL : crb_${clusterrolename}"
        exit $RESPONSE_CREATECLUSTERROLEBINDING_FAIL
fi
echo "Done"

echo Client Certificate generated at `ls ${user}.crt`
echo Client key generated at `ls ${user}.key`

suffix1=""

echo "Creating PKCS12 Keystore and Trust Store"
while [ -r ${user}_k8s_client${suffix1}.jks ]
do
	suffix1=$[ suffix1 + 1 ]
	#echo "Keystore ${user}_k8s_client.jks already exists. Remove/rename and try again."
	#exit $ERR_KEYSTORE_EXIST
done

suffix2=""
while [ -r ${user}_k8s_trust${suffix2}.jks ]
do
	suffix2=$[ suffix2 + 1 ]
	#echo "Trust store ${user}_k8s_trust.jks already exists. Remove/rename and try again."
	#exit $ERR_TRUST_STORE_EXIST
done

user_jks=${user}_k8s_client${suffix1}.jks
trust_jks=${user}_k8s_trust${suffix2}.jks

export TRUST_STORE_PASS=`openssl rand -base64 128 | tr -dc _A-Z-a-z-0-9 | head -c64`
export CERT_STORE_PASS=`openssl rand -base64 128 | tr -dc _A-Z-a-z-0-9 | head -c64`
openssl pkcs12 -export -in ${user}.crt -inkey ${user}.key -out ${user}.p12 -name ${user}_client -passout env:CERT_STORE_PASS 
keytool -importkeystore -destkeystore $user_jks -deststorepass:env CERT_STORE_PASS -alias ${user}_client -srckeystore ${user}.p12 -srcstoretype PKCS12 -srcstorepass:env CERT_STORE_PASS
keytool -import -keystore $trust_jks  -alias ${user}_k8s_server -file $k8s_server_cert -noprompt -deststorepass:env TRUST_STORE_PASS


#RESPONSE_SUCCESS_CONFIG_ATTHUB_INSTRUCTION="\n---------NEXT STEPS------------\nPlace the generated keystore (.jks) files:`ls $user_jks $trust_jks` in attestation hub configuration folder.\nAdd the following to your tenant configuration:\nkubernetes.client.keystore /opt/attestation-hub/configuration/${user_jks}\nkubernetes.client.keystore.password ${CERT_STORE_PASS}\nkubernetes.server.keystore /opt/attestation-hub/configuration/${trust_jks}\nkubernetes.server.keystore.password ${TRUST_STORE_PASS}"



cat > ${user}_keystore.properties<<EOF
kubernetes.client.keystore=/opt/attestation-hub/configuration/`ls $user_jks`
kubernetes.client.keystore.password=${CERT_STORE_PASS}
kubernetes.server.keystore=/opt/attestation-hub/configuration/`ls $trust_jks`
kubernetes.server.keystore.password=${TRUST_STORE_PASS}
EOF

if [ "$vm_worker_enabled" == "true" ]
then
        cat > ${user}_JKS_NextSteps<<EOF
			{
			"name": "YOURTENANTNAME",
			"plugins": [{
					"name": "kubernetes",
					"properties": [{
						"key": "api.endpoint",
						"value": "https://YOUR-K8S-END-POINT-URL:6443"
					},
					{
						"key": "tenant.name",
						"value": "YOURTENANTNAME"
					},
					{
						"key": "plugin.provider",
						"value": "com.intel.attestationhub.plugin.kubernetes.KubernetesPluginImpl"
					},
					{
						"key": "tenant.kubernetes.keystore.config",
						"value": "/opt/attestation-hub/configuration/`ls ${user}_keystore.properties`"
					},
					{
						"key": "keystone.version",
						"value": "${keystone_version}"
					},
					{
						"key": "vm.worker.enabled",
						"value": "${vm_worker_enabled}"
					},
					{
						"key": "openstack.tenant.name",
						"value": "${openstack_tenant}"
					},
					{
						"key": "openstack.scope",
						"value": "${openstack_scope}"
					},
					{
						"key": "openstack.username",
						"value": "${openstack_username}"
					},
					{
						"key": "openstack.pass",
						"value": "${openstack_pass}"
					},
					{
						"key": "openstack.uri",
						"value": "${openstack_uri}"
					}
			]
	}]
}
EOF
			
			
		cat > ${user}_nova_tenant.json<<EOF
			{
			"name": "<<tenant name>>",
			"plugins": [{
				"name": "nova",
				"properties": [{
						"key": "plugin.provider",
						"value": "com.intel.attestationhub.plugin.nova.NovaPluginImpl"
					},
					{
						"key": "api.endpoint",
						"value": "<<openstack api endpoint>>"
					},
					{
						"key": "auth.endpoint",
						"value": "<<openstack auth endpoint>>"
					},
					{
						"key": "auth.version",
						"value": "v${keystone_version}"
					},
					{
						"key": "user.name",
						"value": "${openstack_username}"
					},
					{
						"key": "user.password",
						"value": "${openstack_pass}"
					},
					{
						"key": "tenant.name",
						"value": "${openstack_tenant}"
					},
					{
						"key": "domain.name",
						"value": "<<domain name>>"
					}
				]
			}]
			}
EOF
			
else
	cat > ${user}_JKS_NextSteps<<EOF
			{
			"name": "YOURTENANTNAME",
			"plugins": [
			  {
				"name": "kubernetes",
				"properties": [
                                  {
                                       "key": "user.name",
                                        "value": "admin"
                                  },
                                  {
                                       "key": "user.password",
                                        "value": "password"
                                  },
				  {
					"key": "api.endpoint",
					"value": "https://YOUR-K8S-END-POINT-URL:6443"
				  },
				  {
					"key": "tenant.name",
					"value": "YOURTENANTNAME"
				  },
				  {
					"key": "plugin.provider",
					"value": "com.intel.attestationhub.plugin.kubernetes.KubernetesPluginImpl"
				  },
				  {
					"key": "tenant.kubernetes.keystore.config",
					"value": "/opt/attestation-hub/configuration/`ls ${user}_keystore.properties`"
				  }		
				]
			  }
			]
			}
EOF
	

fi


export CLIENT_KEYSTORE=`ls $user_jks`
export SERVER_KEYSTORE=`ls $trust_jks`
export CLIENT_JKS=`ls ${user}_JKS_NextSteps`
export KEYSTORE_CONF=`ls ${user}_keystore.properties`

# Executing Ansible script
#ansible-playbook --extra-vars "directory=$baremetal_path server_keystore=$SERVER_KEYSTORE client_keystore=$CLIENT_KEYSTORE 
#client_jks=$CLIENT_JKS keystore_conf=$KEYSTORE_CONF container=$container docker_dir=$container_path docker=$docker_flag" keystores_copy.yml 

# Clean up
rm ${user}.p12 ${user}.csr ${user}-csr.json

unset TRUST_STORE_PASS
unset CERT_STORE_PASS
unset CLIENT_KEYSTORE
unset SERVER_KEYSTORE
unset CLIENT_JKS
unset KEYSTORE_CONF 
