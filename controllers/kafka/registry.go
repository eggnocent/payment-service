package kafka

type Registry struct {
	brokers []string
}

type IKafkaRegistry interface {
	GetKafkaProducer() IKafka
}

func NewRegistry(brokers []string) IKafkaRegistry {
	return &Registry{
		brokers: brokers,
	}
}

func (r *Registry) GetKafkaProducer() IKafka {
	return NewKafkaProducer(r.brokers)
}
