# Docker CI/CD Setup

This repository is configured with GitHub Actions to automatically build and deploy the application to DockerHub.

## Prerequisites

To use the automated Docker builds, you need to configure the following secrets in your GitHub repository:

### GitHub Secrets Setup

1. Go to your GitHub repository
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Add the following repository secrets:

#### Required Secrets

- **DOCKERHUB_USERNAME**: Your DockerHub username
- **DOCKERHUB_TOKEN**: DockerHub Personal Access Token

### Creating DockerHub Personal Access Token

1. Log in to [DockerHub](https://hub.docker.com/)
2. Go to **Account Settings** → **Security**
3. Click **New Access Token**
4. Give it a name (e.g., "GitHub Actions")
5. Set permissions to **Read, Write, Delete**
6. Copy the generated token and add it as `DOCKERHUB_TOKEN` secret

## Workflow

The CI/CD pipeline will:

1. **Multi-stage Docker Build**: 
   - **Frontend Stage**: Builds the Vue.js frontend using Node.js and npm
   - **Backend Stage**: Builds the Go backend and embeds the frontend assets
   - **Final Stage**: Creates a lightweight Alpine-based production image
2. **Push to DockerHub**: Pushes the image to `sivanov2018/new_billing` repository

### Triggered On

- Push to `main` or `master` branch
- Pull requests to `main` or `master` branch

### Docker Tags

- `latest`: Latest version from the default branch
- `sha-<commit>`: Tagged with Git commit SHA
- Branch/PR names for non-default branches

## Manual Docker Commands

To build and run locally:

```bash
# Build Docker image (includes frontend build)
docker build -t new-billing .

# Run container
docker run -p 8080:8080 new-billing
```

The Docker build automatically handles the frontend build process, so you don't need to build the frontend manually.

## Pull from DockerHub

Once the pipeline runs, you can pull the image:

```bash
docker pull sivanov2018/new_billing:latest
docker run -p 8080:8080 sivanov2018/new_billing:latest
```