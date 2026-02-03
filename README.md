# go-pkg-database

A comprehensive Go package providing database utilities for MongoDB and
PostgreSQL with GORM, featuring connection management, transactions, health
checks, and enhanced data types.

## 🚀 Features

- **🗄️ MongoDB**: Connection management, transactions, health checks
- **🐘 PostgreSQL**: GORM integration with connection pooling
- **🆔 Base Models**: UUID-based models with timestamps and soft delete
- **📋 JSONB Support**: PostgreSQL JSONB column type with helper methods
- **🔄 Transaction Management**: Safe transaction handling with rollback and
  savepoints
- **❤️ Health Checks**: Built-in connection health monitoring
- **⚡ Connection Pooling**: Optimized connection management
- **🔒 Type Safety**: Strongly-typed database operations

## 📦 Installation

```bash
go get github.com/kerimovok/go-pkg-database
```

## 🏗️ Package Structure

```
go-pkg-database/
├── mongo/             # MongoDB client and utilities
└── sql/              # PostgreSQL/GORM utilities and base models
    ├── base.go       # Base models and UUID types
    ├── gorm.go       # GORM configuration and connection
    ├── jsonb.go      # JSONB data type implementation
    └── transaction.go # Transaction management utilities
```

## 📖 Quick Start

### MongoDB Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/kerimovok/go-pkg-database/mongo"
)

func main() {
    config := mongo.MongoConfig{
        URI:             "mongodb://localhost:27017",
        DBName:          "myapp",
        Timeout:         10 * time.Second,
        MaxPoolSize:     100,
        MinPoolSize:     5,
        MaxIdleTime:     5 * time.Minute,
        ReadPreference:  "primary",
        RetryWrites:     true,
        RetryReads:      true,
    }

    client, err := mongo.Connect(config)
    if err != nil {
        log.Fatal("Failed to connect to MongoDB:", err)
    }
    defer client.Disconnect(context.Background())

    // Health check
    if client.IsHealthy(context.Background()) {
        log.Println("MongoDB connection is healthy")
    }

    // Get database and collection
    db := client.Database()
    collection := client.Collection("users")

    // Your MongoDB operations...
}
```

### PostgreSQL with GORM

```go
package main

import (
    "log"
    "time"

    "github.com/kerimovok/go-pkg-database/sql"
    "gorm.io/gorm/logger"
)

type User struct {
    sql.BaseModel
    Name  string     `gorm:"not null" json:"name"`
    Email string     `gorm:"unique;not null" json:"email"`
    Profile sql.JSONB `gorm:"type:jsonb" json:"profile"`
}

