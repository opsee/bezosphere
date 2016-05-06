package spanxcreds

import (
	"testing"
	"time"

	"github.com/opsee/basic/schema"
	"github.com/stretchr/testify/assert"
)

const TestCustomerId = "00000000-0000-0000-0000-000000000000"

func TestSpanxProviderRetrieve(t *testing.T) {
	p := &SpanxProvider{
		user: &schema.User{
			CustomerId: TestCustomerId,
		},
		Client: &SpanxTestClient{
			ExpiryString: "2014-12-16T01:51:37Z",
		},
	}
	creds, err := p.Retrieve()

	assert.Nil(t, err, "Expect no error, %v", err)
	assert.Equal(t, "accessKey", creds.AccessKeyID, "Expect access key ID to match")
	assert.Equal(t, "secret", creds.SecretAccessKey, "Expect secret access key to match")
	assert.Equal(t, "token", creds.SessionToken, "Expect session token to match")
}

func TestSpanxProviderIsExpired(t *testing.T) {
	p := &SpanxProvider{
		user: &schema.User{
			CustomerId: TestCustomerId,
		},
		Client: &SpanxTestClient{
			ExpiryString: "2014-12-16T01:51:37Z",
		},
	}
	p.CurrentTime = func() time.Time {
		return time.Date(2014, 12, 15, 21, 26, 0, 0, time.UTC)
	}

	assert.True(t, p.IsExpired(), "Expect creds to be expired before retrieve.")

	_, err := p.Retrieve()
	assert.Nil(t, err, "Expect no error, %v", err)

	assert.False(t, p.IsExpired(), "Expect creds to not be expired after retrieve.")

	p.CurrentTime = func() time.Time {
		return time.Date(3014, 12, 15, 21, 26, 0, 0, time.UTC)
	}

	assert.True(t, p.IsExpired(), "Expect creds to be expired.")
}

func TestSpanxProviderExpiryWindowIsExpired(t *testing.T) {
	p := &SpanxProvider{
		user: &schema.User{
			CustomerId: TestCustomerId,
		},
		Client: &SpanxTestClient{
			ExpiryString: "2014-12-16T01:51:37Z",
		},
		ExpiryWindow: time.Hour * 1,
	}
	p.CurrentTime = func() time.Time {
		return time.Date(2014, 12, 15, 0, 51, 37, 0, time.UTC)
	}

	assert.True(t, p.IsExpired(), "Expect creds to be expired before retrieve.")

	_, err := p.Retrieve()
	assert.Nil(t, err, "Expect no error, %v", err)

	assert.False(t, p.IsExpired(), "Expect creds to not be expired after retrieve.")

	p.CurrentTime = func() time.Time {
		return time.Date(2014, 12, 16, 0, 55, 37, 0, time.UTC)
	}

	assert.True(t, p.IsExpired(), "Expect creds to be expired.")
}
