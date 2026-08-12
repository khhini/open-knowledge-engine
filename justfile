# Enterprise Open Knowledge Engine (OKF v0.2) - Justfile
# Run `just` or `just help` to list all available commands.

default:
    @just --list

# ------------------------------------------------------------------------------
# Local Development & Testing
# ------------------------------------------------------------------------------

# Run local development server
dev:
    go run cmd/server/main.go

# Run unit and integration tests
test:
    go test -v ./...

# Build local Go executable binary
build:
    @mkdir -p bin
    go build -o bin/server cmd/server/main.go

# Format and tidy Go code
fmt:
    go fmt ./...
    go mod tidy

# ------------------------------------------------------------------------------
# Docker Operations
# ------------------------------------------------------------------------------

# Build local Docker image
docker-build TAG="latest":
    docker build -t open-knowledge-engine:{{TAG}} .

# Run local Docker container
docker-run PORT="8080" TAG="latest":
    docker run -p {{PORT}}:8080 -e PORT=8080 open-knowledge-engine:{{TAG}}

# ------------------------------------------------------------------------------
# Google Cloud Run Deployment
# ------------------------------------------------------------------------------

# Configure GCP project and enable required Cloud Run & Container APIs
gcp-init PROJECT_ID:
    gcloud config set project {{PROJECT_ID}}
    gcloud services enable \
        run.googleapis.com \
        artifactregistry.googleapis.com \
        cloudbuild.googleapis.com

# Deploy directly from source code to Google Cloud Run (Recommended)
deploy-source PROJECT_ID REGION="asia-southeast2" SERVICE_NAME="open-knowledge-engine":
    gcloud run deploy {{SERVICE_NAME}} \
        --project={{PROJECT_ID}} \
        --region={{REGION}} \
        --source=. \
        --platform=managed \
        --allow-unauthenticated \
        --port=8080 \
        --memory=512Mi \
        --cpu=1

# Deploy via Artifact Registry & Cloud Build container build pipeline
deploy-build PROJECT_ID REGION="asia-southeast2" SERVICE_NAME="open-knowledge-engine" REPO="okf-repo":
    # Ensure Artifact Registry repository exists
    -gcloud artifacts repositories create {{REPO}} \
        --repository-format=docker \
        --location={{REGION}} \
        --description="OKF Knowledge Engine Docker Repository" \
        --project={{PROJECT_ID}}
    # Build container image via Google Cloud Build
    gcloud builds submit \
        --tag {{REGION}}-docker.pkg.dev/{{PROJECT_ID}}/{{REPO}}/{{SERVICE_NAME}}:latest \
        --project={{PROJECT_ID}}
    # Deploy container image to Google Cloud Run
    gcloud run deploy {{SERVICE_NAME}} \
        --image={{REGION}}-docker.pkg.dev/{{PROJECT_ID}}/{{REPO}}/{{SERVICE_NAME}}:latest \
        --project={{PROJECT_ID}} \
        --region={{REGION}} \
        --platform=managed \
        --allow-unauthenticated \
        --port=8080 \
        --memory=512Mi \
        --cpu=1

# Tail live logs from deployed Google Cloud Run service
logs PROJECT_ID REGION="asia-southeast2" SERVICE_NAME="open-knowledge-engine":
    gcloud run services logs tail {{SERVICE_NAME}} \
        --project={{PROJECT_ID}} \
        --region={{REGION}}

# Describe deployed Google Cloud Run service status and obtain public service URL
status PROJECT_ID REGION="asia-southeast2" SERVICE_NAME="open-knowledge-engine":
    gcloud run services describe {{SERVICE_NAME}} \
        --project={{PROJECT_ID}} \
        --region={{REGION}}
