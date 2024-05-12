# Kafka consumer with http sendler

Kafka's consumer implementation service. After the application is launched, asynchronous cache filling begins for future data enrichment. It then runs an infinite loop that listens to the kafka topic, validates and processes the data, and sends it over http asynchronously to a third-party service.

## Getting Started

To get started with the project, follow these steps:

1. Clone the repository:

```
git clone https://github.com/PureLeach/go-kafka-consumer-with-http-sendler.git
```

2. Copy the settings file and change them:

```
cp example.env .env
```

3. Build the project:

```
make build   
```

4. Run the project:

```
make start
```

