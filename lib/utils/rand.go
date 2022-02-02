package utils

// adapted from https://github.com/henrycg/prio/blob/master/utils/rand.go

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"sync"
)

type PRGKey [aes.BlockSize]byte

var prgMutex sync.Mutex
var bufPrgReader *BufPRGReader

const bufSize = 8192

// We use the AES-CTR to generate pseudo-random  numbers using a
// stream cipher. Go's native rand.Reader is extremely slow because
// it makes tons of system calls to generate a small number of
// pseudo-random bytes.
//
// We pay the overhead of using a sync.Mutex to synchronize calls
// to AES-CTR, but this is relatively cheap.
type PRGReader struct {
	Key    PRGKey
	stream cipher.Stream
}

type BufPRGReader struct {
	Key    PRGKey
	stream *bufio.Reader
}

func NewPRG(key *PRGKey) *PRGReader {
	out := new(PRGReader)
	out.Key = *key

	var err error
	var iv [aes.BlockSize]byte

	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}

	out.stream = cipher.NewCTR(block, iv[:])
	return out
}

func RandomPRGKey() *PRGKey {
	var key PRGKey
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		panic(err)
	}

	return &key
}

func RandomPRG() *PRGReader {
	return NewPRG(RandomPRGKey())
}

func (s *PRGReader) Read(p []byte) (int, error) {
	if len(p) < aes.BlockSize {
		var buf [aes.BlockSize]byte
		s.stream.XORKeyStream(buf[:], buf[:])
		copy(p[:], buf[:])
	} else {
		s.stream.XORKeyStream(p, p)
	}

	return len(p), nil
}

func NewBufPRG(prg *PRGReader) *BufPRGReader {
	out := new(BufPRGReader)
	out.Key = prg.Key
	out.stream = bufio.NewReaderSize(prg, bufSize)
	return out
}

func init() {
	bufPrgReader = NewBufPRG(RandomPRG())
}
