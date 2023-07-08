package config

import (
	"errors"
	testify "github.com/stretchr/testify/assert"
	"testing"
)

type TestConfig struct {
	Sta    int      `toml:"started" json:"started"`
	Albums []string `toml:"albums" json:"albums"`
}

func TestJsonConfig(t *testing.T) {
	assert := testify.New(t)

	jsonBlob := `
{
	"bands": {"started": 1970, "albums": ["The J. Geils Band", "Full House", "Blow Your Face Out"]},
	"bandsxx": {"started": 1970, "albums": ["The J. Geils Band", "Full House", "Blow Your Face Out"]}
}
`

	loader, err := Json([]byte(jsonBlob))
	assert.Nil(err)
	if err != nil {
		return
	}

	err = loader.Register("bands", func(conf TestConfig) error {
		assert.Equal(conf.Sta, 1970)
		assert.Equal(conf.Albums[0], "The J. Geils Band")
		return nil
	})
	assert.Nil(err)

}

func TestTomlConfig(t *testing.T) {
	assert := testify.New(t)
	tomlBlob := `
		[bands]
		started = 1970
		albums = ["The J. Geils Band", "Full House", "Blow Your Face Out"]
		[bandsxx]
		started = 1970
		albums = ["The J. Geils Band", "Full House", "Blow Your Face Out"]
		`

	loader, err := Toml([]byte(tomlBlob))
	assert.Nil(err)
	if err != nil {
		return
	}

	err = loader.Register("bandsxx", func(conf *TestConfig) error {
		assert.Equal(conf.Sta, 1970)
		assert.Equal(conf.Albums[0], "The J. Geils Band")
		return errors.New("sdsd")
	})
	assert.Equal(err.Error(), "sdsd")
}
