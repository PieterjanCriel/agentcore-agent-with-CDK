FROM public.ecr.aws/docker/library/golang:tip-trixie

WORKDIR /app

# Copy entire project (respecting .dockerignore)
COPY . .

# Install dependencies
RUN go mod download

# Build the application
RUN go build -o /app/main

# Set AWS region environment variable
ENV AWS_REGION=eu-central-1
ENV AWS_DEFAULT_REGION=eu-central-1

# Signal that this is running in Docker for host binding logic
ENV DOCKER_CONTAINER=1

# Create non-root user
RUN useradd -m -u 1000 bedrock_agentcore
USER bedrock_agentcore

EXPOSE 8080

CMD ["/app/main"]
