package config

type Stripe struct {
	Key string `yaml:"key" json:"key" mapstructure:"key"`
}

func NewStripe() *Stripe {
	return &Stripe{}
}

func (stripe *Stripe) SetKey(key string) {
	stripe.Key = key
}

func (stripe *Stripe) GetKeyu() string {
	return stripe.Key
}
