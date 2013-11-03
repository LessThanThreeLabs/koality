package rpc

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
	deadLetterExchangeName       = "rpc_dlx"
	deadLetterExchangeType       = "fanout"
	deadLetterExchangeDurable    = false
	deadLetterExchangeAutoDelete = false
	deadLetterExchangeInternal   = false
	deadLetterExchangeNoWait     = false
	deadLetterExchangeTtl        = 10000
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

const (
	clientResponseQueueName       = ""
	clientResponseQueueDurable    = false
	clientResponseQueueAutoDelete = true
	clientResponseQueueExclusive  = true
	clientResponseQueueNoWait     = false
	clientResponseQueueAutoAck    = true
	clientResponseQueueNoLocal    = true
)

const (
	serverResponseQueueNamePrefix = "rpc:"
	serverResponseQueueDurable    = false
	serverResponseQueueAutoDelete = true
	serverResponseQueueExclusive  = false
	serverResponseQueueNoWait     = false
	serverResponseQueueBindNoWait = false
	serverResponseQueueAutoAck    = false
	serverResponseQueueNoLocal    = true
	serverResponseQueueQos        = 3
)
