FROM ruby:2.7-alpine

RUN addgroup -g 1000 -S appgroup \
  && adduser -u 1000 -S appuser -G appgroup \
  && apk update \
  && gem install bundler

WORKDIR /app

COPY Gemfile Gemfile.lock ./
RUN bundle install

COPY app.rb ./
COPY views/ ./views
COPY lib/ ./lib
RUN mkdir /app/data

RUN chown -R appuser:appgroup /app

USER 1000

CMD ["ruby", "app.rb", "-o", "0.0.0.0"]
