package services

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/vovanwin/api-my-site/ent"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"

	// Required by ent
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/vovanwin/api-my-site/config"
	// Require by ent
	_ "github.com/vovanwin/api-my-site/ent/runtime"
)

// Container содержит все службы, используемые приложением, и предоставляет простой способ обработки зависимостей
// инъекция, в том числе в рамках тестов
type Container struct {
	// Validator stores a validator
	Validator *Validator

	// Web хранит веб-фреймворк
	Web *echo.Echo

	// Config stores the application configuration
	Config *config.Config

	// Cache contains the cache client
	Cache *CacheClient

	// Database stores the connection to the database
	Database *sql.DB

	// ORM stores a client to the ORM
	ORM *ent.Client

	// Auth stores an authentication client
	Auth *AuthClient

	// Tasks stores the task client
	Tasks *TaskClient
}

// NewContainer creates and initializes a new Container
func NewContainer() *Container {
	c := new(Container)
	c.initConfig()
	c.initValidator()
	c.initWeb()
	c.initCache()
	c.initDatabase()
	c.initORM()
	c.initAuth()
	c.initTasks()
	return c
}

// Shutdown shuts the Container down and disconnects all connections
func (c *Container) Shutdown() error {
	if err := c.Tasks.Close(); err != nil {
		return err
	}
	if err := c.Cache.Close(); err != nil {
		return err
	}
	if err := c.ORM.Close(); err != nil {
		return err
	}
	if err := c.Database.Close(); err != nil {
		return err
	}

	return nil
}

// initConfig initializes configuration
func (c *Container) initConfig() {
	cfg, err := config.GetConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	c.Config = &cfg
}

// initValidator initializes the validator
func (c *Container) initValidator() {
	c.Validator = NewValidator()
}

// initWeb initializes the web framework
func (c *Container) initWeb() {
	c.Web = echo.New()

	// Configure logging
	switch c.Config.App.Environment {
	case config.EnvProduction:
		c.Web.Logger.SetLevel(log.WARN)
	default:
		c.Web.Logger.SetLevel(log.DEBUG)
	}

	c.Web.Validator = c.Validator
}

// initCache initializes the cache
func (c *Container) initCache() {
	var err error
	if c.Cache, err = NewCacheClient(c.Config); err != nil {
		panic(err)
	}
}

// initDatabase initializes the database
// If the environment is set to test, the test database will be used and will be dropped, recreated and migrated
func (c *Container) initDatabase() {
	var err error

	getAddr := func(dbName string) string {
		return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
			c.Config.Database.User,
			c.Config.Database.Password,
			c.Config.Database.Hostname,
			c.Config.Database.Port,
			dbName,
		)
	}

	c.Database, err = sql.Open("pgx", getAddr(c.Config.Database.Database))
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	// Check if this is a test environment
	if c.Config.App.Environment == config.EnvTest {
		// Drop the test database, ignoring errors in case it doesn't yet exist
		_, _ = c.Database.Exec("DROP DATABASE " + c.Config.Database.TestDatabase)

		// Create the test database
		if _, err = c.Database.Exec("CREATE DATABASE " + c.Config.Database.TestDatabase); err != nil {
			panic(fmt.Sprintf("не удалось создать тестовую базу данных: %v", err))
		}

		// Connect to the test database
		if err = c.Database.Close(); err != nil {
			panic(fmt.Sprintf("не удалось закрыть подключение к базе данных: %v", err))
		}
		c.Database, err = sql.Open("pgx", getAddr(c.Config.Database.TestDatabase))
		if err != nil {
			panic(fmt.Sprintf("не удалось подключиться к базе данных: %v", err))
		}
	}
}

// initORM initializes the ORM
func (c *Container) initORM() {
	drv := entsql.OpenDB(dialect.Postgres, c.Database)
	c.ORM = ent.NewClient(ent.Driver(drv))
	if err := c.ORM.Schema.Create(context.Background(), schema.WithAtlas(true)); err != nil {
		panic(fmt.Sprintf("не удалось создать базу данных schema: %v", err))
	}
}

// initAuth initializes the authentication client
func (c *Container) initAuth() {
	c.Auth = NewAuthClient(c.Config, c.ORM)
}

// initTasks initializes the task client
func (c *Container) initTasks() {
	c.Tasks = NewTaskClient(c.Config)
}
