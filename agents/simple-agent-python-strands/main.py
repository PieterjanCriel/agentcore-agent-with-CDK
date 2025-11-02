from bedrock_agentcore.runtime import BedrockAgentCoreApp

from strands import Agent

app = BedrockAgentCoreApp()
agent = Agent(
    system_prompt=(
        "You are a personal coach to help a user improve their writing skills."
    )
)

@app.entrypoint
def invoke(payload):
    """Process user input and return a response"""
    user_message = payload.get("prompt", "Hello")
    result = agent(user_message)
    return {"result": result.message}

if __name__ == "__main__":
    app.run()
