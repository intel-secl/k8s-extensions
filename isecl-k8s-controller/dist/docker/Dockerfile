#/*
# * Copyright (C) 2020 Intel Corporation
# * SPDX-License-Identifier: BSD-3-Clause
# */

FROM ubuntu:focal

RUN apt-get update -y && apt-get install -y curl
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl
RUN chmod +x kubectl
RUN mkdir -p /var/log/isecl-k8s-controller && apt-get clean && apt-get autoclean
COPY isecl-k8s-controller /isecl-k8s-controller

ENTRYPOINT ["/isecl-k8s-controller"]
