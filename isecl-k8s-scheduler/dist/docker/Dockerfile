#/*
# * Copyright (C) 2020 Intel Corporation
# * SPDX-License-Identifier: BSD-3-Clause
# */

FROM registry.access.redhat.com/ubi8-init:8.1

RUN mkdir -p /var/log/isecl-k8s-extensions
RUN mkdir -p /opt/isecl-k8s-extensions/config
RUN mkdir -p /opt/isecl-k8s-extensions/isecl-k8s-scheduler/config

COPY isecl-k8s-scheduler /isecl-k8s-scheduler

ENTRYPOINT ["/isecl-k8s-scheduler"]
