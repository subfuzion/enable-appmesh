import {Mesh, VirtualNode, VirtualRouter, VirtualService} from "@aws-cdk/aws-appmesh";
import appmesh = require("@aws-cdk/aws-appmesh");
import {Port, SecurityGroup, SubnetType, Vpc} from "@aws-cdk/aws-ec2";
import {Cluster, ContainerImage, FargateService, FargateTaskDefinition, LogDriver, Protocol} from "@aws-cdk/aws-ecs";
import {ApplicationLoadBalancer} from "@aws-cdk/aws-elasticloadbalancingv2";
import {ManagedPolicy, Role, ServicePrincipal} from "@aws-cdk/aws-iam";
import {LogGroup, RetentionDays} from "@aws-cdk/aws-logs";
import cloudmap = require("@aws-cdk/aws-servicediscovery")
import {CfnOutput, Construct, Duration, RemovalPolicy, Stack, StackProps} from "@aws-cdk/core";

/**
 * Deploys the resources necessary to demo the Color App *before* and *after* enabling App Mesh.
 * This stack deploys
 * - a vpc with private subnets in 2 AZs, and a public ALB
 * - the Color App (a gateway and two colorteller (blue & green) services)
 * - an App Mesh mesh (ready to go for mesh-enabling the app)
 */
export class MeshDemoStack extends Stack {

  // Demo customization
  //
  // Gateway
  // You can use either either of these:
  // - "226767807331.dkr.ecr.us-west-2.amazonaws.com/gateway:latest"
  // - "subfuzion/colorgateway:v2"
  // - your own image on Docker Hub or ECR for your own account
  readonly GatewayImage = "subfuzion/colorgateway:v2";

  // ColorTeller
  // You can use either either of these:
  // - "226767807331.dkr.ecr.us-west-2.amazonaws.com/colorteller:latest"
  // - "subfuzion/colorteller:v2"
  // - your own image on Docker Hub or ECR for your own account
  readonly ColorTellerImage = "subfuzion/colorteller:v2";

  // Gateway and ColorTeller server port
  readonly APP_PORT = 8080;

  // ColorTeller services to run
  readonly colors = ["blue", "green", "red"];

  // service domain / namespace
  readonly namespace: string = "mesh.local";

  // might want to experiment with different ttl during testing
  readonly DEF_TTL = Duration.seconds(10);
  //
  // end: Demo customization


  name: string;
  taskRole: Role;
  taskExecutionRole: Role;
  vpc: Vpc;
  cluster: Cluster;
  internalSecurityGroup: SecurityGroup;
  externalSecurityGroup: SecurityGroup;
  logGroup: LogGroup;
  mesh: Mesh;


  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    // store for convenience
    this.name = props && props.stackName ? props.stackName : "demo";

