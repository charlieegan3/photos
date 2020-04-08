FROM ruby:2.7-alpine

RUN apk add curl python3 git bash
RUN curl -sSL https://sdk.cloud.google.com | bash
ENV PATH $PATH:/root/google-cloud-sdk/bin

RUN pip3 install awscli b2

WORKDIR /app
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh
