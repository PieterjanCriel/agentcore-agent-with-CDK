#!/usr/bin/env node
import assert from 'assert';

import * as cdk from 'aws-cdk-lib';
import { AgentcoreCdkStack } from '../lib/agentcore-cdk-stack';

const env = {
    account: process.env.CDK_DEFAULT_ACCOUNT || process.env.AWS_PROFILE?.split('-')[1],
    region: process.env.CDK_DEFAULT_REGION || process.env.AWS_REGION || process.env.AWS_DEFAULT_REGION,
};

assert(env.account, 'Missing account environment variable');
assert(env.region, 'Missing region environment variable');

const app = new cdk.App();
new AgentcoreCdkStack(app, 'AgentcoreCdkStack', {
    env: env,
    description: 'AgentCoreCDK',
    tags: {
        Project: 'AgentCore',
    },
});
