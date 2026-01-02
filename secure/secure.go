package secure

import (
	"cache/logger"
	"cache/storage"
	"os"

	"github.com/joho/godotenv"
)

func getPassword() string {
	if err := godotenv.Load(".env"); err != nil {
		logger.FatalError("Failed to load .env file: %v", err)
	}

	pass := os.Getenv("PASSWORD")

	return pass
}

func SecureConnection(c *storage.Cache) bool {
	return getPassword() != ""
}

func Authenticate(password string) bool {
	if getPassword() == "" {
		return true
	}

	return getPassword() == password
}

func SetPassword(password string) {
	env, err := godotenv.Read(".env")
	if err != nil {
		logger.FatalError("Failed to load .env file: %v", err)
	}

	env["PASSWORD"] = password

	if err := godotenv.Write(env, ".env"); err != nil {
		logger.Error("Failed to write .env file: %v", err)
		return
	}

	logger.Success("Password set successfully")
}

func ReadOnlyMiddleware(c *storage.Cache, fn func()) {
	if c.GetMode() == storage.ReadOnly {
		logger.Error("Cache is in read-only mode")
		return
	}
	fn()
}

func ChangeMode(c *storage.Cache, mode string) {
	switch mode {
	case "READONLY":
		c.SetMode(storage.ReadOnly)
		logger.Success("Cache mode changed to READONLY")
	case "READWRITE":
		c.SetMode(storage.ReadWrite)
		logger.Success("Cache mode changed to READWRITE")
	default:
		logger.Error("Unknown mode: %s", mode)
	}
}
