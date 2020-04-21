FROM registry.access.redhat.com/ubi8-init:8.1

RUN yum -y install curl

RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl
RUN chmod +x kubectl
RUN mkdir -p /var/log/isecl-k8s-extensions
RUN mkdir -p /opt/isecl-k8s-extensions/config/

COPY isecl-k8s-controller /isecl-k8s-controller

ENTRYPOINT ["/isecl-k8s-controller"]
