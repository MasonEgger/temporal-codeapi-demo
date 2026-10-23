# ABOUTME: Codapi box for the Java IP geolocation tutorial: a Temporal dev server, Worker, and client per run.
# ABOUTME: Learner sources compile with javac against the SDK jars; tests run on the JUnit console launcher.
ARG BASE_IMAGE=temporal-codapi/base:latest
FROM ${BASE_IMAGE} AS base

# Resolve the SDK, test libraries, and their transitive dependencies into a flat directory of jars.
FROM maven:3-eclipse-temurin-25 AS deps
WORKDIR /build
COPY codapi/sandboxes/temporal-java/pom.xml ./
RUN mvn --quiet dependency:copy-dependencies -DoutputDirectory=/libs -DincludeScope=runtime

FROM eclipse-temurin:25-jdk

COPY --from=base /usr/local/bin/temporal /usr/local/bin/temporal-serve /usr/local/bin/temporal-run /usr/local/bin/
COPY --from=deps /libs /opt/temporal/libs
COPY codapi/sandboxes/temporal-java/java.args /opt/temporal/java.args

RUN apt-get update \
    && apt-get install -y --no-install-recommends curl jq \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home sandbox \
    && mkdir /app \
    && chown sandbox:sandbox /app

ENV TEMPORAL_APP_DIR=/app

WORKDIR /app
USER sandbox
