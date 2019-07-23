# Conclusion

In this brief guide we demonstrated a convenient workflow for enabling App Mesh integration using the console for ECS tasks — this works whether you are using Fargate or EC2 launch types. Using the console, you can easily update task definitions to have an Envoy service mesh proxy container added and configured with defaults, requiring just a few inputs from you.  
  
This is useful for experimenting with App Mesh. Using the demo application, we were able to see how we can use the a traffic management feature of App Mesh to easily apply a routing rule to distribute traffic against two different service versions. After experimenting, you can take the generated task definition configuration and apply it to your own automated deployment process.  
  
App Mesh makes it simple to apply various release strategies, such as blue-green deployments, canary releases, and A/B tests. With a mesh in place, it is also easy to gain insights into application behavior and performance because service proxies are already instrumented for you.   
  
App Mesh has out of the box support for Amazon CloudWatch logs and metrics and AWS X-Ray for distributed tracing. There is also a healthy ecosystem of commercial and open source providers, including enterprise support for Envoy.  
  
Want to know more about App Mesh? Visit the [docs](https://docs.aws.amazon.com/app-mesh/), [examples repo](https://github.com/aws/aws-app-mesh-examples), and check out the [roadmap](https://github.com/aws/aws-app-mesh-roadmap). You can get the demo for this article [here](https://github.com/subfuzion/enable-appmesh).

