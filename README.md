# Traefik AWS Plugin

This is a [Traefik middleware plugin](https://plugins.traefik.io) which pushes data to and pulls data from Amazon Web Services (AWS) for a Traefik instance running in [Amazon Elastic Container Service (ECS)](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/Welcome.html). The following is currently supported:

* [Amazon Simple Storage Service (S3)](https://docs.aws.amazon.com/AmazonS3/latest/userguide/Welcome.html): PUT
* [Amazon DynamoDB](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Introduction.html): pending

## Configuration

traefik.yml:

```yaml
providers:
  ecs:
    exposedByDefault: false

experimental:
  plugins:
    aws:
      moduleName: github.com/bluecatengineering/traefik-aws-plugin
      version: v0.1.2
```

Example labels for a given router:

```text
"traefik.enable" : "true"
"traefik.http.routers.my-router.service" : "noop@internal"
"traefik.http.routers.my-router.rule" : "Host(`aws.myhostexample.io`)"
"traefik.http.routers.my-router.middlewares" : "my-aws"
```

S3 example labels, `prefix` includes leading slash:

```text
"traefik.http.middlewares.my-aws.plugin.aws.service" : "s3"
"traefik.http.middlewares.my-aws.plugin.aws.bucket" : "my-bucket"
"traefik.http.middlewares.my-aws.plugin.aws.region" : "us-west-2"
"traefik.http.middlewares.my-aws.plugin.aws.prefix" : "/prefix"
}
```

Local directory example labels:

```text
"traefik.http.middlewares.my-aws.plugin.aws.type" : "local"
"traefik.http.middlewares.my-aws.plugin.aws.directory" : "aws-local-directory"
```

## Development

To develop `traefik-aws-plugin` in a local workspace:

* Install [go](https://go.dev/doc/install)
* Install [golangci-lint](https://golangci-lint.run/usage/install/)
