FROM test-hello:v2

LABEL maintainer="Sebastien Vincent" \
      description="Hello world in nodejs" \
      version="0.0.1" 

COPY ./src /app
WORKDIR /app

EXPOSE 8080

CMD ["echo", "v3"]

