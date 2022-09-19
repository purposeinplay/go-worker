package redisdocker

import (
	"fmt"
	"io"

	"github.com/gomodule/redigo/redis"
	"github.com/ory/dockertest/v3"
)

// ensure Container implements the io.Closer interface.
var _ io.Closer = (*Container)(nil)

// Container represents a Docker container
// running a PostgreSQL image.
type Container struct {
	res  *dockertest.Resource
	port string
}

// Port returns the container host port mapped
// to the database running inside it.
func (c Container) Port() string {
	return c.port
}

// Close removes the Docker container.
func (c Container) Close() error {
	return c.res.Close()
}

// NewContainer starts a new redis client in a docker container.
func NewContainer() (
	*Container,
	error,
) {
	var err error

	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker %w", err)
	}

	resource, err := pool.Run("redis", "3.2", nil)
	if err != nil {
		return nil, fmt.Errorf("could not start resource: %w", err)
	}

	err = pool.Retry(func() error {
		addr := fmt.Sprintf("localhost:%s", resource.GetPort("6379/tcp"))

		conn, err := redis.Dial("tcp", addr)
		if err != nil {
			return err
		}

		_, err = redis.String(conn.Do("PING"))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	return &Container{
		res:  resource,
		port: resource.GetPort("6379/tcp"),
	}, nil
}
