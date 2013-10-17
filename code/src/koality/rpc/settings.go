package rpc

const (
	amqpUri = "amqp://localhost:5672/"
)

const (
	exchangeName       = "model:rpc" // TODO: change to "rpc"
	exchangeType       = "direct"
	exchangeDurable    = false // TODO: make durable?
	exchangeAutoDelete = false
	exchangeInternal   = false
	exchangeNoWait     = false
	exchangeMandatory  = true
	exchangeImmediate  = false
)

const (
	deadLetterExchangeName       = "model:rpc_dlx" // TODO: change to "rpc-dlx"
	deadLetterExchangeType       = "fanout"
	deadLetterExchangeDurable    = false // TODO: make durable?
	deadLetterExchangeAutoDelete = false
	deadLetterExchangeInternal   = false
	deadLetterExchangeNoWait     = false
	deadLetterExchangeTtl        = 10000
)

const (
	responseQueueName       = ""
	responseQueueDurable    = false
	responseQueueAutoDelete = true
	responseQueueExclusive  = true
	responseQueueNoWait     = false
	responseQueueAutoAck    = true
	responseQueueNoLocal    = true
)

const (
	deadLetterQueueName       = ""
	deadLetterQueueDurable    = false
	deadLetterQueueAutoDelete = true
	deadLetterQueueExclusive  = true
	deadLetterQueueNoWait     = false
	deadLetterQueueBindNoWait = false
	deadLetterQueueAutoAck    = true
	deadLetterQueueNoLocal    = true
)
