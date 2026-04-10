# 🚀 Getting Started - Auth Service

## Quick Start (5 minutes)

### Step 1: Setup Environment

```bash
cd ~/Workspace/golang

# Copy environment template
cp .env.example .env

# Edit .env with your PostgreSQL credentials
# (Optional: use default credentials if using docker-compose)
```

### Step 2: Start PostgreSQL

**Option A: Using Docker Compose (Recommended)**
```bash
docker-compose up -d
```

**Option B: Using Local PostgreSQL**
```bash
# Create database
createdb auth_service

# The service will auto-create tables on first run
```

### Step 3: Download Dependencies & Build

```bash
go mod download
go mod tidy
go build -o auth-service
```

### Step 4: Run the Server

```bash
./auth-service
# Server starts on http://localhost:8080
```

You should see:
```
Database connected successfully
Migrations completed successfully
Server starting on port 8080
```

## Testing the API

### Quick Test (after server is running)

**1. Register a user:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123456",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

**2. Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123456"
  }'
```

You'll get a response with a JWT token:
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": { ... },
    "expires_at": "2024-04-09T..."
  }
}
```

**3. Use the token to access protected endpoints:**
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## Advanced Testing

### Using the Test Script

```bash
# Make script executable
chmod +x test_api.sh

# Run comprehensive tests
./test_api.sh
```

### Using Postman

1. Import `postman_collection.json` into Postman
2. Set base URL: `http://localhost:8080`
3. Run requests from the collection

## Project Structure

```
golang/
├── Go Source Files (*.go)
│   ├── main.go              # Entry point
│   ├── config.go            # Configuration
│   ├── database.go          # Database setup
│   ├── models.go            # Data types
│   ├── auth.go              # Auth logic
│   ├── jwt.go               # JWT handling
│   ├── middleware.go        # HTTP middleware
│   ├── handlers.go          # API handlers
│   └── utils.go             # Helpers
│
├── Configuration
│   ├── .env.example         # Template
│   ├── .env                 # Your config (create from example)
│   ├── go.mod               # Dependencies
│   └── docker-compose.yml   # PostgreSQL setup
│
├── Documentation
│   ├── README.md            # Full documentation
│   ├── ARCHITECTURE.md      # Architecture details
│   └── GETTING_STARTED.md   # This file
│
└── Tools
    ├── setup.sh             # Automated setup
    ├── test_api.sh          # API testing
    └── postman_collection.json # Postman import
```

## API Endpoints Summary

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/auth/register` | ❌ | Register new user |
| POST | `/api/v1/auth/login` | ❌ | Login & get token |
| GET | `/api/v1/auth/me` | ✅ | Get current user |
| POST | `/api/v1/auth/change-password` | ✅ | Change password |
| GET | `/api/v1/admin/users` | ✅ | List all users |
| GET | `/api/v1/admin/users/{id}` | ✅ | Get user details |
| PUT | `/api/v1/admin/users/{id}` | ✅ | Update user |
| DELETE | `/api/v1/admin/users/{id}` | ✅ | Delete user |
| GET | `/api/v1/admin/roles` | ✅ | List roles |
| POST | `/api/v1/admin/roles` | ✅ | Create role |
| GET | `/api/v1/admin/permissions` | ✅ | List permissions |

## Default Credentials

After first run, the service:
- Creates tables automatically
- Adds default roles: `admin`, `moderator`, `user`
- Adds default permissions for CRUD operations

### Test User Setup

**Register an admin:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "Admin@12345",
    "first_name": "Admin",
    "last_name": "User"
  }'
```

Then assign admin role via database or API after they're logged in.

## Troubleshooting

### Server won't start
- ✅ Check PostgreSQL is running
- ✅ Verify `.env` has correct credentials
- ✅ Check port 8080 is not in use

### Database connection error
```bash
# Test PostgreSQL connection
psql -U postgres -h localhost -d auth_service

# If using docker
docker-compose ps  # Check if postgres is running
docker-compose logs postgres  # View logs
```

### JWT token errors
- ✅ Ensure token is in `Authorization: Bearer <token>` format
- ✅ Check token hasn't expired (24 hours)
- ✅ Verify token was correctly copied without extra spaces

### Permission denied errors
- ✅ Check user has correct role
- ✅ Verify role has the required permission
- ✅ Use `/api/v1/auth/me` to see user's permissions

## Common Operations

### Create a new user with admin role

1. Register user via API
2. Login and get user ID from `/api/v1/auth/me`
3. Assign admin role:
```bash
curl -X POST http://localhost:8080/api/v1/admin/users/{user_id}/roles \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role_id": 1}'
```

### Change user password
```bash
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "OldPass@123",
    "new_password": "NewPass@123"
  }'
```

### Create a custom role
```bash
curl -X POST http://localhost:8080/api/v1/admin/roles \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "editor",
    "description": "Can edit published content"
  }'
```

## Important Security Notes

1. **Change JWT Secret**: Update `JWT_SECRET` in `.env` to a strong, random value
2. **Use HTTPS**: Always use HTTPS in production
3. **Keep .env Secure**: Never commit `.env` to git
4. **Token Expiration**: Tokens expire after 24 hours
5. **Rate Limiting**: Consider adding rate limiting in production
6. **Password Policy**: Minimum 6 characters (customize as needed)

## Next Steps

1. ✅ Complete the Quick Start above
2. ✅ Test API endpoints with curl or Postman
3. ✅ Read `README.md` for full documentation
4. ✅ Read `ARCHITECTURE.md` for technical details
5. ✅ Customize roles, permissions, and features as needed

## Support & Documentation

- **Full API Reference**: See `README.md`
- **Architecture Details**: See `ARCHITECTURE.md`
- **Source Code**: Read individual `.go` files (well-commented)
- **Database Schema**: See `README.md` Database Schema section

## Production Deployment Checklist

- [ ] Change JWT_SECRET to strong random value
- [ ] Configure PostgreSQL credentials
- [ ] Enable HTTPS/TLS
- [ ] Set PORT environment variable appropriately
- [ ] Add rate limiting middleware
- [ ] Enable CORS if needed
- [ ] Add request validation
- [ ] Setup logging and monitoring
- [ ] Configure database backups
- [ ] Add email verification
- [ ] Implement password reset flow
- [ ] Add audit logging
- [ ] Review security best practices

---

Happy coding! 🎉
