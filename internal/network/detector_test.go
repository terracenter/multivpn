package network

import (
	"testing"
)

func TestIsHostDualStack(t *testing.T) {
	// Solo verifica que la función se ejecute sin pánicos y devuelva un booleano válido
	result := IsHostDualStack()
	t.Logf("IsHostDualStack() = %v", result)
}
