FROM ubuntu:18.04

RUN apt-get update
RUN apt-get install -y software-properties-common curl
RUN apt-add-repository ppa:brightbox/ruby-ng

RUN export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)" && \
		echo "deb http://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
		curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -

RUN apt-get update
RUN apt-get install -y ruby2.4 google-cloud-sdk build-essential python3-pip git
RUN pip3 install awscli b2

WORKDIR /app
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh
