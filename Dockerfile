FROM ruby:2.6-alpine

RUN addgroup -g 1000 -S appgroup \
  && adduser -u 1000 -S appuser -G appgroup \
  && apk update \
  && gem install bundler \
  && bundle config set without 'development'

WORKDIR /app

COPY Gemfile Gemfile.lock ./
RUN bundle install

COPY app.rb helpers.rb ./
COPY views/ ./views
COPY data/ ./data

RUN chown -R appuser:appgroup /app

USER 1000

CMD ["ruby", "app.rb", "-o", "0.0.0.0"]
