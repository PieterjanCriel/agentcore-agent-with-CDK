# Deploying an Agent with AWS AgentCore using CDK

This repo shows how to deploy a **Fence** or **Strands** agent on **AWS AgentCore** using **AWS CDK** instead of the CLI toolkit.

> ⚠️ CDK support for AgentCore is experimental and may change.

---

## Why CDK?

CDK makes AgentCore deployments **versioned**, **reproducible**, and **maintainable**.  
It handles Docker builds, ECR uploads, IAM, and CloudWatch automatically — perfect for scaling from local tests to production.

---

## Agent Examples

### Fence
```python
from bedrock_agentcore.runtime import BedrockAgentCoreApp
from fence.agents.bedrock import BedrockAgent
from fence.models.bedrock import NovaPro

app = BedrockAgentCoreApp()
agent = BedrockAgent(
  identifier="editor_agent",
  model=NovaPro(region="eu-central-1", cross_region="eu"),
  description="Coach users to improve their writing.",
  tools=[], mcp_clients=[]
)

@app.entrypoint
def invoke(payload):
    user_message = payload.get("prompt", "Hello")
    return {"result": agent.run(user_message).answer}
```

### Strands

```python
from bedrock_agentcore.runtime import BedrockAgentCoreApp
from strands import Agent

app = BedrockAgentCoreApp()
agent = Agent(system_prompt="Coach users to improve writing.")

@app.entrypoint
def invoke(payload):
    user_message = payload.get("prompt", "Hello")
    return {"result": agent(user_message).message}
```

### Structure

```
agentcore-cdk/
├── bin/
├── lib/
└── agents/
    ├── simple-agent-python-fence/
    └── simple-agent-python-strands/

```

Each agent has its own Dockerfile and requirements.txt.
CDK builds and deploys them as AgentCore runtimes with IAM roles, ECR, and CloudWatch.

### Deploy

```bash
npm install
npm run cdk bootstrap
npm run cdk deploy
```

After deployment, CDK outputs your runtime ARNs to invoke via Bedrock.