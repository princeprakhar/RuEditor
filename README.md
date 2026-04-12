# RuEditor - LaTeX Resume Editor

A professional LaTeX resume editor with real-time collaboration features, inspired by Overleaf. Built with Go backend and React + TypeScript frontend, containerized with Docker.

## Features

- **LaTeX Editor**: Monaco-based editor with syntax highlighting and auto-completion
- **Real-time Collaboration**: Multiple users can edit the same document simultaneously
- **PDF Compilation**: Compile LaTeX to PDF with instant preview
- **User Authentication**: Secure JWT-based authentication
- **Document Management**: Create, edit, delete, and share documents
- **Modern UI**: Beautiful, responsive interface with Tailwind CSS

## Architecture

```
RuEditor/
├── backend/          # Go backend
│   ├── cmd/          # Main entry point
│   └── internal/     # Internal packages
│       ├── api/      # HTTP API handlers
│       ├── models/   # Data models
│       ├── services/ # Business logic (LaTeX compilation)
│       └── websocket/ # WebSocket manager for real-time
├── frontend/        # React + TypeScript frontend
│   ├── src/
│   │   ├── components/  # React components
│   │   ├── hooks/      # Custom hooks
│   │   ├── services/   # API and WebSocket services
│   │   ├── store/      # Redux store
│   │   └── types/      # TypeScript types
│   └── public/         # Static assets
├── docker-compose.yml  # Docker orchestration
└── README.md
```

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- 4GB+ RAM recommended for LaTeX compilation

### Running the Application

1. Clone the repository
2. Start all services:
   ```bash
   docker-compose up --build
   ```

3. Access the application:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - WebSocket: ws://localhost:8081

### Development Mode (Without Docker)

#### Backend
```bash
cd backend
go mod download
go run ./cmd/server
```

#### Frontend
```bash
cd frontend
npm install
npm start
```

## Environment Variables

### Backend
| Variable | Description | Default |
|----------|-------------|---------|
| PORT | HTTP server port | 8080 |
| WS_PORT | WebSocket port | 8081 |
| JWT_SECRET | JWT signing secret | supersecretkey123 |
| DATABASE_URL | PostgreSQL connection string | postgres://... |
| REDIS_URL | Redis connection string | redis://... |

### Frontend
| Variable | Description | Default |
|----------|-------------|---------|
| REACT_APP_API_URL | Backend API URL | http://localhost:8080 |
| REACT_APP_WS_URL | WebSocket URL | ws://localhost:8081 |

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user

### Documents
- `GET /api/v1/documents` - List user documents
- `POST /api/v1/documents` - Create new document
- `GET /api/v1/documents/:id` - Get document
- `PUT /api/v1/documents/:id` - Update document
- `DELETE /api/v1/documents/:id` - Delete document
- `POST /api/v1/documents/:id/compile` - Compile LaTeX to PDF
- `POST /api/v1/documents/:id/collaborate` - Invite collaborator

### WebSocket
- `ws://host/ws?document_id=<id>&user_id=<id>` - Real-time collaboration

## Tech Stack

### Backend
- Go 1.21
- Gin (HTTP framework)
- Gorilla WebSocket
- JWT authentication
- LaTeX (pdflatex)

### Frontend
- React 18
- TypeScript
- Tailwind CSS
- Monaco Editor
- Redux Toolkit
- TanStack Query

### Infrastructure
- Docker & Docker Compose
- PostgreSQL
- Redis

## License

MIT License
