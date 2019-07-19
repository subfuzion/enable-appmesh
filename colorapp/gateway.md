# README

## colorteller app environment variables

* `COLOR` - \(optional\) override default color \("black"\)
* `ENABLE_ENVOY_XRAY_TRACING` - \(optional\) set to "1" to emit traces for AWS X-Ray
* `SERVER_PORT` - \(optional\) override default listening port \("8080"\)

## colorteller image

You can use a colorteller image from either of these repos:

* 226767807331.dkr.ecr.us-west-2.amazonaws.com/colorteller \(ECR for AWS accounts only\)
* subfuzion/colorteller \(public Docker Hub\)

You can use `deploy.sh` to build and push a colorteller image to your own ECR repo. Make sure you have created a `colorteller` repo under your account first. These are the environment variables used by the script:

* `COLOR_TELLER_IMAGE` - ex: 226767807331.dkr.ecr.us-west-2.amazonaws.com/colorteller:latest

Make sure the image name reflects the region you want to push to. The image will be pushed to this region for your account regardless of your AWS CLI profile region setting or `AWS_DEFAULT_REGION` environment variable setting.

If pushing to another repo across accounts, you will also need to set `AWS_DEFAULT_PROFILE` to override your `[default]` CLI profile \(see [environment variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html)\). For example:

```text
AWS_DEFAULT_PROFILE=other COLOR_TELLER_IMAGE=226767807331.dkr.ecr.us-west-2.amazonaws.com/colorteller:latest ./deploy.sh
```

