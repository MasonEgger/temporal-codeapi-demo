# ABOUTME: Codapi box for running Temporal Java SDK programs against a per-run dev server.
# ABOUTME: Learner code runs through the single-file source launcher with the SDK jars on the classpath.
ARG BASE_IMAGE=temporal-codapi/base:latest
FROM ${BASE_IMAGE} AS base

# Resolve the SDK and its transitive dependencies into a flat directory of jars.
FROM maven:3-eclipse-temurin-25 AS deps
WORKDIR /build
COPY pom.xml ./
RUN mvn --quiet dependency:copy-dependencies -DoutputDirectory=/libs -DincludeScope=runtime

FROM eclipse-temurin:25-jdk

COPY --from=base /usr/local/bin/temporal /usr/local/bin/temporal-run /usr/local/bin/
COPY --from=deps /libs /opt/temporal/libs
COPY java.args /opt/temporal/java.args

RUN apt-get update \
    && apt-get install -y --no-install-recommends jq \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --create-home sandbox \
    && mkdir /app \
    && chown sandbox:sandbox /app

# JVM flags live in java.args (passed as @/opt/temporal/java.args) so the launcher stays quiet.
ENV CLASSPATH=/opt/temporal/libs/* \
    TEMPORAL_APP_DIR=/app

WORKDIR /app
USER sandbox
