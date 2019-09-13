#!/bin/bash

# Last Modified: 10/31/2017
# Author manux.ullas@intel.com

# The script is used by the Kubernetes Cluster Admin to sign a server certificate for the ISecL Extended Scheduler process.

# Pre-requisites:
# Script should be executed:
# 1) On the Kubernetes Master Node
# 2) As root user
# 3) After the Kubernetes Root CA has been generated

# Dependencies:
# 1) cfssl and cfssljson - for cert generation and signing

# Input Parameters:
# 1) -d workdir defaults to . 
# 2) -n "Common Name" - this will be populated in the Common Name entry in the server cert
# 3) -s* "hostname1.mydomain.net,hostname2" - these entries will populated in the 
#	Subject Alternative Names (SAN) field in the server cert
# 4) -k* /path/to/K8S_SERVER_CA_KEY - path to the Kubernetes Root CA Key
# 5) -c* /path/to/K8S_SERVER_CA_CERT - path to the Kubernetes Root CA Certificate 
# * - indicates mandatory arguments

# Output:
# The path to the newly signed certificate is printed.

# Example:
# ./create_k8s_extsched_cert.sh -n "K8S Extended Scheduler" \
# -s "10.242.131.141,kubernetesmaster3-desktop,kubernetesmaster3-desktop.iind.intel.com" \
# -c /etc/kubernetes/pki/ca.crt -k /etc/kubernetes/pki/ca.key
# open ca.pem: no such file or directory
# Creating Cert Signing Request
# 2017/10/31 13:44:11 [INFO] generate received request
# 2017/10/31 13:44:11 [INFO] received CSR
# 2017/10/31 13:44:11 [INFO] generating key: rsa-2048
# 2017/10/31 13:44:12 [INFO] encoded CSR
# Signing cert
# 2017/10/31 13:44:12 [INFO] signed certificate with serial number 79013915964554129834277509502675427720790484993
# Created certs for Extended Scheduler successfully.
# /home/kubernetesmaster3/tmp
# ./K8SExtendedScheduler_extsched_20181031.key
# ./K8SExtendedScheduler_extsched_20181031.crt

echo -e "Starting ISecL Kubernetes Extended Scheduler Server Certs Generation\n-------------------------------\n\n"

K8S_CERT_CN="`hostname` Kubernetes Extended Scheduler"
K8S_WORKDIR=.

ERROR_CFSSL_MISSING=-255
ERROR_CFSSLJSON_MISSING=-254
ERROR_INVALID_SERVER_CA_CERT_PATH=-252
ERROR_INVALID_SERVER_CA_KEY_PATH=-251
ERROR_WORKDIR_NO_WRITE_PERMS=-250
ERROR_SANS_NOT_PROVIDED=-249

#check if user running as root
if [ ! `id -u` -eq 0 ]
then
	echo "Error: script must be run as root! Aborting..."
	exit -251
fi

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

check_exec "which cfssl" "Error: cfssl not found on PATH, aborting..." $ERROR_CFSSL_MISSING
check_exec "which cfssljson" "Error: cfssljson not found on PATH, aborting..." $ERROR_CFSSLJSON_MISSING

while getopts :d:n:c:k:s:h opt
do
case "$opt" in
	n)	K8S_CERT_CN="${OPTARG}";;
	s)	K8S_SANS="${OPTARG}";;
	d)	K8S_WORKDIR="${OPTARG}";;
	k)	K8S_SERVER_CA_KEY="${OPTARG}";;
	c)	K8S_SERVER_CA_CERT="${OPTARG}";;
	h)	echo 'Usage: $0 [-d /working/directory] [-n CommonName] -c /path/to/K8S_SERVER_CA_CERT -k /path/to/K8S_SERVER_CA_KEY -s "hostname1.mydomain.net,hostname2,hostname3.yourdomain.com"'; exit;;
esac
done


if [ ! -r "$K8S_SERVER_CA_KEY" ]
then
	echo "Error: missing server CA key or invalid path. Aborting..."
	exit $ERROR_INVALID_SERVER_CA_KEY_PATH
fi

if [ ! -r "$K8S_SERVER_CA_CERT" ]
then
	echo "Error: missing server CA cert or invalid path. Aborting..."
	exit $ERROR_INVALID_SERVER_CA_CERT_PATH
fi

if [ -z "$K8S_SANS" ]
then
	echo "Error: Subject Alternative Names for the cert have not been provided. Aborting..."
	exit $ERROR_SANS_NOT_PROVIDED
fi



if [ ! -w $K8S_WORKDIR ]
then
        echo "Error: No write permissions for workdir, $K8S_WORKDIR. Aborting..."
        exit $ERROR_WORKDIR_NO_WRITE_PERMS
fi

cd $K8S_WORKDIR


NAME_PREFIX=`echo $K8S_CERT_CN | sed 's/[^a-zA-Z0-9]//g'`


# creating extended scheduler cert request
echo "Creating Cert Signing Request"

cat > k8sextscheduler.json <<EOF 
{
  "hosts": [
	`echo $K8S_SANS | tr -s "[:space:]"| sed 's/,/\",\"/g' | sed 's/^/\"/' | sed 's/$/\"/' | sed 's/,/,\n/g'`
  ],
  "CN": "$K8S_CERT_CN",
  "key": {
	"algo": "rsa",
	"size": 2048
  }
}
EOF

cfssl genkey k8sextscheduler.json | cfssljson -bare k8sextscheduler

if [ $? -ne 0 ]
then
	echo "Error generating CSR. Aborting..."
	exit $ERROR_GEN_CSR
fi

# sign cert
echo "Signing cert"
cfssl sign -csr=k8sextscheduler.csr -ca-key=${K8S_SERVER_CA_KEY} -ca=${K8S_SERVER_CA_CERT} | cfssljson -bare k8sextscheduler
if [ $? -ne 0 ]
then
	echo "Error signing cert. Aborting..."
	exit $ERROR_GEN_CERT
fi

exp_date=`cfssl certinfo -cert k8sextscheduler.pem | grep not_after | cut -d T -f 1 | tr -d '[:space:]-"' | cut -d : -f 2`


chmod 700 k8sextscheduler.pem
chmod 700 k8sextscheduler-key.pem

exp_date=`cfssl certinfo -cert k8sextscheduler.pem | grep not_after | cut -d T -f 1 | tr -d '[:space:]-"' | cut -d : -f 2`

mv  k8sextscheduler-key.pem server.key
mv k8sextscheduler.pem server.crt

echo "Created certs for Extended Scheduler successfully."

cd -

ls ${K8S_WORKDIR}/server.key
ls ${K8S_WORKDIR}/server.crt

# Cleaning up
rm  ${K8S_WORKDIR}/k8sextscheduler.csr  ${K8S_WORKDIR}/k8sextscheduler.json
