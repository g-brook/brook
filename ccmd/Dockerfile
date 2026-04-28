# ---------- 运行阶段 ----------
FROM debian:bookworm-slim

LABEL stage="runtime" \
      app_version="1.0.0" \
      description="Brook server image built from precompiled dist artifacts"

WORKDIR /app

ARG APP_NAME=brook-cli
ARG TARGETARCH
ENV APP_NAME=$APP_NAME
LABEL container_name=$APP_NAME

COPY dist/${APP_NAME}_linux_${TARGETARCH}/ ./

RUN chmod 755 /app && chmod -R 755 /app/*

CMD ["./${APP_NAME} --container=true"]
