package internal

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm/schema"
)

type Function[T interface{}] func(*context.Context, T) error

func ParseUUIDString(uid string) (*uuid.UUID, error) {
	uidParsed, err := uuid.Parse(uid)
	if err != nil {
		return nil, err
	}
	return &uidParsed, nil
}

func GenerateUUIdV6() uuid.UUID {
	uuid, err := uuid.NewV6()
	if err != nil {
		panic(err)
	}
	return uuid
}

func XOR(a []byte, b []byte) []byte {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	res := make([]byte, n)
	for i := 0; i < n; i++ {
		res[i] = a[i] ^ b[i]
	}
	return res
}

// Generate unique uuid from 2 uuid's using XOR Operation which is commutative
// @param uuid1 uuid.UUID:
// @param uuid2 uuid.UUID:
// @return uuid.UUID
func GenerateUUIDFromUUIDs(uid1 uuid.UUID, uid2 uuid.UUID) uuid.UUID {
	u1, err := uid1.MarshalBinary()
	if err != nil {
		panic(err)
	}
	u2, err := uid2.MarshalBinary()
	if err != nil {
		panic(err)
	}
	uidBytes := XOR(u1, u2)
	uid, err := uuid.FromBytes(uidBytes)
	if err != nil {
		panic(err)
	}
	uidString := uid.String()
	uid = uuid.NewSHA1(uuid.NameSpaceX500, []byte(uidString))
	return uid
}

func GetDbFieldName(fieldName string, entity interface{}) (string, error) {
	s, err := schema.Parse(entity, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return "", err
	}
	for _, field := range s.Fields {
		dbName := field.DBName
		modelName := field.Name
		if modelName == fieldName {
			return dbName, nil
		}
	}
	return "", nil
}

// Run task concurrently using goroutines
// @param fn Function to run
// @param inputs Inputs to pass to the function
// @return error If any of the tasks return an error
func Parallelize[I interface{}](ctx *context.Context, fn Function[I], inputs []I) error {
	errCh := make(chan error, len(inputs))
	for _, input := range inputs {
		go func() {
			errCh <- fn(ctx, input)
		}()
	}
	for range inputs {
		err := <-errCh
		if err != nil {
			return err
		}
	}
	return nil
}
