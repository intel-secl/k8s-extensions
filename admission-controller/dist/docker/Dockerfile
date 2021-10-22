#/*
# * Copyright (C) 2021 Intel Corporation
# * SPDX-License-Identifier: BSD-3-Clause
# */

FROM centos:8

COPY isecl-k8s-admission-controller /usr/local/bin/isecl-k8s-admission-controller
RUN chmod +x /usr/local/bin/isecl-k8s-admission-controller

ENTRYPOINT ["isecl-k8s-admission-controller"]

