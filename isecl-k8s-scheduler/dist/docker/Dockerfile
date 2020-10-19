#/*
# * Copyright (C) 2020 Intel Corporation
# * SPDX-License-Identifier: BSD-3-Clause
# */

FROM centos:8

RUN useradd -u 1001 --home /opt/isecl-k8s-extensions  --shell /bin/false isecl-scheduler

RUN mkdir -p /var/log/isecl-k8s-extensions

COPY isecl-k8s-scheduler /isecl-k8s-scheduler

ENTRYPOINT ["/isecl-k8s-scheduler"]
