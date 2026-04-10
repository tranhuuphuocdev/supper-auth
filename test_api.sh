#!/bin/bash

# Auth Service - API Testing Script with curl

BASE_URL="http://localhost:8080"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🧪 Auth Service API Testing Script${NC}\n"

# Colors for output
print_section() {
    echo -e "\n${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}$1${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

# 1. Health Check
print_section "1. Health Check"
curl -X GET "$BASE_URL/health" \
  -H "Content-Type: application/json" \
  -w "\nStatus: %{http_code}\n\n"

# 2. Register User (Admin)
print_section "2. Register Admin User"
ADMIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123456",
    "first_name": "Admin",
    "last_name": "User"
  }')

echo "$ADMIN_RESPONSE" | jq '.' 2>/dev/null || echo "$ADMIN_RESPONSE"
ADMIN_ID=$(echo "$ADMIN_RESPONSE" | jq -r '.data.id' 2>/dev/null || echo "1")

# 3. Register User (Regular)
print_section "3. Register Regular User"
USER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "user123456",
    "first_name": "John",
    "last_name": "Doe"
  }')

echo "$USER_RESPONSE" | jq '.' 2>/dev/null || echo "$USER_RESPONSE"
USER_ID=$(echo "$USER_RESPONSE" | jq -r '.data.id' 2>/dev/null || echo "2")

# 4. Login Admin
print_section "4. Login as Admin"
ADMIN_LOGIN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123456"
  }')

echo "$ADMIN_LOGIN" | jq '.' 2>/dev/null || echo "$ADMIN_LOGIN"
ADMIN_TOKEN=$(echo "$ADMIN_LOGIN" | jq -r '.data.token' 2>/dev/null || echo "")

# 5. Login User
print_section "5. Login as Regular User"
USER_LOGIN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "user123456"
  }')

echo "$USER_LOGIN" | jq '.' 2>/dev/null || echo "$USER_LOGIN"
USER_TOKEN=$(echo "$USER_LOGIN" | jq -r '.data.token' 2>/dev/null || echo "")

# 6. Get Current User (Admin)
print_section "6. Get Current User (Admin)"
curl -s -X GET "$BASE_URL/api/v1/auth/me" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.' 2>/dev/null || echo "Check token"

# 7. Get Current User (Regular User)
print_section "7. Get Current User (Regular User)"
curl -s -X GET "$BASE_URL/api/v1/auth/me" \
  -H "Authorization: Bearer $USER_TOKEN" | jq '.' 2>/dev/null || echo "Check token"

# 8. Assign Admin Role to First User (if token is valid)
if [ ! -z "$ADMIN_TOKEN" ] && [ "$ADMIN_TOKEN" != "null" ]; then
    print_section "8. Assign Admin Role to User"
    curl -s -X POST "$BASE_URL/api/v1/admin/users/$USER_ID/roles" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "role_id": 1
      }' | jq '.' 2>/dev/null || echo "Failed to assign role"
fi

# 9. List All Users
print_section "9. List All Users"
curl -s -X GET "$BASE_URL/api/v1/admin/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.' 2>/dev/null || echo "Failed to list users"

# 10. Get Specific User
print_section "10. Get Specific User (Admin)"
curl -s -X GET "$BASE_URL/api/v1/admin/users/$ADMIN_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.' 2>/dev/null || echo "Failed to get user"

# 11. List All Roles
print_section "11. List All Roles"
curl -s -X GET "$BASE_URL/api/v1/admin/roles" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.' 2>/dev/null || echo "Failed to list roles"

# 12. List All Permissions
print_section "12. List All Permissions"
curl -s -X GET "$BASE_URL/api/v1/admin/permissions" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.' 2>/dev/null || echo "Failed to list permissions"

# 13. Change Password
print_section "13. Change Password"
curl -s -X POST "$BASE_URL/api/v1/auth/change-password" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "user123456",
    "new_password": "newuser123456"
  }' | jq '.' 2>/dev/null || echo "Failed to change password"

# 14. Update User
print_section "14. Update User Info"
curl -s -X PUT "$BASE_URL/api/v1/admin/users/$USER_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Jane",
    "last_name": "Smith",
    "is_active": true
  }' | jq '.' 2>/dev/null || echo "Failed to update user"

# 15. Create New Role
print_section "15. Create New Role"
ROLE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/admin/roles" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "editor",
    "description": "Editor with publishing permissions"
  }')

echo "$ROLE_RESPONSE" | jq '.' 2>/dev/null || echo "$ROLE_RESPONSE"

print_section "✅ Testing Complete"
echo -e "${GREEN}All tests completed. Check the output above for results.${NC}\n"
echo "Note: Some operations may fail if user doesn't have the required permissions."
echo "Make sure the admin user has been assigned the admin role."
