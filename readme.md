<div align="center">
    <a href="https://discord.gg/rpZs5TaSuV">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="./docs/logo/dark.svg">
      <img alt="Show Acontext logo" src="./docs/logo/light.svg" width="418">
    </picture>
  </a>
  <h4>Context Data Platform for Self-learning Agents</h4>
  <p>
    <a href="https://pypi.org/project/acontext/">
      <img src="https://img.shields.io/pypi/v/acontext.svg">
    </a>
    <a href="https://www.npmjs.com/package/@acontext/acontext">
      <img src="https://img.shields.io/npm/v/@acontext/acontext.svg?logo=npm&logoColor=fff&style=flat&labelColor=2C2C2C&color=28CF8D">
    </a>
    <a href="https://github.com/memodb-io/acontext/actions/workflows/core-test.yaml">
      <img src="https://github.com/memodb-io/acontext/actions/workflows/core-test.yaml/badge.svg">
    </a>
    <a href="https://github.com/memodb-io/acontext/actions/workflows/api-test.yaml">
      <img src="https://github.com/memodb-io/acontext/actions/workflows/api-test.yaml/badge.svg">
    </a>
    <a href="https://github.com/memodb-io/acontext/actions/workflows/cli-test.yaml">
      <img src="https://github.com/memodb-io/acontext/actions/workflows/cli-test.yaml/badge.svg">
    </a>
  </p>
  <p>
    <a href="https://discord.gg/rpZs5TaSuV">
      <img src="https://dcbadge.limes.pink/api/server/rpZs5TaSuV?style=flat">
    </a>
  </p>
</div>

Acontext is a context data platform that:

- Store contexts & artifacts, using postgres and S3
- Observe agents' tasks and user feedbacks, and offer an nice Dashboard
- Enable agents' self-learning by collecting experiences (or SOPs).

We're building it because we believe Acontext can help you to build a more scalable agent product, and improve it overtime to provide more values to your users.





## How to Start It?

[📖 doc](https://docs.acontext.io/local)

We have a `acontext-cli` to help you do quick proof-of-concept. Download it first in your terminal:

```bash
curl -fsSL https://install.acontext.io | sh
```

You should have [docker](https://www.docker.com/get-started/) installed, and an OpenAI API Key to start a Acontext backend in your computer:

```bash
acontext docker up
```

> [📖 doc](https://docs.acontext.io/settings/core)
>
> Acontext requires a llm provider and an embedding provider. 
>
> We support OpenAI and Anthropic SDK format and OpenAI and jina.ai embedding api format

Once it’s done, you can the following endpoints enabled:

- Acontext API Base URL: http://localhost:8029/api/v1
- Acontext Dashboard: http://localhost:3000/



## How to Use It?

We're maintaining Python [![pypi](https://img.shields.io/pypi/v/acontext.svg)](https://pypi.org/project/acontext/) and Typescript [![npm](https://img.shields.io/npm/v/@acontext/acontext.svg?logo=npm&logoColor=fff&style=flat&labelColor=2C2C2C&color=28CF8D)]("https://www.npmjs.com/package/@acontext/acontext") SDKs. Below snippets are using Python.

### Init Client

```python
from acontext import AcontextClient

client = AcontextClient(
    base_url="http://localhost:8029/api/v1"
    api_key="sk-ac-your-root-api-bearer-token"
)
client.ping()

# yes, the default api_key is sk-ac-your-root-api-bearer-token
```

> [📖 async client doc](https://docs.acontext.io/settings/core)



### Store

[📖 doc](https://docs.acontext.io/store)



### Observe

[📖 doc](https://docs.acontext.io/observe)



### Self-learning

[📖 doc](https://docs.acontext.io/learn)








Download Examples

You can choose some examples to play with, in an interactive way:

```bash
acontext create
```

Or you can just check every examples in [Acontext-Examples](https://github.com/memodb-io/Acontext-Examples)

