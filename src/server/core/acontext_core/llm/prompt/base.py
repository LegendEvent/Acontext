from ...schema.llm import ToolSchema


class BasePrompt:
    @classmethod
    def system_prompt(cls, *args, **kwargs) -> str:
        pass

    @classmethod
    def pack_task_input(cls, *args, **kwargs) -> str:
        pass

    @classmethod
    def prompt_kwargs(cls) -> str:
        pass

    @classmethod
    def tool_schema(cls) -> list[ToolSchema]:
        pass
