#!/usr/bin/env node
import 'source-map-support/register';
import cdk = require('@aws-cdk/core');
import { MeshdemoStack } from '../lib/meshdemo-stack';

const app = new cdk.App();
new MeshdemoStack(app, 'MeshdemoStack');