func main() {
    config := sql.GormConfig{
        Host:                      "localhost",
        User:                      "postgres",
        Password:                  "password",
        Name:                      "myapp",
        Port:                      "5432",
        SSLMode:                   "disable",
        Timezone:                  "UTC",
        MaxIdleConns:              10,
        MaxOpenConns:              100,
        ConnMaxLifetime:           30 * time.Minute,
        ConnMaxIdleTime:           10 * time.Minute,
        TranslateErrors:           true,
        LogLevel:                  logger.Info,
        SlowThreshold:             200 * time.Millisecond,
        IgnoreRecordNotFoundError: false,
    }

    db, err := sql.OpenGorm(config, &User{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

    // Health check
    ctx := context.Background()
    if err := db.Ping(ctx); err != nil {
        log.Printf("Database health check failed: %v", err)
    } else {
        log.Println("Database connection is healthy")
    }

    // Your GORM operations...
}
```

## 🗄️ MongoDB Features

### Connection Management

```go
import "github.com/kerimovok/go-pkg-database/mongo"

// Comprehensive configuration
config := mongo.MongoConfig{
    URI:            "mongodb://user:pass@localhost:27017/mydb",
    DBName:         "myapp",
    Timeout:        10 * time.Second,
    MaxPoolSize:    100,        // Maximum connections in pool
    MinPoolSize:    5,          // Minimum connections to maintain
    MaxIdleTime:    5 * time.Minute, // Connection idle timeout
    MaxConnecting:  10,         // Max concurrent connection attempts
    ReadPreference: "primaryPreferred", // Read preference
    RetryWrites:    true,       // Enable write retries
    RetryReads:     true,       // Enable read retries
}

client, err := mongo.Connect(config)
if err != nil {
    log.Fatal(err)
}
```

### MongoDB Operations

```go
import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
)

// Get collection
users := client.Collection("users")

// Insert document
user := bson.M{
    "name":  "John Doe",
    "email": "john@example.com",
    "age":   30,
}

result, err := users.InsertOne(context.Background(), user)
if err != nil {
    log.Fatal(err)
}
log.Printf("Inserted document with ID: %v", result.InsertedID)

// Find documents
var foundUsers []bson.M
cursor, err := users.Find(context.Background(), bson.M{"age": bson.M{"$gte": 18}})
if err != nil {
    log.Fatal(err)
}
defer cursor.Close(context.Background())

if err = cursor.All(context.Background(), &foundUsers); err != nil {
    log.Fatal(err)
}

// Update document
filter := bson.M{"email": "john@example.com"}
update := bson.M{"$set": bson.M{"age": 31}}
_, err = users.UpdateOne(context.Background(), filter, update)
if err != nil {
    log.Fatal(err)
}
```

### MongoDB Transactions

```go
// Execute operations within a transaction
err := client.WithTransaction(context.Background(), func(sc mongo.SessionContext) error {
    // All operations within this function will be part of the transaction

    // Insert user
    _, err := users.InsertOne(sc, bson.M{
        "name":  "Jane Doe",
        "email": "jane@example.com",
    })
    if err != nil {
        return err // Will trigger rollback
    }

    // Insert user profile
    _, err = profiles.InsertOne(sc, bson.M{
        "user_email": "jane@example.com",
        "bio":        "Software Developer",
    })
    if err != nil {
        return err // Will trigger rollback
    }

    return nil // Will commit transaction
})

if err != nil {
    log.Printf("Transaction failed: %v", err)
}
```

## 🐘 PostgreSQL/GORM Features

### Base Models

```go
import "github.com/kerimovok/go-pkg-database/sql"

// Using BaseModel (includes ID, CreatedAt, UpdatedAt, DeletedAt)
type User struct {
    sql.BaseModel
    Name  string `gorm:"not null" json:"name"`
    Email string `gorm:"unique;not null" json:"email"`
}

// Using individual components
type Product struct {
    sql.ID         // UUID ID only
    sql.Timestamp  // CreatedAt, UpdatedAt only
    Name  string   `gorm:"not null" json:"name"`
    Price float64  `json:"price"`
}

// Using soft delete functionality
type Category struct {
    sql.ID
    sql.Timestamp
    sql.SoftDelete // DeletedAt field with helper methods
    Name string `gorm:"not null" json:"name"`
}

// Check if soft deleted
if category.IsDeleted() {
    log.Println("Category is soft deleted")
}

// Soft delete a record
category.Delete()

// Restore a soft deleted record
category.Restore()
```

### JSONB Support

```go
type UserProfile struct {
    sql.BaseModel
    UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
    Data   sql.JSONB `gorm:"type:jsonb" json:"data"`
}

// Create user profile with JSONB data
profile := UserProfile{
    UserID: userID,
    Data: sql.JSONB{
        "preferences": map[string]interface{}{
            "theme":      "dark",
            "language":   "en",
            "timezone":   "UTC",
        },
        "settings": map[string]interface{}{
            "notifications": true,
            "email_updates": false,
        },
    },
}

db.Create(&profile)

// Working with JSONB data
profile.Data.Set("preferences.theme", "light")
profile.Data.Set("new_field", "value")

// Get values with type safety
theme, exists := profile.Data.GetString("preferences.theme")
if exists {
    log.Printf("User theme: %s", theme)
}

notifications, exists := profile.Data.GetBool("settings.notifications")
if exists {
    log.Printf("Notifications enabled: %t", notifications)
}

// Check if key exists
if profile.Data.Has("preferences.timezone") {
    timezone, _ := profile.Data.GetString("preferences.timezone")
    log.Printf("User timezone: %s", timezone)
}

// Get all keys
keys := profile.Data.Keys()
log.Printf("Available keys: %v", keys)

// Delete a key
profile.Data.Delete("old_field")

// Clone JSONB data
cloned := profile.Data.Clone()
```

### JSONB Array Support

```go
type UserActivity struct {
    sql.BaseModel
    UserID uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
    Events sql.JSONBArray `gorm:"type:jsonb" json:"events"`
}

// Create with array data
activity := UserActivity{
    UserID: userID,
    Events: sql.JSONBArray{
        map[string]interface{}{
            "type":      "login",
            "timestamp": time.Now(),
            "ip":        "192.168.1.1",
        },
        map[string]interface{}{
            "type":      "page_view",
            "timestamp": time.Now(),
            "page":      "/dashboard",
        },
    },
}

db.Create(&activity)
```

### Transactions

```go
import "github.com/kerimovok/go-pkg-database/sql"

// Simple transaction
err := sql.WithTransaction(db.DB, func(tx *gorm.DB) error {
    // Create user
    user := User{Name: "John", Email: "john@example.com"}
    if err := tx.Create(&user).Error; err != nil {
        return err // Will trigger rollback
    }

    // Create user profile
    profile := UserProfile{
        UserID: user.ID,
        Data: sql.JSONB{"bio": "Software Developer"},
    }
    if err := tx.Create(&profile).Error; err != nil {
        return err // Will trigger rollback
    }

    return nil // Will commit transaction
})

if err != nil {
    log.Printf("Transaction failed: %v", err)
}

// Transaction with context
ctx := context.Background()
err = sql.WithTransactionContext(ctx, db.DB, func(tx *gorm.DB) error {
    // Your transactional operations with context support
    return nil
})

// Nested transactions (using savepoints)
err = sql.WithNestedTransaction(db.DB, func(tx *gorm.DB) error {
    // This can be called within another transaction
    // Uses savepoints for nested transaction support
    return nil
})
```

### Advanced Features

```go
// Database statistics
stats := db.Stats()
log.Printf("Open connections: %d", stats.OpenConnections)
log.Printf("In use: %d", stats.InUse)
log.Printf("Idle: %d", stats.Idle)

// Connection health check with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := db.Ping(ctx); err != nil {
    log.Printf("Database ping failed: %v", err)
} else {
    log.Println("Database is responsive")
}

