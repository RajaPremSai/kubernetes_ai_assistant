## Kubernetes AI Assistant (Go)

An AI‑powered Kubernetes YAML assistant written in Go. This CLI tool takes a natural‑language prompt (for example, “create a Deployment with 3 replicas of nginx with a ClusterIP Service”) and uses OpenAI (or Azure OpenAI / compatible APIs) to generate Kubernetes manifests. It can optionally apply the generated resources directly to your cluster using the Kubernetes API.

The assistant can:
- **Generate valid Kubernetes YAML** from English descriptions.
- **Leverage your cluster’s OpenAPI schema** (with function calling) to stay aligned with the actual API surface.
- **Prompt you for confirmation** before applying, or print raw YAML only.
- **Work with OpenAI, Azure OpenAI, or local/OpenAI‑compatible endpoints.**

---

## Features

- **Natural‑language to YAML**: Turn prompts into Kubernetes manifests.
- **Direct apply to cluster**:
  - Parses the generated YAML into unstructured objects.
  - Uses `client-go` with server‑side apply (`ApplyOptions`) to create/update resources.
  - Detects and uses the current context/namespace from your kubeconfig (or `--namespace` flag).
- **Schema‑aware generation (optional)**:
  - When `--use-k8s-api` is enabled, the model can call helper “functions”:
    - `findSchemaNames` — list fully‑qualified Kubernetes schema names that match a given resource name.
    - `getSchema` — retrieve the OpenAPI schema for a given resource type.
  - Schemas are fetched either from the Kubernetes API server (`kubectl get --raw /openapi/v2`) or from a configurable URL.
- **Flexible OpenAI configuration**:
  - Supports **OpenAI API** and **Azure OpenAI**.
  - Supports model mapping for Azure deployments.
- **Interactive confirmation**: Uses `promptui` to show the manifest and ask whether to apply.
- **Spinner and debug logging** for a nicer CLI experience.

---

## Requirements

- **Go**: `go 1.24+` (as specified in `go.mod`).
- **kubectl**: Required when using schema‑aware mode (`--use-k8s-api`) because it shells out to `kubectl get --raw /openapi/v2`.
- **Kubeconfig**:
  - Default: `~/.kube/config`
  - Or explicitly via standard kubeconfig flags (e.g. `--kubeconfig`, `--namespace`) exposed by `k8s.io/cli-runtime`.
- **OpenAI / Azure OpenAI / compatible API**:
  - A valid API key.
  - Correct endpoint configuration (see Configuration section).

---

## Installation

Clone the repository:

```bash
git clone https://github.com/RajaPremSai/kubernetes_ai_assistant.git
cd kubernetes_ai_assistant
```

Build the binary:

```bash
go build -o kubernetes_ai_assistant .
```

Optionally, move it somewhere on your `PATH`:

```bash
mv kubernetes_ai_assistant /usr/local/bin/
```

---

## Quick Start

1. **Set your OpenAI API key**:

```bash
export OPENAI_API_KEY="sk-..."
```

2. **Run the assistant with a prompt**:

```bash
./kubernetes_ai_assistant "Create a Deployment with 3 nginx replicas and expose it via a ClusterIP Service on port 80"
```

3. The tool will:
   - Call OpenAI (or compatible endpoint) with a prompt that instructs it to only produce Kubernetes YAML.
   - Show you the generated manifest in the terminal.
   - Ask whether you want to apply it to the current Kubernetes context.

4. Choose:
   - **Apply** — apply the manifest using server‑side apply.
   - **Don't Apply** — exit without making changes.

If you prefer to see only the YAML (without any prompt to apply), use the `--raw` flag:

```bash
./kubernetes_ai_assistant --raw "Create a ConfigMap named app-config with key 'ENV' set to 'production'"
```

---

## CLI Usage

### Basic invocation

```bash
kubernetes_ai_assistant [flags] <prompt...>
```

- **`prompt`** (required): One or more words describing the desired Kubernetes object(s). For example:
  - `"Create a StatefulSet with 3 replicas of redis and a headless Service"`
  - `"Give me a Job that runs a one‑off backup pod"`

If you run the command with no prompt, it will error with `prompt must be provided`.

### Key flags and environment variables

Most configuration is available both as **flags** and **environment variables** (using `github.com/walles/env` under the hood). Environment variables serve as defaults that flags can override.

- **OpenAI configuration**
  - **`--openai-api-key`** / **`OPENAI_API_KEY`** (required):
    - API key for OpenAI / Azure OpenAI / compatible service.
    - If missing, the program exits with: `Please provide an OpenAI key.`
  - **`--openai-endpoint`** / **`OPENAI_ENDPOINT`**:
    - Default: `https://api.openai.com/v1`
    - Set this to your custom endpoint (e.g. Azure OpenAI `https://<resource>.openai.azure.com/` or a local OpenAI‑compatible API).
  - **`--openai-deployment-name`** / **`OPENAI_DEPLOYMENT_NAME`**:
    - Name of the model/deployment to use.
    - Default: `gpt-3.5-turbo-0301`
  - **`--azure-openai-map`** / **`AZURE_OPENAI_MAP`**:
    - Map of model name → Azure deployment name, used when pointing at an Azure OpenAI endpoint.
    - Example env: `AZURE_OPENAI_MAP="gpt-3.5-turbo-0301=my-azure-deployment"`

- **Generation behavior**
  - **`--temperature`** / **`TEMPERATURE`**:
    - Float in \[0, 1\]. Default: `0.0`.
    - Lower values → more deterministic, less “creative” YAML.
  - **`--raw`**:
    - If set, prints the raw YAML and exits (no confirmation / apply step).

