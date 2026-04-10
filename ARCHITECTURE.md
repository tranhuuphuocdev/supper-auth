# Auth Service Project Structure & Documentation

## 📁 Directory Structure

```
golang/
├── main.go                     # Entry point and server initialization
├── config.go                   # Configuration management (env vars)
├── database.go                 # Database connection and migrations
├── models.go                   # Data structures and types
├── auth.go                     # Authentication business logic
├── jwt.go                      # JWT token generation and validation
├── middleware.go               # HTTP middleware (auth, permissions)
├── handlers.go                 # REST API endpoint handlers
├── utils.go                    # Helper functions (responses)
├── go.mod                      # Go module file
├── go.sum                      # Go module checksums
├── auth-service               # Compiled binary (generated)
│
├── .env.example                # Environment variables template
├── .env                        # Environment variables (create from example)
│
├── README.md                   # Main documentation
├── ARCHITECTURE.md             # This file
├── setup.sh                    # Automated setup script
├── test_api.sh                 # API testing script
├── postman_collection.json     # Postman API collection
│
├── docker-compose.yml          # PostgreSQL Docker setup
└── Dockerfile                  # Optional: Docker container for the app

```

## 🏗️ Architecture Overview

### Layered Architecture

```
┌─────────────────────────────────┐
│     HTTP Requests               │
└──────────┬──────────────────────┘
           │
┌──────────▼──────────────────────┐
│     Middleware Layer            │
│  (Auth, Permissions, CORS)      │
└──────────┬──────────────────────┘
           │
┌──────────▼──────────────────────┐
│     Handlers Layer              │
│  (API Endpoint Logic)           │
└──────────┬──────────────────────┘
           │
┌──────────▼──────────────────────┐
│     Business Logic Layer        │
│  (Auth, User Mgmt, Roles)      │
└──────────┬──────────────────────┘
           │
┌──────────▼──────────────────────┐
│     Database Layer              │
│  (PostgreSQL + Models)          │
└─────────────────────────────────┘
```

## 📄 File Descriptions

### **main.go**
- Server entry point
- HTTP server initialization
- Router setup
- Configuration loading
- Database initialization
- Graceful shutdown

### **config.go**
- Loads environment variables from `.env`
- Validates required configuration
- Centralizes all config values
- Provides getEnv helper function

### **database.go**
- PostgreSQL connection management
- Connection pooling setup
- Database migration execution
- Table creation (users, roles, permissions, mappings)
- Index creation for performance
- Seed default roles and permissions

#### Database Schema:
- **users**: User accounts with hashed passwords
- **roles**: User roles (admin, moderator, user)
- **permissions**: System permissions (user.*, role.*, etc.)
- **user_roles**: Many-to-many mapping
- **role_permissions**: Many-to-many mapping

### **models.go**
- User struct
- Role struct
- Permission struct
- Request/Response DTOs
- JWT Claims structure

### **auth.go**
- User registration
- User authentication (login)
- Password validation (bcrypt)
- Password change
- User retrieval
- Role assignment/removal
- Permission checking
- Role and permission fetching

### **jwt.go**
- JWT token generation (HS256)
- Token validation
- Claims extraction
- Token expiration handling (24 hours default)

### **middleware.go**
- AuthMiddleware: Validates JWT tokens
- PermissionMiddleware: Checks user permissions
- RoleMiddleware: Checks user roles
- Header extraction and validation

### **handlers.go**
- Health check
- User registration
- User login
- Password change
- Current user info
- User management (CRUD)
- Role management
- Permission listing
- Router setup (SetupRouter)

### **utils.go**
- SendSuccess: Success response formatter
- SendError: Error response formatter
- SendErrorWithDetails: Error with details

## 🔐 Security Features

1. **Password Hashing**: bcrypt with default cost (10)
2. **JWT Authentication**: HS256 signed tokens
3. **Token Expiration**: 24-hour expiration
4. **SQL Injection Prevention**: Parameterized queries
5. **Permission-Based Access Control**: Fine-grained permissions
6. **Role-Based Access Control**: Hierarchical roles
7. **Active User Status**: Deactivate users without deletion
8. **Secure Header Handling**: Bearer token extraction

