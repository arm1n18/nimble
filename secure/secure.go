package secure

import (
	"nimble/formatter"
	"nimble/storage"
	"os"

	"github.com/joho/godotenv"
)

type Session struct {
	Authorized bool
}

func getPassword() string {
	if err := godotenv.Load("../../.env"); err != nil {
		formatter.FatalError("Failed to load .env file: %v", err)
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

	// parts := strings.Fields(password)
	// res := regexp.MustCompile(`"[^"]*"|\S+`)
	// if len(parts) == 0 {
	// 	return false
	// }

	// p := strings.Join(res.FindAllString(strings.Join(parts, " "), -1), " ")

	return getPassword() == password
}

func SetPassword(password string) {
	env, err := godotenv.Read(".env")
	if err != nil {
		formatter.FatalError("Failed to load .env file: %v", err)
	}

	env["PASSWORD"] = password

	if err := godotenv.Write(env, ".env"); err != nil {
		formatter.ErrorMessage("Failed to write .env file: %v", err)
		return
	}

	formatter.SuccessMessage("Password set successfully")
}

func ReadOnlyMiddleware(c *storage.Cache, fn func() string) string {
	if c.GetMode() == storage.ReadOnly {
		return formatter.ErrorMessage("Cache is in read-only mode")
	}
	return fn()
}

func ChangeMode(c *storage.Cache, mode string) string {
	var result string

	switch mode {
	case "READONLY":
		c.SetMode(storage.ReadOnly)
		result = formatter.SuccessMessage("Cache mode changed to READONLY")
	case "READWRITE":
		c.SetMode(storage.ReadWrite)
		result = formatter.SuccessMessage("Cache mode changed to READWRITE")
	default:
		result = formatter.ErrorMessage("Unknown mode: %s", mode)
	}

	return result
}