- **Confirmation & logging**
  - **`--require-confirmation`** / **`REQUIRE_CONFIRMATION`**:
    - Boolean, default `true`.
    - When `false`, the tool will **skip the interactive confirmation** and apply directly.
  - **`--debug`** / **`DEBUG`**:
    - Enables debug logging (using `logrus`).
    - Prints OpenAI endpoint, deployment name, Azure map, temperature, and Kubernetes API usage flags.

- **Kubernetes / schema integration**
  - **`--use-k8s-api`** / **`USE_K8S_API`**:
    - Boolean, default `false`.
    - When enabled, the chat model can call helper functions (`findSchemaNames`, `getSchema`) that pull in real OpenAPI schema information from your cluster or configured URL.
  - **`--k8s-openapi-url`** / **`K8S_OPENAPI_URL`**:
    - Optional URL for a Kubernetes OpenAPI v2 JSON document.
    - If empty, the assistant uses `kubectl get --raw /openapi/v2 --kubeconfig <path>` to fetch it from the current cluster.

- **Kubeconfig flags**
  - The command embeds `k8s.io/cli-runtime`’s standard kubeconfig flags via `genericclioptions.NewConfigFlags`.
  - You can use:
    - `--kubeconfig` to point to a specific kubeconfig file.
    - `--namespace` to override the namespace; otherwise, it uses the namespace in your current context or falls back to `default`.

---

## How It Works (High‑Level)

- **Entry point**:
  - `main.go` simply calls `cli.InitAndExecute()`.
  - `RootCmd()` sets up a Cobra command (`Use: kubernetes_ai_assistant`) and binds flags and standard kubeconfig flags.

- **Prompt → YAML**:
  - `gptCompletion` constructs a system‑style prompt that instructs the model to **only** generate Kubernetes YAML (no explanations, no fenced code blocks).
  - It retries with exponential backoff on 429 (Too Many Requests).
  - Depending on the configured deployment name:
    - Uses `CreateCompletion` for non‑chat models (`code-davinci-002`, `text-davinci-003`).
    - Uses `CreateChatCompletion` with optional function calling for chat models.

- **Optional function calling (schema mode)**:
  - When `--use-k8s-api` is enabled, the chat request includes function definitions:
    - `findSchemaNames` and `getSchema`.
  - The model may choose to call them; the tool will:
    - Parse the function arguments.
    - Fetch the relevant schema sections (via `kubectl` or `--k8s-openapi-url`).
    - Feed those back into the conversation until the model returns plain YAML.

- **Apply to cluster**:
  - The YAML text is:
    - Split into runtime objects using `yamlutil.NewYAMLOrJSONDecoder`.
    - Converted to `unstructured.Unstructured`.
    - Applied via a `dynamic.Interface` using `Apply` with server‑side apply.
  - Namespaces:
    - If the manifest has no namespace and the resource is namespaced, it injects the current context’s namespace (or `default`).

---

## Examples

- **Basic Deployment + Service**

```bash
kubernetes_ai_assistant \
  "Create a Deployment called web with 2 replicas of nginx:1.25 and a ClusterIP Service called web-svc on port 80 targeting containerPort 80"
```

- **ConfigMap and a Pod consuming it**

```bash
kubernetes_ai_assistant \
  "Create a ConfigMap named app-config with key LOG_LEVEL=debug and a Pod that mounts it as env vars"
```

- **Using schema‑aware mode against your current cluster**

```bash
export USE_K8S_API=true
kubernetes_ai_assistant \
  "Create a NetworkPolicy that only allows ingress on port 443 from pods labeled app=frontend"
```

---

## Azure OpenAI / Custom Endpoint Configuration

To use **Azure OpenAI**:

```bash
export OPENAI_API_KEY="azure-key"
export OPENAI_ENDPOINT="https://<your-resource>.openai.azure.com/"
export OPENAI_DEPLOYMENT_NAME="gpt-35-turbo"
export AZURE_OPENAI_MAP="gpt-35-turbo=my-azure-deployment-name"

kubernetes_ai_assistant "Create a Deployment for my Go app"
```

To use a **local or compatible OpenAI‑style API**, point `OPENAI_ENDPOINT` (or `--openai-endpoint`) at that URL and keep `OPENAI_API_KEY` as required by that service.

---

## Development

- **Dependencies** are managed via `go.mod` / `go.sum`.
- Key packages:
  - `github.com/sashabaranov/go-openai` — OpenAI / Azure OpenAI client.
  - `k8s.io/client-go`, `k8s.io/apimachinery`, `k8s.io/cli-runtime` — Kubernetes API clients and CLI utilities.
  - `github.com/spf13/cobra`, `github.com/spf13/pflag` — CLI framework.
  - `github.com/briandowns/spinner`, `github.com/manifoldco/promptui`, `github.com/sirupsen/logrus` — CLI UX and logging.

Run tests / build as usual:

```bash
go test ./...
go build ./...
```

---

## Caveats & Notes

- **Safety**:
  - The tool applies manifests directly to your cluster; ensure you are in the correct context and namespace.
  - Review the generated YAML carefully (especially in production environments).
- **Model limitations**:
  - Models can hallucinate or generate invalid manifests; schema‑aware mode helps but is not a guarantee.
  - Always validate critical manifests (e.g. via `kubectl apply --dry-run=server -f -`).
- **OpenAPI schema key**:
  - The code currently expects the schema’s definitions under the `definations` key (matching the spelling in the existing code); be aware of this if you supply a custom OpenAPI doc.

---

## License

This project is open source. See the repository for licensing details (or add a `LICENSE` file if not already present).

