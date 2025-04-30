package logger

import (
	"os"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	elogrus "gopkg.in/sohlich/elogrus.v7"
)

var Logger = logrus.New()

// InitLogger initializes the Logrus logger with an Elasticsearch hook
func InitLogger(serviceName string) {

	// Elasticsearch URL
	elasticURL := "http://elasticsearch:9200"

	// Create an Elasticsearch client
	client, err := elastic.NewClient(elastic.SetURL(elasticURL), elastic.SetSniff(false))
	if err != nil {
		logrus.Fatalf("Failed to create Elasticsearch client: %v", err)
	}

	// Create an ELK hook for Logrus
	hook, err := elogrus.NewElasticHook(client, serviceName, logrus.InfoLevel, serviceName+"-logs")
	if err != nil {
		logrus.Fatalf("Failed to create Elasticsearch hook: %v", err)
	}

	// Add the hook to Logrus
	Logger.Hooks.Add(hook)

	// Set Logrus formatter
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	// Set Logrus output to stdout
	Logger.SetOutput(os.Stdout)

	// Set Logrus log level
	Logger.SetLevel(logrus.InfoLevel)

	Logger.WithField("service", serviceName).Info("Logger initialized and connected to Elasticsearch")
}
