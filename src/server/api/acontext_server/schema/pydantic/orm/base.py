from pydantic import BaseModel
from datetime import datetime
from uuid import UUID


class ProjectMixin(BaseModel):
    project_id: UUID


class ProjectRow(BaseModel):
    configs: dict


class SpaceRow(ProjectMixin):
    configs: dict
