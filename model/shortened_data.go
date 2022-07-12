//go:generate msgp
package model

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/alexadhy/shortener/internal/hash"
)

var (
	tld        = []string{"com", "biz", "xyz", "net", "org", "dev"}
	urlFormats = []string{
		"http://www.%s/",
		"https://www.%s/",
		"http://%s/",
		"https://%s/",
		"http://www.%s/%s",
		"https://www.%s/%s",
		"http://%s/%s",
		"https://%s/%s",
		"http://%s/%s.html",
		"https://%s/%s.html",
		"http://%s/%s.php",
		"https://%s/%s.php",
	}
)

// ShortenedData is the structure that will be used to store to persistence layer
// It will be stored in the format of msgpack
type ShortenedData struct {
	Key    string    `msg:"-"`
	Orig   string    `msg:"original"`
	Hash   string    `msg:"hash"`
	Short  string    `msg:"short"`
	Expiry time.Time `msg:"expiry"`
}

const (
	defaultExpiry = 24 * 30 * time.Hour
	minExpiry     = 5 * time.Minute
)

// New takes an original URL and returns *ShortenedData and error if any
func New(orig string, ttl time.Duration) (*ShortenedData, error) {
	if ttl < minExpiry {
		ttl = defaultExpiry
	}

	s := &ShortenedData{
		Orig:   orig,
		Expiry: time.Now().UTC().Add(ttl),
	}

	sum, short := hash.Hash(orig)
	s.Short = short
	s.Key = short
	s.Hash = sum

	return s, nil
}

// GenFake creates n number of iterations for *ShortenedData
// it will return []*ShortenedData and error if any
func GenFake(n int) ([]*ShortenedData, error) {
	res := make([]*ShortenedData, n)

	for i := 0; i < n; i++ {
		sd, err := New(genFakeURL(), 10*time.Minute)
		if err != nil {
			return nil, err
		}
		res[i] = sd
	}
	return res, nil
}

func genFakeURL() string {
	fakeURLFormat := urlFormats[rand.Int()%len(urlFormats)]
	countVerbs := strings.Count(fakeURLFormat, "%s")
	randomDomain := randomDomain(7)
	if countVerbs == 1 {
		return fmt.Sprintf(fakeURLFormat, randomDomain)
	}
	randomUser := randomString(7)
	return fmt.Sprintf(fakeURLFormat, randomDomain, randomUser)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randomDomain(n int) string {
	sb := strings.Builder{}
	sb.Grow(n + 4) // 4 is . and 3 characters of tld
	sb.WriteString(randomString(n))
	sb.WriteRune('.')
	sb.WriteString(tld[rand.Intn(len(tld)-1)])
	return sb.String()
}

func randomString(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}
