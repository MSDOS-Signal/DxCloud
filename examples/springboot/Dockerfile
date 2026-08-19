# Spring Boot 项目在 DxCloud Pipeline 中的构建运行环境。
# 流水线中的 maven-package 步骤会把 target/*.jar 写入 workspace，这里直接 COPY。
FROM eclipse-temurin:17-jre
WORKDIR /app
COPY target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "/app/app.jar"]
