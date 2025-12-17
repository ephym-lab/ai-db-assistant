# API Endpoints Documentation

This document provides a comprehensive list of all API endpoints, their request formats, and response structures.

## Base URL
```
http://localhost:8080
```

---

## Authentication Endpoints

### 1. User Signup
**Endpoint:** `POST /api/auth/signup`  
**Authentication:** Not required  
**Description:** Create a new user account

**Request Body:**
```json
{
  "name": "string",
  "email": "string",
  "password": "string"
}
```

**Success Response (201 Created):**
```json
{
  "success": true,
  "message": "User created successfully",
  "data": {
    "user": {
      "id": 1,
      "username": "John Doe",
      "email": "john@example.com",
      "created_at": "2025-12-17T12:00:00Z",
      "updated_at": "2025-12-17T12:00:00Z"
    }
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body or missing required fields
- `409 Conflict`: User with this email already exists

---

### 2. User Login
**Endpoint:** `POST /api/auth/login`  
**Authentication:** Not required  
**Description:** Login and receive JWT token

**Request Body:**
```json
{
  "email": "string",
  "password": "string"
}
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "John Doe",
      "email": "john@example.com",
      "created_at": "2025-12-17T12:00:00Z",
      "updated_at": "2025-12-17T12:00:00Z"
    }
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body or missing required fields
- `401 Unauthorized`: Invalid email or password

---

## Project Endpoints

### 3. Create Project
**Endpoint:** `POST /api/projects`  
**Authentication:** Required (Bearer token)  
**Description:** Create a new database project

**Request Body:**
```json
{
  "name": "string",
  "description": "string",
  "database_type": "postgresql | mysql",
  "connection_string": "string",
  "allow_ddl": true,
  "allow_write": true,
  "allow_read": true,
  "allow_delete": true
}
```

**Note:** Permission fields (`allow_ddl`, `allow_write`, `allow_read`, `allow_delete`) are optional and default to `true`.

**Success Response (201 Created):**
```json
{
  "success": true,
  "message": "Project created successfully",
  "data": {
    "id": 1,
    "user_id": 1,
    "name": "My Database",
    "description": "Production database",
    "database_type": "postgresql",
    "connection_string": "postgresql://user:pass@localhost:5432/mydb",
    "created_at": "2025-12-17T12:00:00Z",
    "updated_at": "2025-12-17T12:00:00Z",
    "permission": {
      "id": 1,
      "project_id": 1,
      "allow_ddl": true,
      "allow_write": true,
      "allow_read": true,
      "allow_delete": true,
      "created_at": "2025-12-17T12:00:00Z",
      "updated_at": "2025-12-17T12:00:00Z"
    }
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body or database type
- `401 Unauthorized`: Missing or invalid token

---

### 4. Get All Projects
**Endpoint:** `GET /api/projects`  
**Authentication:** Required (Bearer token)  
**Description:** Get all projects for the authenticated user

**Success Response (200 OK):**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "name": "My Database",
    "description": "Production database",
    "database_type": "postgresql",
    "connection_string": "postgresql://user:pass@localhost:5432/mydb",
    "created_at": "2025-12-17T12:00:00Z",
    "updated_at": "2025-12-17T12:00:00Z",
    "user": {
      "id": 1,
      "username": "John Doe",
      "email": "john@example.com",
      "created_at": "2025-12-17T12:00:00Z",
      "updated_at": "2025-12-17T12:00:00Z"
    },
    "permission": {
      "id": 1,
      "project_id": 1,
      "allow_ddl": true,
      "allow_write": true,
      "allow_read": true,
      "allow_delete": true,
      "created_at": "2025-12-17T12:00:00Z",
      "updated_at": "2025-12-17T12:00:00Z"
    }
  }
]
```

**Error Responses:**
- `401 Unauthorized`: Missing or invalid token

---

### 5. Get Project by ID
**Endpoint:** `GET /api/projects/{id}`  
**Authentication:** Required (Bearer token)  
**Description:** Get a specific project by ID