    this.createLogGroup();
    this.createVpc();
    this.createCluster();
    this.createMesh();
    let router = this.createVirtualRouter();
    let backend = this.createVirtualService(router);
    this.createGateway([backend]);
    let virtualNodes = this.createColorTeller(...this.colors);
    this.createRoute(router, virtualNodes);    
  }

  createLogGroup() {
    this.logGroup = new LogGroup(this, "LogGroup", {
      logGroupName: this.name,
      retention: RetentionDays.ONE_DAY,
      removalPolicy: RemovalPolicy.DESTROY,
    });
  }

  createVpc() {
    // The VPC will have 2 AZs, 2 NAT gateways, and an internet gateway
    this.vpc = new Vpc(this, "VPC", {
      cidr: "10.0.0.0/16",
      maxAzs: 2,
      subnetConfiguration: [
        {
          cidrMask: 24,
          name: "ingress",
          subnetType: SubnetType.PUBLIC,
        },
        {
          cidrMask: 24,
          name: "application",
          subnetType: SubnetType.PRIVATE,
        },
      ],
    });

    // Allow public inbound web traffic on port 80
    this.externalSecurityGroup = new SecurityGroup(this, "ExternalSG", {
      vpc: this.vpc,
      allowAllOutbound: true,
    });
    this.externalSecurityGroup.connections.allowFromAnyIpv4(Port.tcp(80));

    // Allow communication within the vpc for the app and envoy containers
    // inbound 8080, 9901, 15000; all outbound
    // - 8080: default app port for gateway and colorteller
    // - 9901: envoy admin interface, used for health check
    // - 15000: envoy ingress ports (egress over 15001 will be allowed by allowAllOutbound)
    this.internalSecurityGroup = new SecurityGroup(this, "InternalSG", {
      vpc: this.vpc,
      allowAllOutbound: true,
    });
    [Port.tcp(this.APP_PORT), Port.tcp(9901), Port.tcp(15000)].forEach(port => {
      this.internalSecurityGroup.connections.allowInternally(port);
    });
  }

  createCluster() {
    // Deploy a Fargate cluster on ECS
    this.cluster = new Cluster(this, "Cluster", {
      vpc: this.vpc,
    });

    // Use Cloud Map for service discovery within the cluster, which
    // relies on either ECS Service Discovery or App Mesh integration
    // (default: cloudmap.NamespaceType.DNS_PRIVATE)
    let ns = this.cluster.addDefaultCloudMapNamespace({
      name: this.namespace,
    });
    // we need to ensure the service record is created for after we enable app mesh
    // (there is no resource we create here that will make this happen implicitly
    // since CDK won't all two services to register the same service name in
    // Cloud Map, even though we can discriminate between them using service attributes
    // based on ECS_TASK_DEFINITION_FAMILY
    // let serviceName = new Service(this, "colorteller", {
    //   name: 'colorteller',
    //   namespace: ns,
    //   dnsTtl: this.DEF_TTL,
    // });
    // serviceName.dependsOn(ns);

    // grant cloudwatch and xray permissions to IAM task role for color app tasks
    this.taskRole = new Role(this, "TaskRole", {
      assumedBy: new ServicePrincipal("ecs-tasks.amazonaws.com"),
      managedPolicies: [
        ManagedPolicy.fromAwsManagedPolicyName("CloudWatchLogsFullAccess"),
        ManagedPolicy.fromAwsManagedPolicyName("AWSXRayDaemonWriteAccess"),
        ManagedPolicy.fromAwsManagedPolicyName("AWSAppMeshEnvoyAccess"),
      ],
    });

    // grant ECR pull permission to IAM task execution role for ECS agent
    this.taskExecutionRole = new Role(this, "TaskExecutionRole", {
      assumedBy: new ServicePrincipal("ecs-tasks.amazonaws.com"),
      managedPolicies: [
        ManagedPolicy.fromAwsManagedPolicyName("AmazonEC2ContainerRegistryReadOnly"),
      ],
    });
    // CDK will print after finished deploying stack
    new CfnOutput(this, "ClusterName", {
      description: "ECS/Fargate cluster name",
      value: this.cluster.clusterName,
    });
  }

  createGateway(backends: VirtualService[]) {
    let gatewayTaskDef = new FargateTaskDefinition(this, "GatewayTaskDef", {
      family: "gateway",
      taskRole: this.taskRole,
      executionRole: this.taskExecutionRole,
      cpu: 512,
      memoryLimitMiB: 1024,
    });

    // let envoyContainer = gatewayTaskDef.addContainer("envoy", {
    //   image: ContainerImage.fromRegistry("subfuzion/aws-appmesh-envoy:v1.11.1.1"),
    //   user: "1337",
    //   memoryLimitMiB: 500,
    //   healthCheck: {
    //     command: [
    //       "CMD-SHELL",
    //       "curl -s http://localhost:9901/server_info | grep state | grep -q LIVE",
    //     ],
    //     interval: Duration.seconds(5),
    //     timeout: Duration.seconds(2),
    //     startPeriod: Duration.seconds(10),
    //     retries: 3,
    //   },
    //   environment: {
    //     APPMESH_VIRTUAL_NODE_NAME: "mesh/demo/virtualNode/gateway-vn",
    //     ENABLE_ENVOY_XRAY_TRACING: "1",
    //     ENABLE_ENVOY_STATS_TAGS: "1",
    //   },
    // });
    // envoyContainer.addPortMappings({
    //   containerPort: 9901,
    // }, {
    //   containerPort: 15000,
    // });
    // TODO: doesn't work right now
    // let cfnTaskDef = gatewayTaskDef.node.findChild("Resource") as CfnTaskDefinition;
    // cfnTaskDef.addPropertyOverride("ProxyConfiguration", {
    //   Type: "APPMESH",
    //   ContainerName: "envoy",
    //   ProxyConfigurationProperties: [
    //     {
    //       Name: "AppPorts",
    //       Value: [this.APP_PORT],
    //     },
    //     {
    //       Name: "IgnoredUID",
    //       Value: "1337",
    //     },
    //     {
    //       Name: "ProxyIngressPort",
    //       Value: "15000",
    //     },
    //     {
    //       Name: "ProxyEgressPort",
    //       Value: "15001",
    //     },
    //     {
    //       Name: "EgressIgnoredIPs",
    //       Value: "169.254.170.2,169.254.169.254",
    //     },
    //   ],
    // });

    let gatewayContainer = gatewayTaskDef.addContainer("app", {
      image: ContainerImage.fromRegistry(this.GatewayImage),
      environment: {
        SERVER_PORT: `${this.APP_PORT}`,
        COLOR_TELLER_ENDPOINT: `colorteller.${this.namespace}:${this.APP_PORT}`,
      },
      logging: LogDriver.awsLogs({
        logGroup: this.logGroup,
        streamPrefix: "gateway",
      }),
    });
    gatewayContainer.addPortMappings({
      containerPort: this.APP_PORT,
    });

    let xrayContainer = gatewayTaskDef.addContainer("xray", {
      image: ContainerImage.fromRegistry("amazon/aws-xray-daemon"),
      user: "1337",
      memoryReservationMiB: 256,
      cpu: 32,
    });
    xrayContainer.addPortMappings({
      containerPort: 2000,
      protocol: Protocol.UDP,
    });

    let gatewayService = new FargateService(this, "GatewayService", {
      cluster: this.cluster,
      serviceName: "gateway",
      taskDefinition: gatewayTaskDef,
      desiredCount: 1,
      securityGroup: this.internalSecurityGroup,
      cloudMapOptions: {
        name: "gateway",
        dnsTtl: this.DEF_TTL,
      },
    });

    this.createVirtualNodes("gateway", gatewayService.cloudMapService, backends);

    let alb = new ApplicationLoadBalancer(this, "PublicALB", {
      vpc: this.vpc,
      internetFacing: true,
      securityGroup: this.externalSecurityGroup,
    });
    let albListener = alb.addListener("web", {
      port: 80,
    });
    albListener.addTargets("Target", {
      port: 80,
      targets: [gatewayService],
      healthCheck: {
        path: "/ping",
        port: "traffic-port",
        interval: Duration.seconds(10),
        timeout: Duration.seconds(5),
        "healthyHttpCodes": "200-499",
        healthyThresholdCount: 2,
        unhealthyThresholdCount: 2,
      },
    });
    // CDK will print after finished deploying stack
    new CfnOutput(this, "URL", {
      description: "Color App public URL",
      value: alb.loadBalancerDnsName,
    });
  }

  // TODO: need to factor out all the duplicated code (DRY!) between this and createGateway...
  createColorTeller(...colors: string[]) {
    let create = (color: string, serviceName: string): VirtualNode => {
      let taskDef = new FargateTaskDefinition(this, `${color}_taskdef-v2`, {
        family: color,
        taskRole: this.taskRole,
        executionRole: this.taskExecutionRole,
        cpu: 512,
        memoryLimitMiB: 1024,
      });

      // let envoyContainer = taskDef.addContainer("envoy", {
      //   image: ContainerImage.fromRegistry("subfuzion/aws-appmesh-envoy:v1.11.1.1"),
      //   user: "1337",
      //   memoryLimitMiB: 500,
      //   healthCheck: {
      //     command: [
      //       "CMD-SHELL",
      //       "curl -s http://localhost:9901/server_info | grep state | grep -q LIVE",
      //     ],
      //     interval: Duration.seconds(5),
      //     timeout: Duration.seconds(2),
      //     startPeriod: Duration.seconds(10),
      //     retries: 3,
      //   },
      //   environment: {
      //     "APPMESH_VIRTUAL_NODE_NAME": `mesh/demo/virtualNode/${color}-vn`,
      //     "ENABLE_ENVOY_XRAY_TRACING": "1",
      //     "ENABLE_ENVOY_STATS_TAGS": "1",
      //     "ENVOY_LOG_LEVEL": "debug",
      //   },
      // });
      // envoyContainer.addPortMappings({
      //   containerPort: 9901,
      // }, {
      //   containerPort: 15000,
      // });
      // TODO: doesn't work right now
      // let cfnTaskDef = gatewayTaskDef.node.findChild("Resource") as CfnTaskDefinition;
      // cfnTaskDef.addPropertyOverride("ProxyConfiguration", {
      //   Type: "APPMESH",
      //   ContainerName: "envoy",
      //   ProxyConfigurationProperties: [
      //     {
      //       Name: "AppPorts",
      //       Value: [this.APP_PORT],
      //     },
      //     {
      //       Name: "IgnoredUID",
      //       Value: "1337",
      //     },
      //     {
      //       Name: "ProxyIngressPort",
      //       Value: "15000",
      //     },
      //     {
      //       Name: "ProxyEgressPort",
      //       Value: "15001",
      //     },
      //     {
      //       Name: "EgressIgnoredIPs",
      //       Value: "169.254.170.2,169.254.169.254",
      //     },
      //   ],
      // });

      let container = taskDef.addContainer("app", {
        image: ContainerImage.fromRegistry(this.ColorTellerImage),
        environment: {
          SERVER_PORT: `${this.APP_PORT}`,
          COLOR: color,
        },
        logging: LogDriver.awsLogs({
          logGroup: this.logGroup,
          streamPrefix: `colorteller-${color}`,
        }),
      });
      container.addPortMappings({
        containerPort: this.APP_PORT,
      });

      let xrayContainer = taskDef.addContainer("xray", {
        image: ContainerImage.fromRegistry("amazon/aws-xray-daemon"),
        user: "1337",
        memoryReservationMiB: 256,
        cpu: 32,
      });
      xrayContainer.addPortMappings({
        containerPort: 2000,
        protocol: Protocol.UDP,
      });

      let service = new FargateService(this, `ColorTellerService-${color}`, {
        cluster: this.cluster,
        serviceName: serviceName,
        taskDefinition: taskDef,
        desiredCount: 1,
        securityGroup: this.internalSecurityGroup,
        cloudMapOptions: {
          // overloading discovery name is possible, but unfortunately CDK doesn't support
          // name: "colorteller",
          name: serviceName,
          dnsTtl: this.DEF_TTL,
        },
      });

      return this.createVirtualNodes(color, service.cloudMapService);
    };
    let nodes = new Array<VirtualNode>();
    // initial color is a special case; before we enable app mesh, gateway
    // needs to reference an actual colorteller.mesh.local service (COLOR_TELLER_ENDPOINT);
    // the other colors need a unique namespace for now because CDK won't
    // allow reusing the same service name (although we can do this without
    // CDK; this is supported by Cloud Map / App Mesh, which uses Cloud
    // Map attributes for ECS service discovery: ECS_TASK_DEFINITION_FAMILY
    nodes.push(create(colors[0], "colorteller"));
    colors.slice(1).forEach(color => {
      nodes.push(create(color, `colorteller-${color}`));
    });
    return nodes;
  }

  createMesh() {
    this.mesh = new Mesh(this, "Mesh", {
      // use the same name to make it easy to identify the stack it's associated with
      meshName: this.name,
    });
  }

  createVirtualNodes(name: string, service?: cloudmap.IService, backends?: appmesh.IVirtualService[]) {
    let nodeName = `${name}-vn`;
    let node = this.mesh.addVirtualNode(nodeName, {
      virtualNodeName: nodeName,
      cloudMapService: service,
      cloudMapServiceInstanceAttributes: {
        ECS_TASK_DEFINITION_FAMILY: name
      },
      listener: {
        portMapping: {
          protocol: appmesh.Protocol.HTTP,
          port: this.APP_PORT,
        },
        healthCheck: {
          healthyThreshold: 2,
          interval: Duration.seconds(10),
          path: "/ping",
          port: this.APP_PORT,
          protocol: appmesh.Protocol.HTTP,
          timeout: Duration.seconds(5),
          unhealthyThreshold: 2
        }
      },
      backends
    });
    return node;
  }

  createVirtualRouter(): VirtualRouter {
    let router = this.mesh.addVirtualRouter("ColorTellerVirtualRouter", {
      listener: {
        portMapping: {
          port: this.APP_PORT,
          protocol: appmesh.Protocol.HTTP
        }
      },
      virtualRouterName: "colorteller-vr"
    });
    return router;
  }

  createRoute(router: VirtualRouter, virtualNodes: VirtualNode[]) {
    router.addRoute("ColorRoute", {
      routeName: "color-route",
      routeTargets: [
        {
          virtualNode: virtualNodes[0],
          weight: 1
        },
        {
          virtualNode: virtualNodes[1],
          weight: 1
        }
      ],
      routeType: appmesh.RouteType.HTTP,
      prefix: "/"
    });
  }

  createVirtualService(router: VirtualRouter) {
    let svc = this.mesh.addVirtualService("ColorTellerVirtualService", {
      virtualServiceName: `colorteller.${this.namespace}`,
      virtualRouter: router
    });
    return svc;
  }

}