# The workflow

The previous section got our demo app up and running on AWS Fargate. Now we will walk through the process of using the ECS console to update our task definitions to enable App Mesh for our app. This is the sequence of steps we’ll follow for this workflow:

1. Update our service mesh to take advantage of Cloud Map service discovery and prepare to take over routing for our tasks after we mesh-enable the task descriptions and restart tasks.
2. Create new revisions of our task definitions with App Mesh enabled.
3. Update our services to use the new mesh-enabled task definitions.
4. Confirm that the application continues to work as expected.
5. Update the mesh configuration to begin sending traffic to the green service and confirm this works.
6. Update the mesh configuration again to send all traffic now to the green service and confirm.

