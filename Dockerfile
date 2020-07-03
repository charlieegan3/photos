# last arm < v8 image
FROM ruby:2.6.6-alpine

ARG CLOUD_SDK_VERSION=288.0.0
ENV CLOUD_SDK_VERSION=$CLOUD_SDK_VERSION
ENV CLOUDSDK_PYTHON=python3

ENV PATH /google-cloud-sdk/bin:$PATH
RUN apk --no-cache add \
        curl \
        python3 \
        py3-crcmod \
        bash \
        libc6-compat \
        openssh-client \
        git \
        make \
        gnupg \
        python3 \
        py3-pip \
    && curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    tar xzf google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    rm google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    gcloud config set core/disable_usage_reporting true && \
    gcloud config set component_manager/disable_update_check true && \
    gcloud --version

RUN pip3 install awscli b2

WORKDIR /app
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh
