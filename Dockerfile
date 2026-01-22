# syntax=docker/dockerfile:1
# check=error=true

ARG RUBY_VERSION=4.0.1
FROM docker.io/library/ruby:$RUBY_VERSION-alpine AS base

RUN apk add --no-cache tzdata procps

WORKDIR /rails

ENV RAILS_ENV="production" \
  BUNDLE_DEPLOYMENT="1" \
  BUNDLE_PATH="/usr/local/bundle" \
  BUNDLE_WITHOUT="development:test"

# Build stage
FROM base AS build

RUN apk add --no-cache build-base git yaml-dev tzdata

COPY Gemfile Gemfile.lock ./

RUN bundle install && \
  rm -rf ~/.bundle/ "${BUNDLE_PATH}"/ruby/*/cache "${BUNDLE_PATH}"/ruby/*/bundler/gems/*/.git && \
  bundle exec bootsnap precompile --gemfile

COPY . .

RUN bundle exec bootsnap precompile app/ lib/ && \
  SECRET_KEY_BASE_DUMMY=1 ./bin/rails assets:precompile && \
  rm -rf tmp/cache

# Final stage
FROM base

# Non-root user
RUN addgroup -S rails && adduser -S rails -G rails

COPY --from=build --chown=rails:rails "${BUNDLE_PATH}" "${BUNDLE_PATH}"
COPY --from=build --chown=rails:rails /rails /rails

USER rails:rails

EXPOSE 8080
ENV HTTP_PORT=8080
CMD ["./bin/thrust", "./bin/rails", "server"]
