import * as path from "path";

import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { AgentRuntimeArtifact, Runtime } from '@aws-cdk/aws-bedrock-agentcore-alpha';

export class AgentcoreCdkStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const role = new cdk.aws_iam.Role(this, "AgentRole", {
      assumedBy: new cdk.aws_iam.ServicePrincipal("bedrock-agentcore.amazonaws.com"),
    });

    role.addToPolicy(new cdk.aws_iam.PolicyStatement({
      actions: ["bedrock:InvokeModel", "bedrock:InvokeModelWithResponseStream"],
      resources: ["*"],
    }));

    role.addManagedPolicy(
      cdk.aws_iam.ManagedPolicy.fromAwsManagedPolicyName("CloudWatchFullAccess")
    );

    const agentFenceRuntimeArtifact = AgentRuntimeArtifact.fromAsset(
      path.join(__dirname, '..', 'agents', 'simple-agent-python-fence')
    );

    const runtimeFenceAgent = new Runtime(this, "MySimpleAgentFence", {
      runtimeName: "myFenceAgent",
      executionRole: role,
      agentRuntimeArtifact: agentFenceRuntimeArtifact,
    });

    const agentNoFenceRuntimeArtifact = AgentRuntimeArtifact.fromAsset(
      path.join(__dirname, '..', 'agents', 'simple-agent-python-strands')
    );

    const runtimeNoFenceAgent = new Runtime(this, "MySimpleAgentStrands", {
      runtimeName: "myStrandsAgent",
      executionRole: role,
      agentRuntimeArtifact: agentNoFenceRuntimeArtifact,
    });


  }
}
