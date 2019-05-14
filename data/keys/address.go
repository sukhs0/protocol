package keys

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

//Address to be used as to reference a key-pair.
type Address []byte

func (a Address) String() string {
	return fmt.Sprint("0x", strings.ToUpper(hex.EncodeToString(a)))
}

func (a Address) Bytes() []byte {
	return a
}

func (a Address) Equal(b Address) bool {
	return bytes.Equal(a, b)
}
