package rpc

const (
	amqpUri = "amqp://localhost:5672/"
)

const (
	exchangeName       = "rpc"
	exchangeType       = "direct"
	exchangeDurable    = false
	exchangeAutoDelete = false
	exchangeInternal   = false
	exchangeNoWait     = false
	exchangeMandatory  = true
	exchangeImmediate  = false
)

const (
	deadLetterExchangeName       = "rpc-dlx"
	deadLetterExchangeType       = "fanout"
	deadLetterExchangeDurable    = false
	deadLetterExchangeAutoDelete = false
	deadLetterExchangeInternal   = false
	deadLetterExchangeNoWait     = false
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
	deadLetterQueueAutoAck    = true
	deadLetterQueueNoLocal    = true
)
