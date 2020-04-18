FROM ruby:2.5.7-alpine
MAINTAINER Alexey Zhuravlev <alexey.g.zhuravlev@gmail.com>
LABEL Description="AZ OtusHomework Ruby" Version="1.0"

ARG DEV_PACKAGES="make gcc libc-dev"

RUN apk --no-cache --update add --virtual build-dependencies $DEV_PACKAGES

WORKDIR /app

COPY Gemfile* ./

RUN gem install bundler \
    && bundle config --local without 'development test' \
    && bundle install \
    && apk del build-dependencies

COPY . .

ENV APP_PORT 8000
ENV RACK_ENV production
ENV MAX_THREADS 1

EXPOSE 8000

CMD [ "bin/entry" ]
