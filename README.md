# Traefik AWS Plugin

This is a [Traefik middleware plugin](https://plugins.traefik.io) which pushes data to and pulls data from Amazon Web Services (AWS) for a Traefik instance running in [Amazon Elastic Container Service (ECS)](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/Welcome.html).
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

## Services

### Local storage

To store objects in a local directory, use the following labels (example):

```text
"traefik.http.middlewares.my-aws.plugin.aws.type" : "local"
"traefik.http.middlewares.my-aws.plugin.aws.directory" : "aws-local-directory"
```

`GET` and `PUT` are supported.

### S3

To store objects in [Amazon Simple Storage Service (S3)](https://docs.aws.amazon.com/AmazonS3/latest/userguide), use the following labels (example):

```text
"traefik.http.middlewares.my-aws.plugin.aws.service" : "s3"
"traefik.http.middlewares.my-aws.plugin.aws.bucket" : "my-bucket"
"traefik.http.middlewares.my-aws.plugin.aws.region" : "us-west-2"
"traefik.http.middlewares.my-aws.plugin.aws.prefix" : "/prefix"
}
```

Note that `prefix` must include the leading slash.

[PUT](https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObject.html) and [GET](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html) are supported.

When forwarding the request to S3, the plugin sets the following headers:

* `Host`
* `Authorization` with the [AWS API request signature](https://docs.aws.amazon.com/IAM/latest/UserGuide/create-signed-request.html); the [ECS task IAM role credentials](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html) are used to sign the request
* `date`, if not defined
* `X-Amz-Content-Sha256`
* `X-Amz-Date`
* `x-amz-security-token`

### DynamoDB

[Amazon DynamoDB](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide) support is pending.

## Development

To develop `traefik-aws-plugin` in a local workspace:

* Install [go](https://go.dev/doc/install)
* Install [golangci-lint](https://golangci-lint.run/usage/install/)
