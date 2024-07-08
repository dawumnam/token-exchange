package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PublicHost             string
	Port                   string
	DBUser                 string
	DBPassword             string
	DBAddress              string
	DBName                 string
	JWTExpirationInSeconds int64
	JWTSecret              string
	ChainPrivateKey        string
	PlatformAddress        string
}

var Envs = initConfig()

func initConfig() Config {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf(".env file does not exist")
	}

	return Config{
		PublicHost:             getEnv("PUBLIC_HOST", "http://localhost"),
		Port:                   getEnv("PORT", "8080"),
		DBUser:                 getEnv("DB_USER", "user"),
		DBPassword:             getEnv("DB_PASSWORD", "mypassword"),
		DBAddress:              fmt.Sprintf("%s:%s", getEnv("DB_HOST", "127.0.0.1"), getEnv("DB_PORT", "3306")),
		DBName:                 getEnv("DB_NAME", "token-trader"),
		JWTExpirationInSeconds: getIntEnv("JWT_EXP", 3600*24*1),
		JWTSecret:              getEnv("JWT_SECRET", "somesecret"),
		ChainPrivateKey:        getEnv("CHAIN_PK", "ad80f301c7c1f30bffd51128638d20f6dde70245a5fe5b4ef9560c7d157bf150"),
		PlatformAddress:        getEnv("PLATFORM_ADDR", "0x066322cE1C277E30b1c885D24692D66A186073EE"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getIntEnv(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		someInt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}
		return someInt
	}

	return fallback
}
