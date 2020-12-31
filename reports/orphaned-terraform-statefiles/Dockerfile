FROM ruby:2.7-alpine

RUN apk --update add --virtual build_deps \
    build-base ruby-dev libc-dev linux-headers \
    curl

RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

COPY Gemfile Gemfile.lock ./

RUN bundle config set without 'development'
RUN bundle install

COPY bin ./bin
COPY lib ./lib

RUN apk add python3 py3-pip
RUN pip3 install awscli

ENV KOPS_VERSION=1.17.2
RUN curl -Lo /usr/local/bin/kops https://github.com/kubernetes/kops/releases/download/v${KOPS_VERSION}/kops-linux-amd64
RUN chmod +x /usr/local/bin/kops

RUN chown 1000:1000 /app
USER 1000

CMD ["/bin/sh", "./bin/post-data.sh"]
