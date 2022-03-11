#/*
# * Copyright (C) 2020 Intel Corporation
# * SPDX-License-Identifier: BSD-3-Clause
# */

FROM ubuntu:focal

RUN mkdir -p /var/log/isecl-k8s-scheduler

COPY isecl-k8s-scheduler /isecl-k8s-scheduler

ENTRYPOINT ["/isecl-k8s-scheduler"]
