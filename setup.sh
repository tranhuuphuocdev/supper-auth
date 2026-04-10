#!/bin/bash

# Auth Service - Quick Start Setup Script

echo "🚀 Starting Auth Service Setup..."
echo ""

# 1. Check PostgreSQL
echo "📦 Checking PostgreSQL..."
if ! command -v psql &> /dev/null; then
    echo "❌ PostgreSQL is not installed. Please install it first."
    echo "   macOS: brew install postgresql@15"
    echo "   Linux: sudo apt-get install postgresql"
    exit 1
fi

# 2. Create .env file
if [ ! -f .env ]; then
    echo "⚙️  Creating .env file..."
    cp .env.example .env
    echo "✅ .env file created. Please edit it with your PostgreSQL credentials and domain map."
else
    echo "✅ .env file already exists"
fi

# 3. Download dependencies
echo "📥 Downloading Go dependencies..."
go mod download
go mod tidy

# 4. Build the application
echo "🔨 Building application..."
go build -o auth-service

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Setup completed successfully!"
    echo ""
    echo "📖 Next steps:"
    echo "   1. Ensure PostgreSQL databases are already created"
    echo "   2. Edit .env (DB credentials + DOMAIN_DB_MAP)"
    echo "   3. Run: ./auth-service"
    echo "   4. Service will auto-create tables on first request per domain"
    echo "   5. Visit: http://localhost:8080/health"
    echo ""
else
    echo "❌ Build failed. Please check the errors above."
    exit 1
fi
