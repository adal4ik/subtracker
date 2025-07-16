package loadenv

import (
	"github.com/joho/godotenv"
)

func LoadEnvFile(filename string) error {
	return godotenv.Load(filename)
}
