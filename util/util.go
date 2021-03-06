package util

import (
	"bufio"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
)

func Pad(seq []byte, n int) []byte {
	if n > len(seq) {
		size := n - len(seq)
		appendix := make([]byte, size)
		for i := 0; i < size; i++ {
			appendix[i] = byte(size)
		}
		return append(seq, appendix...)
	}
	return seq
}

func PadTo(seq []byte, size int) []byte {
	if len(seq)%size != 0 {
		return Pad(seq, len(seq)+size-len(seq)%size)
	}
	return Pad(seq, len(seq)+size)
}

var BadPad = errors.New("Bad padding")

func CheckPadding(b []byte) ([]byte, error) {
	last := int(b[len(b)-1])
	canStrip := true
	if last > len(b) || last == 0 {
		return nil, BadPad
	}
	for i := len(b) - last; i < len(b); i++ {
		if b[i] != byte(last) {
			canStrip = false
			break
		}
	}
	if canStrip {
		return b[:len(b)-last], nil
	}
	return b, BadPad
}

func ReadBase64Stdin() []byte {
	scan := bufio.NewScanner(os.Stdin)
	bytes := make([]byte, 0, 1024)

	for scan.Scan() {
		input := scan.Text()
		bs, err := base64.StdEncoding.DecodeString(input)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		bytes = append(bytes, bs...)
	}
	return bytes

}

func ReadFile(filename string) []byte {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	scan := bufio.NewScanner(file)
	bytes := make([]byte, 0, 1024)

	for scan.Scan() {
		input := []byte(scan.Text())
		bytes = append(bytes, input...)
	}
	return bytes
}

func ECBDecrypt(block cipher.Block, dst, src []byte) {
	size := block.BlockSize()
	if len(dst)%size != 0 {
		panic("size of dst and src should be multiples of blocksize")
	}
	for i := 0; i < len(dst); i += size {
		block.Decrypt(dst[i:i+size], src[i:i+size])
	}
}

func ECBEncrypt(block cipher.Block, dst, src []byte) {
	size := block.BlockSize()
	if len(dst)%size != 0 {
		panic("size of dst and src should be multiples of blocksize")
	}
	for i := 0; i < len(dst); i += size {
		block.Encrypt(dst[i:i+size], src[i:i+size])
	}
}

func CBCDecrypt(block cipher.Block, iv, dst, src []byte) {
	size := block.BlockSize()
	if len(iv) != size {
		panic("size of IV not equal to block size")
	}
	if len(dst)%size != 0 {
		panic("size of dst and src should be multiples of blocksize")
	}
	if len(dst) == 0 || len(src) == 0 {
		return
	}
	x := make([]byte, size)
	oldSrc := make([]byte, size)
	copy(x, iv)
	var i int
	for i = 0; i < len(dst) && i < len(src); i += size {
		copy(oldSrc, src[i:i+size])
		block.Decrypt(dst[i:i+size], src[i:i+size])
		for j := 0; j < size; j++ {
			dst[i+j] ^= x[j]
		}
		copy(x, oldSrc)
	}
}

func CBCEncrypt(block cipher.Block, iv, dst, src []byte) {
	size := block.BlockSize()
	if len(iv) != size {
		panic("size of IV not equal to block size")
	}
	if len(dst)%size != 0 {
		panic("size of dst and src should be multiples of blocksize")
	}
	if len(dst) == 0 || len(src) == 0 {
		return
	}
	x := make([]byte, size)
	copy(x, iv)
	for i := 0; i < len(dst) && i < len(src); i += size {
		for j := 0; j < size; j++ {
			x[j] ^= src[i+j]
		}
		block.Encrypt(dst[i:i+size], x)
		copy(x, dst[i:i+size])
	}
}

func CTREncrypt(block cipher.Block, nonce, dst, src []byte) {
	size := block.BlockSize()
	if len(nonce) != size {
		panic("size of IV not equal to block size")
	}
	if len(dst) == 0 || len(src) == 0 {
		return
	}
	// temp key
	key := make([]byte, size)
	// copy of nonce
	n := make([]byte, size)
	copy(n, nonce)
	counter := binary.LittleEndian.Uint64(n[8:])
	for i := 0; i < len(dst) && i < len(src); i += size {
		block.Encrypt(key, n)

		for j := 0; j < size && i+j < len(src); j++ {
			dst[i+j] = src[i+j] ^ key[j]
		}
		counter++
		binary.LittleEndian.PutUint64(n[8:], counter)
	}
	return
}

func RandByte() byte {
	b := make([]byte, 1)
	rand.Read(b)
	return b[0]
}

func RandBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}
	return b
}

func RandAes128() []byte {
	return RandBytes(16)
}
