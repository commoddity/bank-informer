package persistence

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/commoddity/bank-informer/env"
	badger "github.com/dgraph-io/badger/v3"
)

// Set a TTL of 72 hours for all data
const ttl = 72 * time.Hour

type (
	Persistence struct {
		DB *badger.DB
	}
	IPersistence interface {
		Close() error

		GetAverageCryptoValues(key string) (CryptoValues, error)
		WriteCryptoValues(key string, value CryptoValues) error
		ClearOldEntries() error
	}
)

func NewPersistence() *Persistence {
	opts := badger.DefaultOptions(env.DBPath)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal("error opening badger db")
	}

	return &Persistence{DB: db}
}

func (p *Persistence) Close() error {
	return p.DB.Close()
}

type CryptoValues struct {
	CryptoBalance float64 `json:"cryptoBalance"`
	FiatValue     float64 `json:"fiatValue"`
	FiatBalance   float64 `json:"fiatBalance"`
}

func (p *Persistence) GetAverageCryptoValues(key string) (CryptoValues, error) {
	var result CryptoValues
	err := p.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			// Handle the case where the key does not exist
			return err
		}

		return item.Value(func(val []byte) error {
			cryptoValues, err := deserializeCryptoValuesSlice(val)
			if err != nil {
				return err
			}

			var sumCryptoBalance, sumFiatValue, sumFiatBalance float64
			for _, cv := range cryptoValues {
				sumCryptoBalance += cv.CryptoBalance
				sumFiatValue += cv.FiatValue
				sumFiatBalance += cv.FiatBalance
			}
			result = CryptoValues{
				CryptoBalance: sumCryptoBalance / float64(len(cryptoValues)),
				FiatValue:     sumFiatValue / float64(len(cryptoValues)),
				FiatBalance:   sumFiatBalance / float64(len(cryptoValues)),
			}
			return nil
		})
	})

	return result, err
}

func (p *Persistence) WriteCryptoValues(key string, value CryptoValues) error {
	return p.DB.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		var values []CryptoValues

		if err == nil {
			// Key exists, retrieve and deserialize current data
			err = item.Value(func(val []byte) error {
				values, err = deserializeCryptoValuesSlice(val)
				return err
			})
			if err != nil {
				return err
			}
		}

		// Append new value
		values = append(values, value)

		// Serialize updated slice and write back to DB
		data, err := serializeCryptoValuesSlice(values)
		if err != nil {
			return err
		}

		// Use SetEntry to write data with TTL
		e := badger.NewEntry([]byte(key), data).WithTTL(ttl)
		return txn.SetEntry(e)
	})
}

func serializeCryptoValuesSlice(values []CryptoValues) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(values)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeCryptoValuesSlice(data []byte) ([]CryptoValues, error) {
	var values []CryptoValues
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&values)
	if err != nil {
		return nil, err
	}
	return values, nil
}

type KeyValue struct {
	Key    string       `json:"key"`
	Values CryptoValues `json:"values"`
}

func (p *Persistence) ReadAll() (map[string]CryptoValues, error) {
	averages := make(map[string]CryptoValues)

	err := p.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			err := item.Value(func(val []byte) error {
				cryptoValues, err := deserializeCryptoValuesSlice(val)
				if err != nil {
					return fmt.Errorf("Error deserializing data for key %s: %v", key, err)
				}

				var sumCryptoBalance, sumFiatValue, sumFiatBalance float64
				for _, cv := range cryptoValues {
					sumCryptoBalance += cv.CryptoBalance
					sumFiatValue += cv.FiatValue
					sumFiatBalance += cv.FiatBalance
				}
				avgCryptoBalance := sumCryptoBalance / float64(len(cryptoValues))
				avgFiatValue := sumFiatValue / float64(len(cryptoValues))
				avgFiatBalance := sumFiatBalance / float64(len(cryptoValues))

				avgValues := CryptoValues{
					CryptoBalance: avgCryptoBalance,
					FiatValue:     avgFiatValue,
					FiatBalance:   avgFiatBalance,
				}

				averages[string(key)] = avgValues

				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return averages, err
}

func (p *Persistence) ClearOldEntries() error {
	return p.DB.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			expiration := time.Unix(int64(item.ExpiresAt()), 0)
			if time.Since(expiration) > ttl {
				err := txn.Delete(item.Key())
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}