## 🔄 Authentication Flow

```
1. User Registration
   ┌─────────────────┐
   │ Email, Password │
   └────────┬────────┘
            │
            ▼
   ┌─────────────────────────┐
   │ Validate & Hash Password│
   │ (bcrypt)                │
   └────────┬────────────────┘
            │
            ▼
   ┌─────────────────────────┐
   │ Store in Database       │
   │ Assign "user" role      │
   └─────────────────────────┘

2. User Login
   ┌──────────────────┐
   │ Email, Password  │
   └────────┬─────────┘
            │
            ▼
   ┌──────────────────────────┐
   │ Fetch User from Database │
   └────────┬─────────────────┘
            │
            ▼
   ┌──────────────────────────────┐
   │ Verify Password (bcrypt)     │
   └────────┬─────────────────────┘
            │
            ▼
   ┌──────────────────────────────┐
   │ Get User Roles & Permissions │
   └────────┬─────────────────────┘
            │
            ▼
   ┌──────────────────────────────┐
   │ Generate JWT Token           │
   │ Include roles, permissions   │
   └──────────┬───────────────────┘
              │
              ▼
   ┌──────────────────────────────┐
   │ Return Token + User Info     │
   └──────────────────────────────┘

3. API Request with Token
   ┌──────────────────────────┐
   │ GET /api/v1/auth/me      │
   │ Authorization: Bearer... │
   └────────┬─────────────────┘
            │
            ▼
   ┌──────────────────────────┐
   │ Extract Token from Header│
   └────────┬─────────────────┘
            │
            ▼
   ┌──────────────────────────┐
   │ Validate Token Signature │
   │ Check Expiration         │
   └────────┬─────────────────┘
            │
            ▼
   ┌──────────────────────────┐
   │ Extract Claims (UserID,  │
   │ Roles, Permissions)      │
   └────────┬─────────────────┘
            │
            ▼
   ┌──────────────────────────┐
   │ Process Request          │
   │ Return Protected Data    │
   └──────────────────────────┘
```

## 👥 Role Hierarchy

```
┌─────────────────────┐
│      Admin          │
├─────────────────────┤
│ • user.create       │
│ • user.read         │
│ • user.update       │
│ • user.delete       │
│ • role.manage       │
│ • permission.manage │
└─────────────────────┘

┌─────────────────────┐
│    Moderator        │
├─────────────────────┤
│ • user.read         │
│ • user.update       │
└─────────────────────┘

┌─────────────────────┐
│       User          │
├─────────────────────┤
│ • user.read         │
└─────────────────────┘
```

## 📊 Database Schema

```sql
-- Users Table
┌─────────────────────────────┐
│ users                       │
├─────────────────────────────┤
│ id (PK)                     │
│ email (UNIQUE)              │
│ password (hashed)           │
│ first_name                  │
│ last_name                   │
│ is_active                   │
│ created_at                  │
│ updated_at                  │
└─────────────────────────────┘

-- Roles Table
┌─────────────────────────────┐
│ roles                       │
├─────────────────────────────┤
│ id (PK)                     │
│ name (UNIQUE)               │
│ description                 │
│ created_at                  │
└─────────────────────────────┘

-- Permissions Table
┌─────────────────────────────┐
│ permissions                 │
├─────────────────────────────┤
│ id (PK)                     │
│ name (UNIQUE)               │
│ description                 │
│ created_at                  │
└─────────────────────────────┘

-- User Roles (Many-to-Many)
┌─────────────────────────────┐
│ user_roles                  │
├─────────────────────────────┤
│ id (PK)                     │
│ user_id (FK)                │
│ role_id (FK)                │
│ UNIQUE(user_id, role_id)    │
└─────────────────────────────┘

-- Role Permissions (Many-to-Many)
┌─────────────────────────────┐
│ role_permissions            │
├─────────────────────────────┤
│ id (PK)                     │
│ role_id (FK)                │
│ permission_id (FK)          │
│ UNIQUE(role_id, perm_id)    │
└─────────────────────────────┘
```

