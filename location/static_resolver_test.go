package location

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStaticResolverWorks(t *testing.T) {
	resolver := StaticResolver()
	country, err := resolver.ResolveCountry("46.111.111.99")
	assert.NoError(t, err)
	assert.Equal(t, "RU", country)
}
