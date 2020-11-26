#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { RetrospectorStack } from '../lib/retrospector-stack';

const app = new cdk.App();
new RetrospectorStack(app, 'RetrospectorStack');