## 🚀 API Endpoint Groups

### Public Endpoints
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login

### Protected Endpoints (Auth Required)
- `GET /api/v1/auth/me` - Get current user
- `POST /api/v1/auth/change-password` - Change password
- `POST /api/v1/auth/logout` - Logout

### Admin Endpoints (user.read permission required)
- `GET /api/v1/admin/users` - List users
- `GET /api/v1/admin/users/{id}` - Get user
- `PUT /api/v1/admin/users/{id}` - Update user
- `DELETE /api/v1/admin/users/{id}` - Delete user

### Role Management (role.manage permission required)
- `GET /api/v1/admin/roles` - List roles
- `POST /api/v1/admin/roles` - Create role
- `POST /api/v1/admin/users/{id}/roles` - Assign role
- `DELETE /api/v1/admin/users/{id}/roles/{roleId}` - Remove role

### Permission Management (permission.manage permission required)
- `GET /api/v1/admin/permissions` - List permissions

## 🔧 Environment Variables

```env
# Database
DB_HOST=localhost              # PostgreSQL hostname
DB_PORT=5432                   # PostgreSQL port
DB_USER=postgres               # Database user
DB_PASSWORD=password           # Database password
DB_NAME=auth_service           # Database name

# JWT
JWT_SECRET=secret-key-here    # JWT signing secret (min 32 chars recommended)

# Server
PORT=8080                      # Server port
```

## 💾 Data Flow

```
Request → Middleware → Handler → Business Logic → Database
                                ↓
                            Validation
                                ↓
                        Hashing/Encryption
                                ↓
                           Database Ops
                                ↓
Response ← Formatter ← Business Logic ← Database
```

## 🧪 Testing Strategy

1. **Unit Tests**: Test individual functions
2. **Integration Tests**: Test database operations
3. **API Tests**: Use provided `test_api.sh` script
4. **Postman Collection**: Import `postman_collection.json`
5. **Manual Testing**: Use curl commands in `README.md`

## 🐛 Error Handling

- HTTP 400: Bad Request (validation errors)
- HTTP 401: Unauthorized (invalid token or credentials)
- HTTP 403: Forbidden (insufficient permissions)
- HTTP 404: Not Found (resource not found)
- HTTP 409: Conflict (duplicate email, etc.)
- HTTP 500: Internal Server Error

## 📈 Performance Considerations

1. **Database Indices**: Created on frequently queried columns
2. **Connection Pooling**: Default Go sql.DB pooling
3. **Query Optimization**: Specific columns selected, not SELECT *
4. **Prepared Statements**: Using parameterized queries
5. **Token Caching**: Claims cached in request headers

## 🔜 Future Enhancements

- [ ] Email verification on signup
- [ ] Password reset functionality
- [ ] Refresh tokens
- [ ] OAuth2/OIDC support
- [ ] API key authentication
- [ ] Rate limiting
- [ ] Audit logging
- [ ] Two-factor authentication (2FA)
- [ ] CORS middleware
- [ ] Request validation middleware
- [ ] Swagger/OpenAPI docs
- [ ] Database transaction support
- [ ] Soft delete for users
- [ ] User activation/deactivation
- [ ] Permission inheritance

## 📚 Technology Stack

- **Language**: Go 1.21+
- **HTTP Router**: Gorilla Mux
- **Database**: PostgreSQL
- **Authentication**: JWT (HS256)
- **Password Hashing**: bcrypt
- **JSON Encoding**: Go standard library
- **Database Driver**: lib/pq

## 🤝 Contributing

This is a template service. Feel free to:
- Add new roles and permissions
- Extend authentication methods
- Add email verification
- Implement refresh tokens
- Add audit logging
- Create middleware plugins