// Nullable UUID support
type Order struct {
    sql.BaseModel
    UserID    uuid.UUID           `gorm:"type:uuid;not null" json:"user_id"`
    AssigneeID sql.NullableUUID   `gorm:"type:uuid" json:"assignee_id"`
}

order := Order{UserID: userID}

// Set nullable UUID
assigneeUUID := uuid.New()
order.AssigneeID = sql.NullableUUID{UUID: assigneeUUID, Valid: true}

// Check if nullable UUID is set
if order.AssigneeID.Valid {
    log.Printf("Order assigned to: %s", order.AssigneeID.UUID)
}
```

## 🔧 Configuration

### MongoDB Configuration

```go
config := mongo.MongoConfig{
    URI:            "mongodb://localhost:27017", // Connection URI
    DBName:         "myapp",                     // Database name
    Timeout:        10 * time.Second,           // Operation timeout
    MaxPoolSize:    100,                        // Max connections
    MinPoolSize:    5,                          // Min connections
    MaxIdleTime:    5 * time.Minute,           // Connection idle timeout
    MaxConnecting:  10,                         // Max concurrent connections
    ReadPreference: "primary",                  // Read preference
    RetryWrites:    true,                       // Enable write retries
    RetryReads:     true,                       // Enable read retries
}
```

**Read Preference Options:**

- `"primary"` - Read from primary only
- `"secondary"` - Read from secondary only
- `"primaryPreferred"` - Prefer primary, fallback to secondary
- `"secondaryPreferred"` - Prefer secondary, fallback to primary
- `"nearest"` - Read from nearest member

### PostgreSQL Configuration

```go
config := sql.GormConfig{
    Host:                      "localhost",              // Database host
    User:                      "postgres",               // Username
    Password:                  "password",               // Password
    Name:                      "myapp",                  // Database name
    Port:                      "5432",                   // Port
    SSLMode:                   "disable",                // SSL mode
    Timezone:                  "UTC",                    // Timezone
    MaxIdleConns:              10,                       // Max idle connections
    MaxOpenConns:              100,                      // Max open connections
    ConnMaxLifetime:           30 * time.Minute,        // Connection lifetime
    ConnMaxIdleTime:           10 * time.Minute,        // Idle timeout
    TranslateErrors:           true,                     // Translate errors
    LogLevel:                  logger.Info,              // Log level
    SlowThreshold:             200 * time.Millisecond,  // Slow query threshold
    IgnoreRecordNotFoundError: false,                    // Ignore not found errors
}
```

**SSL Mode Options:**

- `"disable"` - No SSL
- `"require"` - SSL required
- `"verify-ca"` - Verify CA certificate
- `"verify-full"` - Full certificate verification

**Log Levels:**

- `logger.Silent` - No logging
- `logger.Error` - Error level only
- `logger.Warn` - Warning and above
- `logger.Info` - Info and above

## ❤️ Health Checks

### MongoDB Health Check

```go
// Simple health check
healthy := client.IsHealthy(context.Background())
log.Printf("MongoDB healthy: %t", healthy)

