#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { RetrocogniteStack } from '../lib/retrocognite-stack';

const app = new cdk.App();
new RetrocogniteStack(app, 'RetrocogniteStack');
