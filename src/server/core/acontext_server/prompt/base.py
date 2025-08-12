from abc import ABC, abstractmethod


class BasePrompt(ABC):
    @abstractmethod
    def system_prompt(self, *args, **kwargs) -> str:
        pass

    @abstractmethod
    def task_input_prompt(self, *args, **kwargs) -> str:
        pass

    @abstractmethod
    def prompt_id(self) -> str:
        pass

    @abstractmethod
    def prompt_parameters(self) -> dict:
        pass