**Success Response (200 OK):**
```json
{
  "id": 1,
  "user_id": 1,
  "name": "My Database",
  "description": "Production database",
  "database_type": "postgresql",
  "connection_string": "postgresql://user:pass@localhost:5432/mydb",
  "created_at": "2025-12-17T12:00:00Z",
  "updated_at": "2025-12-17T12:00:00Z",
  "user": {
    "id": 1,
    "username": "John Doe",
    "email": "john@example.com",
    "created_at": "2025-12-17T12:00:00Z",
    "updated_at": "2025-12-17T12:00:00Z"
  },
  "permission": {
    "id": 1,
    "project_id": 1,
    "allow_ddl": true,
    "allow_write": true,
    "allow_read": true,
    "allow_delete": true,
    "created_at": "2025-12-17T12:00:00Z",
    "updated_at": "2025-12-17T12:00:00Z"
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid project ID
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: Project not found

---

### 6. Update Project
**Endpoint:** `PUT /api/projects/{id}`  
**Authentication:** Required (Bearer token)  
**Description:** Update project details and permissions

**Request Body:**
```json
{
  "name": "string",
  "description": "string",
  "connection_string": "string",
  "allow_ddl": false,
  "allow_write": false,
  "allow_read": true,
  "allow_delete": false
}
```

**Note:** All fields are optional. Only provided fields will be updated.

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "Project updated successfully",
  "data": {
    "id": 1,
    "user_id": 1,
    "name": "Updated Database",
    "description": "Updated description",
    "database_type": "postgresql",
    "connection_string": "postgresql://user:pass@localhost:5432/mydb",
    "created_at": "2025-12-17T12:00:00Z",
    "updated_at": "2025-12-17T12:30:00Z",
    "permission": {
      "id": 1,
      "project_id": 1,
      "allow_ddl": false,
      "allow_write": false,
      "allow_read": true,
      "allow_delete": false,
      "created_at": "2025-12-17T12:00:00Z",
      "updated_at": "2025-12-17T12:30:00Z"
    }
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid project ID or request body
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: Project not found

---

### 7. Delete Project
**Endpoint:** `DELETE /api/projects/{id}`  
**Authentication:** Required (Bearer token)  
**Description:** Delete a project

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "Project deleted successfully",
  "data": null
}
```

**Error Responses:**
- `400 Bad Request`: Invalid project ID
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: Project not found

---

### 8. Get Project Permissions
**Endpoint:** `GET /api/projects/{id}/permissions`  
**Authentication:** Required (Bearer token)  
**Description:** Get permissions for a specific project

