import * as path from 'path';

import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { AgentRuntimeArtifact, Memory, MemoryStrategy, Runtime } from '@aws-cdk/aws-bedrock-agentcore-alpha';

export class GolangAgentStack extends cdk.Stack {
    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const role = new cdk.aws_iam.Role(this, 'AgentRole', {
            assumedBy: new cdk.aws_iam.ServicePrincipal('bedrock-agentcore.amazonaws.com'),
        });

        role.addToPolicy(
            new cdk.aws_iam.PolicyStatement({
                actions: ['bedrock:InvokeModel', 'bedrock:InvokeModelWithResponseStream'],
                resources: ['*'],
            }),
        );

        role.addManagedPolicy(cdk.aws_iam.ManagedPolicy.fromAwsManagedPolicyName('CloudWatchFullAccess'));

        const agentGoRuntimeArtifact = AgentRuntimeArtifact.fromAsset(
            path.join(__dirname, '..', 'agents', 'simple-agent-go-vanilla'),
        );

        const memory = new Memory(this, 'GoAgentMemory', {
            memoryName: 'goAgentMemory',
            description: 'Memory for Go Agent',
            expirationDuration: cdk.Duration.days(30),
            memoryStrategies: [
                MemoryStrategy.usingBuiltInSemantic(),
                MemoryStrategy.usingBuiltInSummarization(),
                MemoryStrategy.usingBuiltInUserPreference(),

            ]
        });

        // output the memory id
        new cdk.CfnOutput(this, 'MemoryId', {
            value: memory.memoryId,
        });

        const runtimeGoAgent = new Runtime(this, 'MySimpleAgentGo', {
            runtimeName: 'myGoAgent',
            executionRole: role,
            agentRuntimeArtifact: agentGoRuntimeArtifact,
            environmentVariables: {
                MEMORY_ID: memory.memoryId,
                MEMORY_USER_PREFERENCES_STRATEGY_NAME: memory.memoryStrategies.filter((s) => s.strategyType === 'USER_PREFERENCE')[0].name
            },
        });
    }
}
