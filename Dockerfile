FROM ruby:2.6-alpine

RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

RUN apk update
RUN gem install bundler

WORKDIR /app

COPY Gemfile Gemfile.lock ./
RUN bundle install --without development

COPY app.rb ./
COPY public/ ./public
COPY views/ ./views
RUN mkdir /app/data

RUN chown -R appuser:appgroup /app

USER 1000

CMD ["ruby", "app.rb", "-o", "0.0.0.0"]
