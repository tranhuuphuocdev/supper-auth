# Auth Service - Golang REST API

A comprehensive authentication service built with Go, featuring JWT-based authentication, PostgreSQL database integration, and role-based access control (RBAC).

## Features

✅ User Registration & Login
✅ JWT Token-based Authentication  
✅ PostgreSQL Database Integration
✅ Role-Based Access Control (RBAC)
✅ Permission Management
✅ Password Hashing (bcrypt)
✅ Protected REST API Endpoints
✅ User Management (Admin)
✅ Role & Permission Management

## Tech Stack

- **Framework**: Gorilla Mux (HTTP Router)
- **Database**: PostgreSQL
- **Authentication**: JWT (golang-jwt/jwt)
- **Password Hashing**: bcrypt
- **Language**: Go 1.21+

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- git

## Installation

### 1. Clone/Setup Project

```bash
cd ~/Workspace/golang
```

### 2. Create `.env` file

```bash
cp .env.example .env
```

Edit the `.env` file with your PostgreSQL credentials:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=auth_service
JWT_SECRET=your-very-secure-secret-key
PORT=8080
```

### 3. Create PostgreSQL Database

```bash
createdb auth_service
```

Or using psql:

```sql
CREATE DATABASE auth_service;
```

### 4. Install Dependencies

```bash
go mod download
go mod tidy
```

## Running the Application

```bash
go run .
```

The server will start on `http://localhost:8080`

## Project Structure

```
.
├── main.go              # Entry point & server setup
├── config.go            # Configuration management
├── database.go          # Database connection & migrations
├── models.go            # Data models & request/response types
├── auth.go              # Authentication business logic
├── jwt.go               # JWT token generation & validation
├── middleware.go        # HTTP middleware (auth, permissions)
├── handlers.go          # REST API handlers
├── utils.go             # Helper functions
├── go.mod               # Module dependencies
└── .env.example         # Environment variables template
```

## API Endpoints

### Public Endpoints

#### Register User
```
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}

Response:
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": 1,
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "is_active": true,
    "created_at": "2024-04-08T10:00:00Z",
    "updated_at": "2024-04-08T10:00:00Z"
  }
}
```

#### Login
```
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}

Response:
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": { ... },
    "expires_at": "2024-04-09T10:00:00Z"
  }
}
```

### Protected Endpoints (Requires Bearer Token)

#### Get Current User
```
GET /api/v1/auth/me
Authorization: Bearer <token>

Response:
{
  "success": true,
  "message": "User retrieved successfully",
  "data": {
    "user": { ... },
    "roles": ["user"],
    "permissions": ["user.read"]
  }
}
```

#### Change Password
```
POST /api/v1/auth/change-password
Authorization: Bearer <token>
Content-Type: application/json

{
  "old_password": "password123",
  "new_password": "newpassword456"
}
```

#### Logout
```
POST /api/v1/auth/logout
Authorization: Bearer <token>
```

### Admin Endpoints (Requires `user.read` Permission)

#### List All Users
```
GET /api/v1/admin/users
Authorization: Bearer <admin_token>

Response:
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 1,
      "email": "user@example.com",
      ...
    }
  ]
}
```

#### Get Specific User
```
GET /api/v1/admin/users/{id}
Authorization: Bearer <admin_token>

Response:
{
  "success": true,
  "message": "User retrieved successfully",
  "data": {
    "user": { ... },
    "roles": ["admin"]
  }
}
```

#### Update User
```
PUT /api/v1/admin/users/{id}
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "first_name": "Jane",
  "last_name": "Smith",
  "is_active": true
}
```

#### Delete User
```
DELETE /api/v1/admin/users/{id}
Authorization: Bearer <admin_token>
```

#### Assign Role to User
```
POST /api/v1/admin/users/{id}/roles
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "role_id": 2
}
```

#### Remove Role from User
```
DELETE /api/v1/admin/users/{id}/roles/{roleId}
Authorization: Bearer <admin_token>
```

### Role Management

