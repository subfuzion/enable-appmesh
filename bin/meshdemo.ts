#!/usr/bin/env node
import 'source-map-support/register';
import cdk = require('@aws-cdk/core');
import { MeshDemoStack } from '../lib/mesh-demo-stack';


const app = new cdk.App();
new MeshDemoStack(app, 'demo');

