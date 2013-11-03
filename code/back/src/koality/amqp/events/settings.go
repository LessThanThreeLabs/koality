package events

const (
	exchangeName       = "events"
	exchangeType       = "direct"
	exchangeDurable    = false
	exchangeAutoDelete = false
	exchangeInternal   = false
	exchangeNoWait     = false
	exchangeMandatory  = true
	exchangeImmediate  = false
)

const (
	receiveQueueName       = ""
	receiveQueueDurable    = false
	receiveQueueAutoDelete = true
	receiveQueueExclusive  = true
	receiveQueueNoWait     = false
	receiveQueueBindNoWait = false
	receiveQueueAutoAck    = false
	receiveQueueNoLocal    = true
	receiveQueueQos        = 3
)
