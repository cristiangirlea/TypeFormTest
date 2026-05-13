# Typeform Lite

A simplified version of Typeform's builder and renderer. This application allows users to create forms, add questions, and collect responses via a unique shareable URL.

## Features

### Form Builder
- **Create New Forms**: Initiate the creation of a new form.
- **Add Questions**: Add multiple questions to your form.
- **Save Forms**: Save forms with at least one question to generate a unique shareable URL.
- **View All Forms**: List all created forms and access their builder or shareable link.

### Form Renderer
- **Access via URL**: Open forms using their unique slug.
- **One-by-One Questions**: Answer questions one at a time for a better experience.
- **Collect Answers**: Responses are stored in the database.
- **Thank You Screen**: Simple completion message after the final question.

## Tech Stack

- **Backend**: Go 1.26
- **Frontend**: Next.js (React 19, Tailwind CSS)
- **Database**: SQLite
- **Deployment**: Docker & Docker Compose

## Getting Started

### Prerequisites

- [Docker](https://www.docker.com/) and Docker Compose
- Or locally:
  - [Go 1.26+](https://go.dev/)
  - [Node.js 20+](https://nodejs.org/)

### Running with Docker (Recommended)

The easiest way to start the entire stack is using Docker Compose.

#### Development (with hot-reload)
This setup uses `air` for Go backend hot-reload and Next.js dev server.
```bash
docker-compose up --build
```

- **Frontend**: [http://localhost:3000](http://localhost:3000)
- **Backend**: [http://localhost:8080](http://localhost:8080)

#### Production
This setup uses optimized multi-stage builds.
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up --build
```
*Note: Make sure to update `NEXT_PUBLIC_API_URL` in `docker-compose.prod.yml` to your actual domain.*

### Manual Development Setup

#### Backend

1. Navigate to the backend directory:
   ```bash
   cd backend
   ```
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Run the server:
   ```bash
   go run cmd/main.go
   ```
   *Note: By default, it uses `forms.db` in the root or specified via `DB_PATH` environment variable.*

#### Frontend

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start the development server:
   ```bash
   npm run dev
   ```

## API Endpoints

- `GET /health` - Health check
- `GET /forms` - List all forms
- `POST /forms` - Create a new form
- `GET /forms/{id}` - Get form details by ID
- `POST /forms/{id}/questions` - Add a question to a form
- `POST /forms/{id}/save` - Save form and generate shareable slug
- `GET /form/{slug}` - Get form details by shareable slug
- `POST /form/{slug}/responses` - Submit a form response

## Scaling to 50,000+ Concurrent Users

To handle 50,000+ users "instantly," the current architecture can be scaled as follows:

1. **Frontend**: Move to a CDN (Cloudflare, Vercel, or AWS CloudFront). This offloads all static asset delivery and initial page loads from the server.
2. **Backend**:
   - **Horizontal Scaling**: Run multiple instances of the Go backend container behind a Load Balancer (Nginx, HAProxy, or AWS ALB).
   - **Connection Limits**: Increase the operating system's open file limit (`ulimit -n 65535`) to allow 50k+ simultaneous sockets.
3. **Database**:
   - **SQLite to PostgreSQL**: While SQLite is fast for reads, 50k concurrent writes require a distributed database like PostgreSQL or MySQL.
   - **Caching**: Implement Redis to cache form definitions (`GET /form/{slug}`), which are the most frequent read operations. [DONE]
4. **Infrastructure**: Deploy on a container orchestrator like **Kubernetes** to manage scaling and health checks automatically.

### Current Caching Implementation
The application now includes a Redis caching layer for the public form renderer. 
- **Look-aside Caching**: The backend checks Redis before querying the database for form definitions by slug.
- **Cache Invalidation**: The cache is automatically invalidated when a form is updated or a new question is added.
- **Performance**: Serving from Redis allows the application to handle massive spikes in form views with sub-millisecond latency.
