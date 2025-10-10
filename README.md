# AI Database Assistant Backend

A modular, scalable Go backend for an AI-powered database assistant platform that allows users to manage database projects, execute queries, and interact with databases through an AI chat interface.

## 🚀 Features

- **User Authentication**: JWT-based authentication with secure password hashing
- **Project Management**: Create, read, update, and delete database projects (PostgreSQL & MySQL)
- **Dashboard Analytics**: View project statistics, query history, and database metadata
- **Chat Interface**: Project-specific chat history (AI integration ready)
- **Secure**: Password hashing with bcrypt, JWT tokens, user authorization
- **Modular Architecture**: Clean separation of concerns with handlers, middleware, and utilities

## 📁 Project Structure

```
ai-db-assistant/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── database/
│   │   └── database.go          # Database connection & migrations
│   ├── handlers/
│   │   ├── auth.go              # Auth handlers (signup, login)
│   │   ├── project.go           # Project CRUD handlers
│   │   ├── chat.go              # Chat message handlers
│   │   └── dashboard.go         # Dashboard statistics handlers
│   ├── middleware/
│   │   └── middleware.go        # Auth, CORS, logging middleware
│   ├── models/
│   │   └── models.go            # Database models
│   └── router/
│       └── router.go            # Route definitions
├── pkg/
│   ├── jwt/
│   │   └── jwt.go               # JWT utilities
│   ├── logger/
│   │   └── logger.go            # Structured logging
│   ├── password/
│   │   └── password.go          # Password hashing utilities
│   └── response/
│       └── response.go          # HTTP response utilities
├── .env.example                 # Environment variables template
├── docker-compose.yml           # PostgreSQL Docker setup
├── Makefile                     # Build automation
└── go.mod                       # Go module dependencies
```

## 🛠️ Tech Stack

- **Language**: Go 1.21+
- **Framework**: `net/http` with Gorilla Mux
- **Database**: PostgreSQL
- **ORM**: GORM
- **Authentication**: JWT tokens
- **Password Hashing**: bcrypt
- **Logging**: Structured JSON logging with `log/slog`

## 📋 Prerequisites

- Go 1.21 or higher
- PostgreSQL 13+ (or use Docker)
- Make (optional, for using Makefile commands)

## 🚀 Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/ai-db-assistant.git
cd ai-db-assistant
```

### 2. Set up environment variables

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
PORT=8080
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/ai_db_assistant?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
ENVIRONMENT=development
```

### 3. Start PostgreSQL (using Docker)

```bash
make docker-up
# or
docker-compose up -d
```

### 4. Install dependencies

```bash
make deps
# or
go mod download
```

### 5. Run the application

```bash
make run
# or
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`

## 📡 API Endpoints

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/auth/signup` | Create new user account | No |
| POST | `/api/auth/login` | Login and get JWT token | No |

### Projects

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/projects` | Create a new project | Yes |
| GET | `/api/projects` | Get all user projects | Yes |
| GET | `/api/projects/:id` | Get specific project | Yes |
| PUT | `/api/projects/:id` | Update project | Yes |
| DELETE | `/api/projects/:id` | Delete project | Yes |

### Dashboard

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/projects/:id/summary` | Get project statistics | Yes |
| GET | `/api/dashboard` | Get user dashboard stats | Yes |

### Chat

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/chat/:project_id` | Send chat message | Yes |
| GET | `/api/chat/:project_id/history` | Get chat history | Yes |

### Health Check

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/health` | Health check endpoint | No |

## 📝 API Usage Examples

### Signup

```bash
curl -X POST http://localhost:8080/api/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

### Create Project

```bash
curl -X POST http://localhost:8080/api/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "My Database",
    "description": "Production database",
    "database_type": "postgresql",
    "connection_string": "postgresql://user:pass@localhost:5432/mydb"
  }'
```

### Get All Projects

```bash
curl -X GET http://localhost:8080/api/projects \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🏗️ Development

### Run tests

```bash
make test
```

### Build the application

```bash
make build
```

### Clean build artifacts

```bash
make clean
```

## 🔒 Security Features

- **Password Hashing**: Uses bcrypt for secure password storage
- **JWT Authentication**: Token-based authentication with configurable expiry
- **Authorization**: Middleware ensures users can only access their own resources
- **CORS**: Configurable CORS middleware for frontend integration
- **Input Validation**: Request validation to prevent invalid data

## 📊 Database Schema

### Users Table
- `id` (Primary Key)
- `name`
- `email` (Unique)
- `password` (Hashed)
- `created_at`
- `updated_at`

### Projects Table
- `id` (Primary Key)
- `user_id` (Foreign Key)
- `name`
- `description`
- `database_type` (mysql/postgresql)
- `connection_string`
- `created_at`
- `updated_at`

### Queries Table
- `id` (Primary Key)
- `project_id` (Foreign Key)
- `query`
- `status` (success/error)
- `result`
- `error`
- `created_at`

### Messages Table
- `id` (Primary Key)
- `project_id` (Foreign Key)
- `role` (user/assistant)
- `content`
- `created_at`

## 🔮 Future Enhancements

- [ ] OpenAI integration for AI chat responses
- [ ] Query execution engine for connected databases
- [ ] Real-time query monitoring
- [ ] Rate limiting
- [ ] API documentation with Swagger
- [ ] Comprehensive test coverage
- [ ] Docker support for full application
- [ ] CI/CD pipeline

## 📄 License

MIT License

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📧 Contact

For questions or support, please open an issue on GitHub.