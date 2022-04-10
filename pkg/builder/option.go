package builder

// Option for configing the builder
type Option func(c *Client)

// SetRegistry will sent the registry which is used to build the Docker Image Name
func SetRegistry(registry string) Option {
	return func(c *Client) {
		c.registry = registry
	}
}

// SetTag used part of the Docker Image Name
func SetTag(tag string) Option {
	return func(c *Client) {
		c.tag = tag
	}
}

// SetNetwork use this network when building
func SetNetwork(network string) Option {
	return func(c *Client) {
		c.network = network
	}
}

func SetPoolSize(poolSize int) Option {
	return func(c *Client) {
		c.poolSize = poolSize
	}
}
