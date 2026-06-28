#!/bin/bash

echo "================================================"
echo "  Running Database Migrations"
echo "================================================"
echo ""

# Check if PostgreSQL is running
echo "Checking PostgreSQL availability..."
if ! docker-compose exec -T postgres pg_isready -U mailgo > /dev/null 2>&1; then
    echo "❌ PostgreSQL is not ready. Please start services first:"
    echo "   ./start.sh"
    exit 1
fi

echo "✅ PostgreSQL is ready"
echo ""

# Check if migrations have already been run
TABLES_COUNT=$(docker-compose exec -T postgres psql -U mailgo -d mailgo -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';" 2>/dev/null | tr -d ' ')

if [ "$TABLES_COUNT" -gt 0 ]; then
    echo "⚠️  Database already has $TABLES_COUNT tables"
    echo ""
    read -p "Do you want to re-run migrations? This will fail if tables exist. (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Skipping migrations."
        exit 0
    fi
fi

echo "Running migration: 000001_initial_schema.up.sql"
echo ""

# Run the migration
if docker-compose exec -T postgres psql -U mailgo -d mailgo < backend/migrations/000001_initial_schema.up.sql > /dev/null 2>&1; then
    echo "✅ Migration completed successfully"
    echo ""
    
    # Show created tables
    echo "Created tables:"
    docker-compose exec -T postgres psql -U mailgo -d mailgo -c "\dt" | grep "public"
    
else
    echo "❌ Migration failed"
    echo ""
    echo "This is normal if migrations have already been run."
    echo "To check current database state:"
    echo "  docker-compose exec postgres psql -U mailgo -d mailgo -c '\dt'"
    exit 1
fi

echo ""
echo "================================================"
echo "  Migration Complete"
echo "================================================"
