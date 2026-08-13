package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerFiltraPorNivel(t *testing.T) {
	var buf bytes.Buffer
	l := New("teste", "1.0.0")
	l.out = &buf
	l.SetLevel(LevelInfo)

	l.Debug("debug nao deve aparecer")
	l.Info("info deve aparecer")

	out := buf.String()
	assert.False(t, strings.Contains(out, "debug nao deve aparecer"))
	assert.True(t, strings.Contains(out, "info deve aparecer"))
}

func TestLoggerJSONEstruturado(t *testing.T) {
	var buf bytes.Buffer
	l := New("teste", "1.0.0")
	l.out = &buf

	l.Info("algo aconteceu", map[string]any{"nota": 7})

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "info", entry["level"])
	assert.Equal(t, "teste", entry["service"])
	assert.Equal(t, "1.0.0", entry["version"])
	assert.Equal(t, "algo aconteceu", entry["message"])
	assert.NotEmpty(t, entry["time"])
	assert.Equal(t, float64(7), entry["nota"])
}

func TestLoggerWithCamposFixos(t *testing.T) {
	var buf bytes.Buffer
	l := New("teste", "1.0.0")
	l.out = &buf

	l.With(map[string]any{"origem": "FATURAMENTO"}).Info("baixa", map[string]any{"nota": 3})

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "FATURAMENTO", entry["origem"])
	assert.Equal(t, float64(3), entry["nota"])
}

func TestParseLevel(t *testing.T) {
	assert.Equal(t, LevelDebug, ParseLevel("debug"))
	assert.Equal(t, LevelInfo, ParseLevel("info"))
	assert.Equal(t, LevelWarn, ParseLevel("warn"))
	assert.Equal(t, LevelError, ParseLevel("error"))
	assert.Equal(t, LevelInfo, ParseLevel("desconhecido"))
}
