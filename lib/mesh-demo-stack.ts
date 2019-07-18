import {Construct, Duration, Stack, StackProps} from '@aws-cdk/core';
import {Port, SecurityGroup, SubnetType, Vpc} from '@aws-cdk/aws-ec2';
import {AwsLogDriver, Cluster, ContainerImage, FargateService, FargateTaskDefinition} from '@aws-cdk/aws-ecs';
import {Effect, ManagedPolicy, PolicyDocument, PolicyStatement, Role, ServicePrincipal} from '@aws-cdk/aws-iam';
import {ApplicationLoadBalancer, HealthCheck} from "@aws-cdk/aws-elasticloadbalancingv2";
import {LoadBalancedFargateService, LoadBalancerType} from "@aws-cdk/aws-ecs-patterns";

export class MeshDemoStack extends Stack {

  readonly APP_TAG = "v1";
  readonly APP_PORT = 8080;

  taskRole: Role;
  taskExecutionRole: Role;

  vpc: Vpc;
  cluster: Cluster;
  namespace: string;

  internalSecurityGroup: SecurityGroup;
  externalSecurityGroup: SecurityGroup;

  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    this.createVpc();
    this.createCluster();
    this.createGateway();
    this.createColorTeller('blue', 'green');
  }

  createVpc() {
    // The VPC will have 2 AZs, 2 NAT gateways, and an internet gateway
    this.vpc = new Vpc(this, 'demoVPC', {
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
    this.externalSecurityGroup = new SecurityGroup(this, 'DemoExternalSG', {
      vpc: this.vpc,
      allowAllOutbound: true,
    });
    this.externalSecurityGroup.connections.allowFromAnyIpv4(Port.tcp(80));

    // Allow communication within the vpc for the app and envoy containers
    // - 8080: default app port for gateway and colorteller
    // - 9901: envoy admin interface, used for health check
    // - 15000: envoy ingress ports (egress over 15001 will be allowed by allowAllOutbound)
    this.internalSecurityGroup = new SecurityGroup(this, 'DemoInternalSG', {
      vpc: this.vpc,
      allowAllOutbound: true,
    });
    [ Port.tcp(this.APP_PORT), Port.tcp(9901), Port.tcp(15000) ].forEach(port => {
      this.internalSecurityGroup.connections.allowInternally(port);
    });
  }

  createCluster() {
    // Deploy a Fargate cluster on ECS
    this.cluster = new Cluster(this, 'DemoCluster', {
      vpc: this.vpc,
    });

    // Use Cloud Map for service discovery within the cluster, which
    // relies on either ECS Service Discovery or App Mesh integration
    // (default: cloudmap.NamespaceType.DNS_PRIVATE)
    this.namespace = "mesh.local";
    this.cluster.addDefaultCloudMapNamespace({
      name: this.namespace,
    });

    // IAM role for the color app tasks
    this.taskRole = new Role(this, 'DemoTaskRole', {
      assumedBy: new ServicePrincipal('ecs-tasks.amazonaws.com'),
      managedPolicies: [
        ManagedPolicy.fromAwsManagedPolicyName('CloudWatchLogsFullAccess'),
        ManagedPolicy.fromAwsManagedPolicyName('AWSXRayDaemonWriteAccess'),
      ]
    });

    // IAM task execution role for the color app tasks to be able to pull images from ECR
    this.taskExecutionRole = new Role(this, 'demoTaskExecutionRole', {
      assumedBy: new ServicePrincipal('ecs-tasks.amazonaws.com'),
      managedPolicies: [
        ManagedPolicy.fromAwsManagedPolicyName('AmazonEC2ContainerRegistryReadOnly'),
      ]
    });
  }
  createGateway() {
    let gatewayTaskDef = new FargateTaskDefinition(this, 'GatewayTaskDef', {
      taskRole: this.taskRole,
      executionRole: this.taskExecutionRole,
      cpu: 512,
      memoryLimitMiB: 1024,
    });

    //repositoryarn: '226767807331.dkr.ecr.us-west-2.amazonaws.com/gateway:latest',
    let gatewayContainer = gatewayTaskDef.addContainer('app', {
      image: ContainerImage.fromRegistry(`subfuzion/colorgateway:${this.APP_TAG}`),
      environment: {
        SERVER_PORT: `${this.APP_PORT}`,
        COLOR_TELLER_ENDPOINT: `colorteller.${this.namespace}:${this.APP_PORT}`,
      },
      logging: new AwsLogDriver({
        streamPrefix: 'app',
      }),
    });
    gatewayContainer.addPortMappings({
      containerPort: this.APP_PORT,
    })

    let gatewayService = new FargateService(this, 'GatewayService', {
      cluster: this.cluster,
      serviceName: 'gateway',
      taskDefinition: gatewayTaskDef,
      desiredCount: 1,
      securityGroup: this.internalSecurityGroup,
      cloudMapOptions: {
        name: 'gateway',
      },
    });

    let healthCheck: HealthCheck = {
      "path": '/ping',
      "port": 'traffic-port',
      "interval": Duration.seconds(30),
      "timeout": Duration.seconds(5),
      "healthyThresholdCount": 2,
      "unhealthyThresholdCount": 2,
      "healthyHttpCodes": "200-499",
    };

    let alb = new ApplicationLoadBalancer(this, 'DemoALB', {
      vpc: this.vpc,
      internetFacing: true,
      securityGroup: this.externalSecurityGroup,
    });
    let albListener = alb.addListener('web', {
      port: 80,
    });
    albListener.addTargets('demotarget', {
      port: 80,
      targets: [ gatewayService ],
      healthCheck: healthCheck,
    });
  }

  createColorTeller(...colors: string[]) {
    colors.forEach(color => {
      let taskDef = new FargateTaskDefinition(this, `${color}_taskdef`, {
        taskRole: this.taskRole,
        executionRole: this.taskExecutionRole,
        cpu: 512,
        memoryLimitMiB: 1024,
      });

      let container = taskDef.addContainer('app', {
        image: ContainerImage.fromRegistry(`subfuzion/colorteller:${this.APP_TAG}`),
        environment: {
          SERVER_PORT: `${this.APP_PORT}`,
          COLOR: color,
        },
        logging: new AwsLogDriver({
          streamPrefix: 'app',
        }),
      });
      container.addPortMappings({
        containerPort: this.APP_PORT,
      })

      let service = new FargateService(this, `ColorTellerService-${color}`, {
        cluster: this.cluster,
        serviceName: 'colorteller-${color}',
        taskDefinition: taskDef,
        desiredCount: 1,
        securityGroup: this.internalSecurityGroup,
        cloudMapOptions: {
          name: `colorteller-${color}`,
          dnsTtl: Duration.minutes(1)
        },
      });
    });

  }

}
