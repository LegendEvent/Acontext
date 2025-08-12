from pydantic import BaseModel


class Env(BaseModel):
    llm_api_key: str
    llm_base_url: str = "https://api.openai.com/v1"
    llm_main_model: str = "gpt-4.1"
