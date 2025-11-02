from bedrock_agentcore.runtime import BedrockAgentCoreApp

from fence.agents.bedrock import BedrockAgent
from fence.models.bedrock import NovaPro

app = BedrockAgentCoreApp()

editor_agent = BedrockAgent(
    identifier="editor_agent",
    model=NovaPro(region="eu-central-1", cross_region="eu"),
    description="You are a personal coach to help a user improve their writing skills.",
    tools=[],
    mcp_clients=[]
)

@app.entrypoint
def invoke(payload):
    """Process user input and return a response"""
    user_message = payload.get("prompt", "Hello")
    response = editor_agent.run(user_message)
    return {"result": response.answer}

if __name__ == "__main__":
    app.run()
