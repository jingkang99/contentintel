package comm

import (
	"fmt"
	"io"
	"sort"
	"math/big"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	
	mrand "math/rand"
)

// https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
// https://flaviocopes.com/go-random/

// math/rand implements a large selection of pseudo-random number generators.
// crypto/rand implements a cryptographically secure pseudo-random number generator with a limited interface.

func init() {
	assertAvailablePRNG()
}

func assertAvailablePRNG() {
	// Assert that a cryptographically secure PRNG is available.
	buf := make([]byte, 1)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(fmt.Sprintf("crypto/rand is unavailable: Read() failed with %#v", err))
	}
}

// GenerateRandomBytes returns securely generated random bytes
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a securely generated random string
func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

// returns a URL-safe, base64 encoded securely generated random string
func GenerateRandomStringURLSafe(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}

// -----  https://programming.guide/go/crypto-rand-int.html
type CryptoSrc struct{}

func (s CryptoSrc) Seed(seed int64) {}

func (s CryptoSrc) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

func (s CryptoSrc) Uint64() (v uint64) {
	err := binary.Read(rand.Reader, binary.BigEndian, &v)
	if err != nil {
		return 0
	}
	return v
}

// ----- Fisher–Yates_Shuffle
// https://github.com/carlmjohnson/go-utils
//Type Interface is similar to a sort.Interface, but there's no Less method
type Interface interface {
	// Len is the number of elements in the collection.
	Len() int
	// Swap swaps the elements with indexes i and j.
	Swap(i, j int)
}

// Shuffle shuffles the data using the following algorithm:
//   To shuffle an array a of n elements (indices 0..n-1):
//     for i from n − 1 downto 1 do
//       j ← random integer with 0 ≤ j ≤ i
//       exchange a[j] and a[i]
func Shuffle0(s Interface) {
	for i := s.Len() - 1; i > 0; i-- {
		j := mrand.Intn(i + 1)
		s.Swap(i, j)
	}
}

// https://github.com/nobleach/go-fisher-yates-shuffle
func Shuffle1(l []int, seed int64) {
	source := mrand.NewSource(seed)
	r := mrand.New(source)

	for i := range l {
		newPosition := r.Intn(len(l) - 1)
		l[i], l[newPosition] = l[newPosition], l[i]
	}
}

func Shuffle(cnt int) []int {
	var cSeed CryptoSrc
	mrand.Seed( int64(cSeed.Uint64()) )

	a := make([]int,cnt)
    for i := 0; i < cnt; i++ {
		a[i] = i
    }

	Shuffle1(a, int64(cSeed.Uint64()))

	s := sort.IntSlice(a)
	Shuffle0(s)
	
	return s
}

func Lotto(cnt int, out int) []int {
	var cSeed CryptoSrc
	mrand.Seed( int64(cSeed.Uint64()) )

	a := make([]int, cnt)
	b := make([]int, out)

    for i := 0; i < cnt; i++ {
		a[i] = i
    }

    for i := 0; i < out; i++ {
		p := mrand.Intn( cnt - i )
		
		//fmt.Println(i, len(a), p, a[p] )
		
		b[i] = a[p] + 1
		a = append( a[0:p], a[p+1:]... )
    }

	return b
}