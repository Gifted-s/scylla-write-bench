package main

import (
	"math/rand"
	"testing"
)

func TestRandStringGenerator(t *testing.T) {
	// Seed the random number generator with a constant value for reproducibility
	rng := rand.New(rand.NewSource(42))
	numStrings := 1000
	stringLength := 10
	uniqueStrings := make(map[string]bool)

	for i := 0; i < numStrings; i++ {
		randomString := randString(rng, stringLength)

		if len(randomString) != stringLength {
			t.Errorf("Incorrect length of the returned string. Expected: %d, Got: %d", stringLength, len(randomString))
		}
		if uniqueStrings[randomString] {
			t.Errorf("Duplicate string generated: %s", randomString)
		}
		uniqueStrings[randomString] = true
	}
}

func TestGenerateRandomData(t *testing.T) {
	// Seed the random number generator with a constant value for reproducibility
	rng := rand.New(rand.NewSource(42))

	keyLength := 10
	valueLength := 10

	numIterations := 1000
	uniqueKeys := make(map[string]bool)
	uniqueValues := make(map[string]bool)

	for i := 0; i < numIterations; i++ {
		data := generateRandomData(rng, keyLength, valueLength)

		// Check if the generated key is unique
		if uniqueKeys[data.key] {
			t.Errorf("Duplicate key generated: %s", data.key)
		} else {
			uniqueKeys[data.key] = true
		}

		// Check if the generated value is unique
		if uniqueValues[data.value] {
			t.Errorf("Duplicate value generated: %s", data.value)
		} else {
			uniqueValues[data.value] = true
		}

		// Check if the key and value are not the same
		if data.key == data.value {
			t.Errorf("Key and value are the same: %s", data.key)
		}
	}
}
