# ---------- 运行阶段 ----------
FROM debian:bookworm-slim

LABEL stage="runtime" \
      app_version="1.0.0" \
      description="Brook server image built from precompiled dist artifacts"

WORKDIR /app

COPY dist/brook-cli/ ./

RUN chmod 755 /app && chmod -R 755 /app/*

CMD ["./brook-cli"]