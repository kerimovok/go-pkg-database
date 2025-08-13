package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoConfig struct {
	URI            string
	DBName         string
	Timeout        time.Duration
	MaxPoolSize    uint64
	MinPoolSize    uint64
	MaxIdleTime    time.Duration
	MaxConnecting  uint64
	ReadPreference string // primary, secondary, primaryPreferred, secondaryPreferred, nearest
	RetryWrites    bool
	RetryReads     bool
}

// Client wraps mongo.Client and provides additional functionality
type Client struct {
	Client *mongo.Client
	config MongoConfig
}

func (c MongoConfig) validate() error {
	if c.URI == "" {
		return fmt.Errorf("URI is required")
	}
	if c.DBName == "" {
		return fmt.Errorf("database name is required")
	}
	return nil
}

func (c MongoConfig) setDefaults() MongoConfig {
	if c.Timeout <= 0 {
		c.Timeout = 10 * time.Second
	}
	if c.MaxPoolSize == 0 {
		c.MaxPoolSize = 100
	}
	if c.MinPoolSize == 0 {
		c.MinPoolSize = 5
	}
	if c.MaxIdleTime <= 0 {
		c.MaxIdleTime = 5 * time.Minute
	}
	if c.MaxConnecting == 0 {
		c.MaxConnecting = 10
	}
	if c.ReadPreference == "" {
		c.ReadPreference = "primary"
	}
	return c
}

func (c MongoConfig) timeout() time.Duration {
	if c.Timeout <= 0 {
		return 10 * time.Second
	}
	return c.Timeout
}

func (c MongoConfig) getReadPreference() (*readpref.ReadPref, error) {
	switch c.ReadPreference {
	case "primary":
		return readpref.Primary(), nil
	case "secondary":
		return readpref.Secondary(), nil
	case "primaryPreferred":
		return readpref.PrimaryPreferred(), nil
	case "secondaryPreferred":
		return readpref.SecondaryPreferred(), nil
	case "nearest":
		return readpref.Nearest(), nil
	default:
		return nil, fmt.Errorf("invalid read preference: %s", c.ReadPreference)
	}
}

func Connect(cfg MongoConfig) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	cfg = cfg.setDefaults()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout())
	defer cancel()

	// Configure client options
	opts := options.Client().ApplyURI(cfg.URI)

	// Set connection pool options
	opts.SetMaxPoolSize(cfg.MaxPoolSize)
	opts.SetMinPoolSize(cfg.MinPoolSize)
	opts.SetMaxConnIdleTime(cfg.MaxIdleTime)
	opts.SetMaxConnecting(cfg.MaxConnecting)

	// Set read preference
	readPref, err := cfg.getReadPreference()
	if err != nil {
		return nil, fmt.Errorf("failed to set read preference: %w", err)
	}
	opts.SetReadPreference(readPref)

	// Set retry options
	opts.SetRetryWrites(cfg.RetryWrites)
	opts.SetRetryReads(cfg.RetryReads)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, readPref); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &Client{
		Client: client,
		config: cfg,
	}, nil
}

// Disconnect closes the MongoDB connection
func (c *Client) Disconnect(ctx context.Context) error {
	if c == nil || c.Client == nil {
		return nil
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), c.config.timeout())
		defer cancel()
	}
	return c.Client.Disconnect(ctx)
}

// Ping tests the MongoDB connection
func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.Client == nil {
		return fmt.Errorf("client not initialized")
	}
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), c.config.timeout())
		defer cancel()
	}

	readPref, err := c.config.getReadPreference()
	if err != nil {
		readPref = readpref.Primary()
	}

	return c.Client.Ping(ctx, readPref)
}

// Database returns a handle to the configured database
func (c *Client) Database() *mongo.Database {
	if c == nil || c.Client == nil {
		return nil
	}
	return c.Client.Database(c.config.DBName)
}

// Collection returns a handle to a collection in the configured database
func (c *Client) Collection(name string) *mongo.Collection {
	db := c.Database()
	if db == nil {
		return nil
	}
	return db.Collection(name)
}

// IsHealthy checks if the MongoDB connection is healthy
func (c *Client) IsHealthy(ctx context.Context) bool {
	return c.Ping(ctx) == nil
}

// WithTransaction executes a function within a MongoDB transaction
func (c *Client) WithTransaction(ctx context.Context, fn func(mongo.SessionContext) error) error {
	if c == nil || c.Client == nil {
		return fmt.Errorf("client not initialized")
	}

	session, err := c.Client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		return nil, fn(sc)
	})

	return err
}
