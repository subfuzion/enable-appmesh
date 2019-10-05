The purpose of this repo was to support a blog post about using the AWS console
to mesh-enable an existing, working example.

However, you can use `deploy-stack` to deploy a fully mesh-enabled demo. The
stack (`demo.yaml`) has updated task definitions that include proxy
configuration and an `envoy` container for a fully working App Mesh demo.

Use `delete-stack` when finished.

