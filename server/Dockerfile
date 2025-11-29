# ---------- 运行阶段 ----------
FROM debian:bookworm-slim

LABEL stage="runtime" \
      app_version="1.0.0" \
      description="Brook server image built from precompiled dist artifacts"

WORKDIR /app

ARG APP_PATH

COPY $APP_PATH ./

RUN chmod 755 /app && chmod -R 755 /app/*

EXPOSE 8000

CMD ["./brook-sev --container true"]