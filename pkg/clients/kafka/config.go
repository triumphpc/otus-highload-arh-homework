package kafka

type Config struct {
	Address          string `env:"KAFKA_ADDRESS" env-default:"localhost:9092"`
	FeedUpdatesTopic string `env:"KAFKA_FEED_UPDATES_TOPIC" env-default:"feed_updates"`
}