#### List All Roles
```
GET /api/v1/admin/roles
Authorization: Bearer <admin_token>

Response:
{
  "success": true,
  "message": "Roles retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "admin",
      "description": "Administrator with full access",
      "created_at": "2024-04-08T10:00:00Z"
    },
    {
      "id": 2,
      "name": "moderator",
      "description": "Moderator can manage content",
      "created_at": "2024-04-08T10:00:00Z"
    },
    {
      "id": 3,
      "name": "user",
      "description": "Regular user with limited access",
      "created_at": "2024-04-08T10:00:00Z"
    }
  ]
}
```

#### Create New Role
```
POST /api/v1/admin/roles
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "editor",
  "description": "Editor role with publishing permissions"
}
```

### Permissions

#### List All Permissions
```
GET /api/v1/admin/permissions
Authorization: Bearer <admin_token>

Response:
{
  "success": true,
  "message": "Permissions retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "user.create",
      "description": "Can create new users",
      "created_at": "2024-04-08T10:00:00Z"
    },
    {
      "id": 2,
      "name": "user.read",
      "description": "Can read user data",
      "created_at": "2024-04-08T10:00:00Z"
    },
    ...
  ]
}
```

## Default Roles & Permissions

### Admin Role
- user.create
- user.read
- user.update
- user.delete
- role.manage
- permission.manage

### Moderator Role
- user.read
- user.update

### User Role
- user.read

## Authentication Flow

1. User registers with email and password
2. Password is hashed using bcrypt
3. User logs in with email and password
4. Server validates credentials and generates JWT token
5. Client includes token in `Authorization: Bearer <token>` header
6. Server validates token and extracts user permissions
7. Access is granted/denied based on permissions

## Using the API

### cURL Examples

#### Register:
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

#### Login:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

#### Get Current User:
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <your_token>"
```

#### List Users (requires user.read permission):
```bash
curl -X GET http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer <admin_token>"
```

## Security Considerations

1. **JWT Secret**: Change the `JWT_SECRET` in `.env` to a strong, random value in production
2. **HTTPS**: Always use HTTPS in production
3. **Password Requirements**: Implement minimum password length requirements (currently 6 chars)
4. **Token Expiration**: Tokens expire after 24 hours (configurable in `jwt.go`)
5. **Rate Limiting**: Implement rate limiting to prevent brute force attacks
6. **CORS**: Add CORS middleware if frontend is on different domain
7. **SQL Injection**: Using parameterized queries to prevent SQL injection
8. **Password Storage**: Using bcrypt with default cost for secure password hashing

## Database Schema

### Users Table
```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Roles Table
```sql
CREATE TABLE roles (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) UNIQUE NOT NULL,
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Permissions Table
```sql
CREATE TABLE permissions (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) UNIQUE NOT NULL,
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### User Roles (Many-to-Many)
```sql
CREATE TABLE user_roles (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  UNIQUE(user_id, role_id)
);
```

### Role Permissions (Many-to-Many)
```sql
CREATE TABLE role_permissions (
  id SERIAL PRIMARY KEY,
  role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  UNIQUE(role_id, permission_id)
);
```

## Extending the Service

### Add New Permission
1. Insert into `permissions` table
2. Assign to roles via `role_permissions` table
3. Use `PermissionMiddleware(permissionName)` in handlers

### Add New Role
1. Use `POST /api/v1/admin/roles` endpoint
2. Assign permissions through database
3. Assign to users via `AssignRoleToUser` function

### Add New Endpoint
1. Create handler function in `handlers.go`
2. Add route in `SetupRouter()` with appropriate middleware
3. Use `AuthMiddleware`, `PermissionMiddleware`, or `RoleMiddleware` as needed

## Troubleshooting

### Database Connection Error
- Check PostgreSQL is running
- Verify `.env` credentials
- Ensure database exists: `createdb auth_service`

### JWT Validation Error
- Ensure token is correctly formatted: `Bearer <token>`
- Check token hasn't expired
- Verify `JWT_SECRET` matches between encoding and decoding

### Permission Denied
- Check user has correct role assigned
- Verify role has permission assigned
- Use `/api/v1/auth/me` to view user's permissions

## Future Enhancements

- [ ] Email verification
- [ ] Password reset functionality
- [ ] Refresh tokens
- [ ] OAuth2 integration
- [ ] Rate limiting
- [ ] Audit logging
- [ ] Two-factor authentication (2FA)
- [ ] API key authentication
- [ ] Swagger/OpenAPI documentation

## License

MIT
