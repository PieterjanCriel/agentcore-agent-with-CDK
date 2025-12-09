import boto3
import json
import os

agentcore_client = boto3.client("bedrock-agentcore", region_name=os.getenv("AWS_REGION", "eu-central-1"))

memory_id = os.environ["MEMORY_ID"]
actor_id = os.getenv("ACTOR_ID", "default_actor")
session_id = os.getenv("SESSION_ID", "1")

preference_memory_strategy_id = os.environ["PREFERENCE_MEMORY_STRATEGY_ID"]
summary_memory_strategy_id = os.environ["SUMMARY_MEMORY_STRATEGY_ID"]

def get_memory_records(memory_id, actor_id, memory_strategy_id):
    path = f"/strategies/{memory_strategy_id}/actors/{actor_id}"

    memory_records = agentcore_client.list_memory_records(
        memoryStrategyId=memory_strategy_id,
        memoryId=memory_id,
        namespace=path
        )
    
    return memory_records

def get_memory_records_with_session(memory_id, actor_id, memory_strategy_id, session_id):
    path = f"/strategies/{memory_strategy_id}/actors/{actor_id}/sessions/{session_id}"

    memory_records = agentcore_client.list_memory_records(
        memoryStrategyId=memory_strategy_id,
        memoryId=memory_id,
        namespace=path
        )
    
    return memory_records


print(f"Preferences for actor: {actor_id}")
preference_memory_records = get_memory_records(memory_id, actor_id, preference_memory_strategy_id)
for record in preference_memory_records["memoryRecordSummaries"]:
    # content = json.loads(record["content"]["text"])
    print(record)
    print("\n")

print("\n")

print(f"Summaries for actor and session: {actor_id}")
summary_memory_records = get_memory_records_with_session(memory_id, actor_id, summary_memory_strategy_id, session_id)
for record in summary_memory_records["memoryRecordSummaries"]:
    # content = json.loads(record["content"]["text"])
    print(record)
    print("\n")


