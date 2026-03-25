package cacher

import (
	"context"
	"crypto/sha256"
	"time"

	"google.golang.org/protobuf/proto"
)

const (
	defaultCachedProtoSwapPeriod       = 5 * time.Minute
	cachedProtoObjectsCountGaugePeriod = 5 * time.Second
)

var unmarshaller = newCachedProtoUnmarshall(context.Background(), "", defaultCachedProtoSwapPeriod)

func hashKey(protoName string, marshalledVal []byte) string {
	// hash the proto name and the marshalled value
	// hashing the protoName is necessary to avoid collisions between different models that have the same data values
	h := sha256.New()
	protoName += "::"
	h.Write([]byte(protoName))
	h.Write(marshalledVal)

	return string(h.Sum(nil))
}

func cachedUnmarshal(bytes []byte, message proto.Message) (proto.Message, error) {
	return unmarshaller.unmarshall(bytes, message)
}

type cachedProtoUnmarshaller struct {
	name  string
	store *swapMap
}

func newCachedProtoUnmarshall(ctx context.Context, name string, ttl time.Duration) *cachedProtoUnmarshaller {
	result := &cachedProtoUnmarshaller{
		name:  name,
		store: newSwapMap(),
	}

	go func() {
		ticker := time.NewTicker(ttl)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				result.store.swap()
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(cachedProtoObjectsCountGaugePeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	return result
}

func (c *cachedProtoUnmarshaller) unmarshall(bytes []byte, message proto.Message) (proto.Message, error) {
	protoName := string(message.ProtoReflect().Descriptor().FullName())
	bytesHash := hashKey(protoName, bytes)

	cachedUnmarshalled, ok := c.store.load(bytesHash)
	if ok {
		protoMsg, ok := cachedUnmarshalled.(proto.Message)
		if !ok {
			return nil, proto.Error
		}

		return protoMsg, nil
	}

	err := proto.Unmarshal(bytes, message)
	if err != nil {
		return nil, err
	}

	c.store.store(bytesHash, message)

	return message, nil
}
