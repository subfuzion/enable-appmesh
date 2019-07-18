import {Construct, Duration, Stack, StackProps} from '@aws-cdk/core';
import {Port, SecurityGroup, SubnetType, Vpc} from '@aws-cdk/aws-ec2';
import {AwsLogDriver, Cluster, ContainerImage, FargateService, FargateTaskDefinition} from '@aws-cdk/aws-ecs';
import {Effect, ManagedPolicy, PolicyDocument, PolicyStatement, Role, ServicePrincipal} from '@aws-cdk/aws-iam';
import {ApplicationLoadBalancer, HealthCheck} from "@aws-cdk/aws-elasticloadbalancingv2";
import {LoadBalancedFargateService, LoadBalancerType} from "@aws-cdk/aws-ecs-patterns";

export class MeshDemoStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    // ========================================================================

    // Deploy a VPC
    // It will have 2 AZs, 2 NAT gateways, and an internet gateway
    const vpc = new Vpc(this, 'demovpc', {
      cidr: '10.0.0.0/16',
      maxAzs: 2,
      subnetConfiguration: [
        {
          cidrMask: 24,
          name: 'ingress',
          subnetType: SubnetType.PUBLIC,
        },
        {
          cidrMask: 24,
          name: 'application',
          subnetType: SubnetType.PRIVATE,
        },
      ]
    });

    // Allow inbound web traffic on port 80
    const externalSecurityGroup = new SecurityGroup(this, 'demoexternalsg', {
      vpc: vpc,
      allowAllOutbound: true,
    });
    externalSecurityGroup.connections.allowFromAnyIpv4(Port.tcp(80));

    // Allow communication within the vpc for the app and envoy containers
    // - 8080: default app port for gateway and colorteller
    // - 9901: envoy admin interface, used for health check
    // - 15000: envoy ingress ports (egress over 15001 will be allowed by allowAllOutbound)
    const internalSecurityGroup = new SecurityGroup(this, 'demointernalsg', {
      vpc: vpc,
      allowAllOutbound: true,
    });
    [ Port.tcp(8080), Port.tcp(9901), Port.tcp(15000) ].forEach(port => {
      internalSecurityGroup.connections.allowInternally(port);
    });

    // ========================================================================

    // Deploy a Fargate cluster on ECS
    const cluster = new Cluster(this, 'democluster', {
      vpc: vpc,
    });

    // Use Cloud Map for service discovery within the cluster, which
    // relies on either ECS Service Discovery or App Mesh integration
    // (default: cloudmap.NamespaceType.DNS_PRIVATE)
    const namespace = "mesh.local";
    cluster.addDefaultCloudMapNamespace({
      name: namespace,
    });

    // IAM role for the color app tasks
    const taskRole = new Role(this, 'demorole', {
      assumedBy: new ServicePrincipal('ecs-tasks.amazonaws.com'),
      managedPolicies: [
        ManagedPolicy.fromAwsManagedPolicyName('CloudWatchLogsFullAccess'),
        ManagedPolicy.fromAwsManagedPolicyName('AWSXRayDaemonWriteAccess'),
        ManagedPolicy.fromAwsManagedPolicyName('AmazonEC2ContainerRegistryReadOnly'),
      ]
    });

    const gatewayTaskDef = new FargateTaskDefinition(this, 'gatewaytaskdef', {
      taskRole: taskRole,
      cpu: 512,
      memoryLimitMiB: 1024,
    });

    //repositoryarn: '226767807331.dkr.ecr.us-west-2.amazonaws.com/gateway:latest',
    const gatewayContainer = gatewayTaskDef.addContainer('gatewaycontainer', {
      image: ContainerImage.fromRegistry('subfuzion/colorgateway:latest'),
      environment: {
        SERVER_PORT: '8080',
        COLOR_TELLER_ENDPOINT: `colorteller.${namespace}:8080`,
      },
      logging: new AwsLogDriver({
        streamPrefix: 'app',
      }),
    });
    gatewayContainer.addPortMappings({
      containerPort: 8080,
    })

    const gatewayService = new FargateService(this, 'gatewayservice', {
      cluster: cluster,
      taskDefinition: gatewayTaskDef,
      desiredCount: 1,
      serviceName: 'gateway',
      securityGroup: internalSecurityGroup,
      cloudMapOptions: {
        name: 'gateway',
      },
    });

    const healthCheck: HealthCheck = {
      "path": '/ping',
      "port": 'traffic-port',
      "interval": Duration.seconds(30),
      "timeout": Duration.seconds(5),
      "healthyThresholdCount": 2,
      "unhealthyThresholdCount": 2,
      "healthyHttpCodes": "200-499",
    };

    const alb = new ApplicationLoadBalancer(this, 'demoalb', {
      vpc: vpc,
      internetFacing: true,
      securityGroup: externalSecurityGroup,
    });
    const albListener = alb.addListener('web', {
      port: 80,
    });
    albListener.addTargets('demotarget', {
      port: 80,
      targets: [ gatewayService ],
      healthCheck: healthCheck,
    });


  }
}
