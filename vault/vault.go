package vault

import (
	"fmt"
	"log"

	"github.com/hashicorp/vault/api"
)

func GetSecret(key string) (string, error) {
	client, err := api.NewClient(&api.Config{
		Address: "http://vault:8200",
	})
	if err != nil {
		log.Fatalf("Failed to create Vault client: %v", err)
		return "", err
	}

	client.SetToken("root") // Use the root token for development

	secret, err := client.Logical().Read("secret/ecommerce")
	if err != nil {
		return "", err
	}

	if secret == nil || secret.Data[key] == nil {
		return "", nil
	}

	return secret.Data[key].(string), nil
}
func InitConfig() {
	databaseURL, err := GetSecret("DATABASE_URL")
	if err != nil {
		log.Fatalf("Failed to retrieve DATABASE_URL: %v", err)
	}

	jwtSecret, err := GetSecret("JWT_SECRET")
	if err != nil {
		log.Fatalf("Failed to retrieve JWT_SECRET: %v", err)
	}

	fmt.Printf("Database URL: %s\n", databaseURL)
	fmt.Printf("JWT Secret: %s\n", jwtSecret)
}
