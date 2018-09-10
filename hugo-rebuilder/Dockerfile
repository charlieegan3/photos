FROM alpine:3.8

# Install curl for hugo install and pulling github zips
RUN apk add curl

# Install Hugo
ENV HUGO_VERSION 0.37
RUN curl -LO https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_${HUGO_VERSION}_Linux-64bit.tar.gz && \
    tar -xzf hugo_${HUGO_VERSION}_Linux-64bit.tar.gz && \
    rm hugo_${HUGO_VERSION}_Linux-64bit.tar.gz README.md && \
    mv hugo /usr/local/bin/

# Install Ruby
RUN apk update && apk upgrade && apk --update add \
    ruby ruby-irb ruby-rake ruby-io-console ruby-bigdecimal ruby-json ruby-bundler \
    libstdc++ tzdata bash ca-certificates \
    && echo 'gem: --no-document' > /etc/gemrc

# Install entrypoint
COPY entrypoint.sh /etc/entrypoint.sh
ENTRYPOINT /etc/entrypoint.sh
