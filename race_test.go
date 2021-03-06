package goavro

import (
	"bytes"
	"fmt"
	"testing"
)

func TestRaceCodecConstructionDecode(t *testing.T) {

	recordSchemaJSON := `{"type": "long"}`
	codec, _ := NewCodec(recordSchemaJSON)
	comms := make(chan []byte, 1000)
	done := make(chan error, 10)

	go func() {

		for i := 0; i < 10000; i++ {

			//Completely unrelated stateful objects were causing races
			if i%100 == 0 {
				recordSchemaJSON := `{"type": "long"}`
				NewCodec(recordSchemaJSON)
			}

			bb := new(bytes.Buffer)
			if err := codec.Encode(bb, int64(i)); err != nil {
				done <- err
				return
			}

			comms <- bb.Bytes()
		}
		close(comms)
	}()

	go func() {
		i := 0
		for encoded := range comms {
			bb := bytes.NewBuffer(encoded)
			decoded, err := codec.Decode(bb)
			if err != nil {
				done <- err
				return
			}
			result := decoded.(int64)
			if result != int64(i) {
				done <- fmt.Errorf("didnt match %v %v", i, result)
				return
			}

			i++
		}

		close(done)
	}()

	err := <-done
	if err != nil {
		t.Fatal(err)
	}

}

func TestRaceCodecConstruction(t *testing.T) {

	comms := make(chan []byte, 1000)
	done := make(chan error, 10)

	go func() {
		recordSchemaJSON := `{"type": "long"}`
		codec, _ := NewCodec(recordSchemaJSON)

		for i := 0; i < 10000; i++ {

			bb := new(bytes.Buffer)
			if err := codec.Encode(bb, int64(i)); err != nil {
				done <- err
				return
			}

			comms <- bb.Bytes()
		}
		close(comms)
	}()

	go func() {
		recordSchemaJSON := `{"type": "long"}`
		codec, _ := NewCodec(recordSchemaJSON)
		i := 0
		for encoded := range comms {
			bb := bytes.NewBuffer(encoded)
			decoded, err := codec.Decode(bb)
			if err != nil {
				done <- err
				return
			}
			result := decoded.(int64)
			if result != int64(i) {
				done <- fmt.Errorf("didnt match %v %v", i, result)
				return
			}

			i++
		}

		close(done)
	}()

	err := <-done
	if err != nil {
		t.Fatal(err)
	}

}
