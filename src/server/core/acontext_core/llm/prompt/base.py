from abc import ABC, abstractmethod


class BasePrompt(ABC):
    @abstractmethod
    def system_prompt(self, *args, **kwargs) -> str:
        pass

    @abstractmethod
    def pack_task_input(self, *args, **kwargs) -> str:
        pass

    @abstractmethod
    def prompt_infos(self) -> str:
        pass