**Success Response (200 OK):**
```json
{
  "id": 1,
  "project_id": 1,
  "allow_ddl": true,
  "allow_write": true,
  "allow_read": true,
  "allow_delete": true,
  "created_at": "2025-12-17T12:00:00Z",
  "updated_at": "2025-12-17T12:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid project ID
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: Project not found or permissions not found

---

## Dashboard Endpoints

### 9. Get User Dashboard
**Endpoint:** `GET /api/dashboard`  
**Authentication:** Required (Bearer token)  
**Description:** Get dashboard statistics for the authenticated user

**Success Response (200 OK):**
```json
{
  "total_projects": 5,
  "total_queries": 127,
  "total_messages": 342
}
```

**Error Responses:**
- `401 Unauthorized`: Missing or invalid token

---

### 10. Get Project Summary
**Endpoint:** `GET /api/projects/{id}/summary`  
**Authentication:** Required (Bearer token)  
**Description:** Get detailed statistics for a specific project

**Success Response (200 OK):**
```json
{
  "project_id": 1,
  "project_name": "My Database",
  "database_type": "postgresql",
  "table_count": 15,
  "total_queries": 45,
  "recent_queries": [
    {
      "id": 1,
      "project_id": 1,
      "query": "SELECT * FROM users",
      "query_type": "SELECT",
      "status": "success",
      "result": "{\"rows\": [...]}",
      "rows_affected": 10,
      "execution_time": 25,
      "created_at": "2025-12-17T12:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request`: Invalid project ID
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: Project not found

---

## Chat Endpoints

### 11. Send Chat Message
**Endpoint:** `POST /api/chat/{project_id}`  
**Authentication:** Required (Bearer token)  
**Description:** Send a chat message and receive AI-generated SQL

**Request Body:**
```json
{
  "content": "string"
}
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "Message sent successfully",
  "data": {
    "user_message": {
      "id": 1,
      "project_id": 1,
      "role": "user",
      "content": "Show me all users",
      "created_at": "2025-12-17T12:00:00Z"
    },
    "ai_message": {
      "id": 2,
      "project_id": 1,
      "role": "assistant",
      "content": "{\"content\":\"Here's the SQL query...\",\"query\":\"SELECT * FROM users\"}",
      "created_at": "2025-12-17T12:00:01Z"
    },
    "ai_response": {
      "content": "Here's the SQL query to show all users:",
      "query": "SELECT * FROM users"
    }
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid project ID or missing content
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: Project not found

---

### 12. Get Chat History
**Endpoint:** `GET /api/chat/{project_id}/history`  
**Authentication:** Required (Bearer token)  
**Description:** Get chat history for a specific project

**Success Response (200 OK):**
```json
[
  {
    "id": 1,
    "project_id": 1,
    "role": "user",
    "content": "Show me all users",
    "created_at": "2025-12-17T12:00:00Z"
  },
  {
    "id": 2,
    "project_id": 1,
    "role": "assistant",
    "ai_response": {
      "content": "Here's the SQL query to show all users:",
      "query": "SELECT * FROM users"
    },
    "created_at": "2025-12-17T12:00:01Z"
  }
]
```

**Error Responses:**
- `400 Bad Request`: Invalid project ID
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: Project not found

---

## Database Operation Endpoints

### 13. Connect to Database
**Endpoint:** `POST /api/projects/{id}/connect-db`  
**Authentication:** Required (Bearer token)  
**Description:** Establish a connection to the project's database via proxy

**Success Response (200 OK):**
```json
{
  "session_id": "uuid-string",
  "message": "Connected to database successfully",
  "database_type": "postgresql"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid project ID
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: Project not found
- `500 Internal Server Error`: Failed to connect to database

---

### 14. Disconnect from Database
**Endpoint:** `POST /api/projects/{id}/disconnect-db`  
**Authentication:** Required (Bearer token)  
**Description:** Close the database connection

**Success Response (200 OK):**
```json
{
  "message": "Disconnected from database successfully"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid project ID
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: Project not found
- `500 Internal Server Error`: Failed to disconnect from database

---

### 15. Execute SQL Query
**Endpoint:** `POST /api/projects/{id}/execute-sql`  
**Authentication:** Required (Bearer token)  
**Description:** Execute a SQL query with permission checks

**Request Body:**
```json
{
  "query": "string",
  "dry_run": false
}
```

**Note:** `dry_run` is optional and defaults to `false`.

**Success Response (200 OK) - SELECT Query:**
```json
{
  "columns": ["id", "name", "email"],
  "rows": [
    [1, "John Doe", "john@example.com"],
    [2, "Jane Smith", "jane@example.com"]
  ],
  "row_count": 2,
  "message": "Query executed successfully"
}
```

**Success Response (200 OK) - INSERT/UPDATE/DELETE Query:**
```json
{
  "affected_rows": 1,
  "message": "Query executed successfully"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid project ID or missing query
- `401 Unauthorized`: Missing or invalid token
- `403 Forbidden`: Operation not allowed based on project permissions
- `404 Not Found`: Project not found
- `500 Internal Server Error`: Failed to execute query

**Permission Requirements:**
- DDL operations (CREATE, ALTER, DROP): `allow_ddl` must be `true`
- Write operations (INSERT, UPDATE): `allow_write` must be `true`
- Read operations (SELECT): `allow_read` must be `true`
- Delete operations (DELETE, TRUNCATE): `allow_delete` must be `true`

---

### 16. Validate SQL Query
**Endpoint:** `POST /api/projects/{id}/validate-sql`  
**Authentication:** Required (Bearer token)  
**Description:** Validate a SQL query without executing it

**Request Body:**
```json
{
  "query": "string"
}
```

**Success Response (200 OK):**
```json
{
  "valid": true,
  "message": "Query is valid",
  "query_type": "SELECT",
  "estimated_cost": "0.00..10.00"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid project ID or missing query
- `401 Unauthorized`: Missing or invalid token
- `403 Forbidden`: Read permission required for validation
- `404 Not Found`: Project not found
- `500 Internal Server Error`: Failed to validate query

---

### 17. Get Database Info
**Endpoint:** `GET /api/projects/{id}/db-info`  
**Authentication:** Required (Bearer token)  
**Description:** Get database connection information and metadata

**Success Response (200 OK):**
```json
{
  "connected": true,
  "database_type": "postgresql",
  "database_name": "mydb",
  "tables": [
    {
      "name": "users",
      "columns": [
        {
          "name": "id",
          "type": "integer",
          "nullable": false
        },
        {
          "name": "email",
          "type": "varchar",
          "nullable": false
        }
      ]
    }
  ]
}
```

**Error Responses:**
- `400 Bad Request`: Invalid project ID
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: Project not found
- `500 Internal Server Error`: Failed to get database info

---

## Health Check Endpoint

### 18. Health Check
**Endpoint:** `GET /health`  
**Authentication:** Not required  
**Description:** Check if the API is running

**Success Response (200 OK):**
```
OK
```

---

## Authentication

All protected endpoints require a JWT token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

The token is obtained from the login endpoint and should be included in all subsequent requests to protected endpoints.

---

## Error Response Format

All error responses follow this format:

```json
{
  "success": false,
  "message": "Error description"
}
```

Common HTTP status codes:
- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource already exists
- `500 Internal Server Error`: Server error

---

## Notes

1. All timestamps are in ISO 8601 format (UTC)
2. The `connection_string` field contains sensitive information and should be handled securely
3. Database operations require an active connection established via the `/connect-db` endpoint
4. Query execution is logged in the `queries` table for audit purposes
5. Chat messages are stored with AI responses in JSON format for structured data
