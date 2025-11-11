# Using Code Agent with Vertex AI - Quick Start Guide

## Your Project Details
- **Project ID**: mycurator-poc-475706
- **Backend**: Vertex AI (GCP project-based authentication)

## Prerequisites

1. **Install Google Cloud CLI** (if not already installed):
   ```bash
   curl https://sdk.cloud.google.com | bash
   exec -l $SHELL
   ```

2. **Set up Application Default Credentials (ADC)**:
   ```bash
   gcloud auth application-default login
   ```
   This will open a browser window to authenticate with your Google account.

3. **Set your default project** (optional but recommended):
   ```bash
   gcloud config set project mycurator-poc-475706
   ```

## Method 1: Using Environment Variables (Recommended)

Set these environment variables and run code-agent:

```bash
export GOOGLE_CLOUD_PROJECT=mycurator-poc-475706
export GOOGLE_CLOUD_LOCATION=us-central1
./code-agent
```

**What happens**:
- Auto-detects Vertex AI backend from GOOGLE_CLOUD_PROJECT variable
- Uses Application Default Credentials (ADC) for authentication
- Shows "gemini-2.5-flash (Vertex AI)" in the banner

**Available Locations**:
- `us-central1` (US, multi-region)
- `us-west1` (US, single region)
- `us-east1` (US, single region)
- `europe-west1` (Europe)
- `asia-southeast1` (Southeast Asia)
- And others depending on your project's enabled regions

## Method 2: Using CLI Flags (Explicit Control)

```bash
./code-agent --backend vertexai --project mycurator-poc-475706 --location us-central1
```

**What happens**:
- Explicitly selects Vertex AI backend
- Uses the specified project and location
- Overrides any environment variables

## Method 3: Quick Start (Automatic Detection)

If you've already run `gcloud auth application-default login` and set GOOGLE_CLOUD_PROJECT:

```bash
./code-agent
```

The system will automatically detect Vertex AI based on your environment configuration.

## Verification

When you start code-agent with Vertex AI, you should see:

```
╔═══════════════════════════════════════════════════════════════════╗
║                       Code Agent v1.0.0                           ║
║              🤖 gemini-2.5-flash (Vertex AI)                       ║
║                                                                    ║
║          Your AI coding assistant powered by Google ADK            ║
╚═══════════════════════════════════════════════════════════════════╝
```

## Troubleshooting

### Error: "Vertex AI backend requires GOOGLE_CLOUD_PROJECT"
**Solution**: Make sure environment variables are set:
```bash
export GOOGLE_CLOUD_PROJECT=mycurator-poc-475706
export GOOGLE_CLOUD_LOCATION=us-central1
```

### Error: "failed to create Vertex AI client"
**Possible causes**:
1. Not authenticated with ADC:
   ```bash
   gcloud auth application-default login
   ```

2. Vertex AI API not enabled in your project:
   ```bash
   gcloud services enable aiplatform.googleapis.com --project=mycurator-poc-475706
   ```

3. Project ID is incorrect - verify it:
   ```bash
   gcloud config list
   ```

### Error: "Permission denied" or "Access Denied"
**Solution**: Check your IAM permissions:
```bash
# Your user account needs these roles:
# - Vertex AI User (roles/aiplatform.user)
# - AI Platform User (roles/ml.user)

gcloud projects get-iam-policy mycurator-poc-475706
```

## Complete Setup Example

```bash
# 1. Authenticate with Google Cloud
gcloud auth application-default login

# 2. Set project
gcloud config set project mycurator-poc-475706

# 3. Enable Vertex AI API (if not already enabled)
gcloud services enable aiplatform.googleapis.com

# 4. Set environment variables
export GOOGLE_CLOUD_PROJECT=mycurator-poc-475706
export GOOGLE_CLOUD_LOCATION=us-central1

# 5. Run code-agent
cd /path/to/code_agent
./code-agent
```

## Using with Docker/Kubernetes

If running in a containerized environment:

```dockerfile
# In your Dockerfile
FROM golang:1.24

WORKDIR /app
COPY . .
RUN go build -o code-agent ./code_agent

# Set up service account authentication
ENV GOOGLE_APPLICATION_CREDENTIALS=/var/secrets/google/key.json
ENV GOOGLE_CLOUD_PROJECT=mycurator-poc-475706
ENV GOOGLE_CLOUD_LOCATION=us-central1

ENTRYPOINT ["./code-agent", "--backend", "vertexai"]
```

Mount your GCP service account key:
```bash
docker run -v ~/.config/gcloud/application_default_credentials.json:/var/secrets/google/key.json \
  -e GOOGLE_CLOUD_PROJECT=mycurator-poc-475706 \
  -e GOOGLE_CLOUD_LOCATION=us-central1 \
  code-agent:latest
```

## Features Available

With Vertex AI backend, you get:
- ✅ Gemini 2.5 Flash model
- ✅ Full tool ecosystem (file operations, terminal execution, code search)
- ✅ Session persistence (with database storage)
- ✅ Multi-session support
- ✅ Token usage tracking
- ✅ Streaming responses
- ✅ System prompt customization

## Next Steps

Once code-agent is running with Vertex AI:

1. **Get help**:
   ```
   ❯ /help
   ```

2. **List available tools**:
   ```
   ❯ /tools
   ```

3. **View system prompt**:
   ```
   ❯ /prompt
   ```

4. **Check token usage**:
   ```
   ❯ /tokens
   ```

5. **Natural language requests**:
   ```
   ❯ Create a README.md for this project
   ❯ Add error handling to main.go
   ❯ Run tests and fix any failures
   ```

## Additional Resources

- **Vertex AI Documentation**: https://cloud.google.com/vertex-ai/docs
- **Project Console**: https://console.cloud.google.com/vertex-ai?project=mycurator-poc-475706
- **Gemini API Docs**: https://ai.google.dev/
- **Code Agent Guide**: See USER_GUIDE.md in the project root

---

**Questions?** Run `./code-agent --help` to see all available flags and options.