// Detailed ping with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := client.Ping(ctx); err != nil {
    log.Printf("MongoDB ping failed: %v", err)
} else {
    log.Println("MongoDB ping successful")
}
```

### PostgreSQL Health Check

```go
// Connection health check
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := db.Ping(ctx); err != nil {
    log.Printf("PostgreSQL ping failed: %v", err)
} else {
    log.Println("PostgreSQL is healthy")
}

// Get connection statistics
stats := db.Stats()
log.Printf(`
Database Statistics:
- Open Connections: %d
- In Use: %d
- Idle: %d
- Wait Count: %d
- Wait Duration: %v
- Max Idle Closed: %d
- Max Lifetime Closed: %d
`,
    stats.OpenConnections,
    stats.InUse,
    stats.Idle,
    stats.WaitCount,
    stats.WaitDuration,
    stats.MaxIdleClosed,
    stats.MaxLifetimeClosed,
)
```

## 🚨 Error Handling

### MongoDB Error Handling

```go
import (
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// Handle duplicate key errors
_, err := collection.InsertOne(ctx, document)
if err != nil {
    if mongo.IsDuplicateKeyError(err) {
        log.Println("Document already exists")
        return ErrDuplicateDocument
    }
    return fmt.Errorf("failed to insert document: %w", err)
}

// Handle write concern errors
writeConcern := writeconcern.New(writeconcern.WMajority(), writeconcern.J(true))
opts := options.Insert().SetWriteConcern(writeConcern)

_, err = collection.InsertOne(ctx, document, opts)
if err != nil {
    if writeconcern.IsWriteConcernError(err) {
        log.Println("Write concern not satisfied")
    }
    return err
}
```

### PostgreSQL Error Handling

```go
import (
    "errors"
    "gorm.io/gorm"
)

// Handle not found errors
var user User
err := db.First(&user, "email = ?", email).Error
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return ErrUserNotFound
    }
    return fmt.Errorf("failed to find user: %w", err)
}

// Handle constraint violations
err = db.Create(&user).Error
if err != nil {
    if strings.Contains(err.Error(), "duplicate key") {
        return ErrEmailAlreadyExists
    }
    return fmt.Errorf("failed to create user: %w", err)
}

// Transaction error handling
err = sql.WithTransaction(db.DB, func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }

    if err := tx.Create(&profile).Error; err != nil {
        return fmt.Errorf("failed to create profile: %w", err)
    }

    return nil
})

if err != nil {
    log.Printf("Transaction failed: %v", err)
    // Transaction was automatically rolled back
}
```

## 🌟 Best Practices

### Database Design

```go
// Use appropriate base models
type User struct {
    sql.BaseModel  // For entities that need full audit trail
    Name  string
    Email string
}

type UserSession struct {
    sql.ID         // For simple entities that only need ID
    sql.Timestamp  // Add timestamps manually when needed
    UserID uuid.UUID
    Token  string
}

// Use JSONB for flexible data
type UserPreferences struct {
    sql.BaseModel
    UserID uuid.UUID `gorm:"type:uuid;unique;not null"`
    Data   sql.JSONB `gorm:"type:jsonb;not null;default:'{}'"`
}

// Index JSONB fields for better performance
// CREATE INDEX idx_user_preferences_theme ON user_preferences USING GIN ((data->'theme'));
```

### Connection Management

```go
// Configure connection pools appropriately
config := sql.GormConfig{
    MaxIdleConns:    10,              // Keep connections ready
    MaxOpenConns:    100,             // Limit total connections
    ConnMaxLifetime: 30 * time.Minute, // Prevent stale connections
    ConnMaxIdleTime: 10 * time.Minute, // Close idle connections
}

// Use transactions for data integrity
err := sql.WithTransaction(db.DB, func(tx *gorm.DB) error {
    // Group related operations
    if err := tx.Create(&user).Error; err != nil {
        return err
    }

    if err := tx.Create(&userProfile).Error; err != nil {
        return err
    }

    return nil
})
```

### Performance Optimization

```go
// Use preloading to avoid N+1 queries
var users []User
db.Preload("Profile").Find(&users)

// Use select to limit fields
db.Select("id", "name", "email").Find(&users)

// Use appropriate indexes
type User struct {
    sql.BaseModel
    Email string `gorm:"uniqueIndex;not null"`
    Name  string `gorm:"index"`
}

// Batch operations for better performance
var users []User
db.CreateInBatches(users, 100)
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file
for details.

## 🙏 Acknowledgments

- Built on top of [GORM](https://gorm.io/) for PostgreSQL support
- Uses the official
  [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver)
- Inspired by modern database patterns and best practices
- Designed for production reliability and developer experience

---

**Note**: This package requires Go 1.22 or later for UUID support and modern
language features.
