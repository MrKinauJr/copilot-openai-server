# CI/CD Workflows

This repository includes two GitHub Actions workflows for continuous integration and deployment.

## CI Workflow

Located at `.github/workflows/ci.yml`, this workflow runs on every push to `main` and on pull requests.

### What it does:
- Sets up Go 1.23.0 environment
- Downloads and verifies dependencies (with caching enabled)
- Checks code formatting with `gofmt`
- Runs static analysis with `go vet`
- Executes tests with race detection and coverage
- Uploads coverage to Codecov
- Builds the application
- Uploads build artifacts

### Configuration

#### Codecov Integration (Optional)
To enable code coverage reporting to Codecov:
1. Sign up at https://codecov.io
2. Add your repository
3. Get your Codecov token
4. Add the token as a repository secret named `CODECOV_TOKEN`:
   - Go to repository Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Name: `CODECOV_TOKEN`
   - Value: your Codecov token

Coverage upload will fail silently if the token is not configured.

## Docker Workflow

Located at `.github/workflows/docker.yml`, this workflow builds and publishes Docker images.

### What it does:
- Builds multi-stage Docker images
- Pushes to GitHub Container Registry (ghcr.io)
- Generates appropriate tags based on the trigger:
  - `latest` for main branch
  - `v1.2.3`, `v1.2`, `v1` for version tags
  - PR number for pull requests
  - Commit SHA for all events

### Triggers:
- Push to `main` branch
- Version tags (e.g., `v1.0.0`)
- Pull requests (build only, no push)
- GitHub releases

### Using the Docker Image

After the workflow runs, images are available at:
```
ghcr.io/mrkinaujr/copilot-openai-server:latest
```

To pull and run:
```bash
docker pull ghcr.io/mrkinaujr/copilot-openai-server:latest
docker run -e GH_TOKEN=your_token -p 8080:8080 ghcr.io/mrkinaujr/copilot-openai-server:latest
```

## Running Workflows

Workflows run automatically on the specified triggers. You can also:
- Manually trigger them from the Actions tab (if `workflow_dispatch` is enabled)
- View workflow runs and logs in the Actions tab
- Download build artifacts from successful CI runs

## Workflow Badges

Add these badges to your README to show build status:

```markdown
![CI](https://github.com/MrKinauJr/copilot-openai-server/workflows/CI/badge.svg)
![Docker](https://github.com/MrKinauJr/copilot-openai-server/workflows/Docker%20Build%20and%20Publish/badge.svg)
```
