package lgn

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/zew/go-questionnaire/cfg"
)

var tokenSaltNotWorking = GeneratePassword(22) // not interoperational between multiple instances of go-questionnaire, transferrer, generator

// set timezone to a constant - this is important for client-server talks, e.g. appengine frankfurt runs on different zone
var fixedLocation = time.FixedZone("UTC_-2", -2*60*60)

// tok rounds time to hours
// and computes a hash from it
func tok(hoursOffset int) string {
	hasher := md5.New()
	io.WriteString(hasher, lgns.Salt)
	t := time.Now().In(fixedLocation)
	if hoursOffset != 0 {
		t = t.Add(time.Duration(hoursOffset) * time.Hour)
	}
	// log.Printf("token time: %v", t.Format("02.01.2006 15"))
	io.WriteString(hasher, t.Format("02.01.2006 15"))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}

// FormToken returns a form token.
// User independent.
// Should we add the user name into the hashed base?
func FormToken() string {
	return tok(0)
}

// ValidateFormToken checks tokens
// against current hour - back to n previous hours.
// Plus one more for bounding glitches / border crossing
// when the rounding jumps from 12:59 to 13:00.
// i.e.
// FormTimeout := 2
// lower bound := -4
// => Checking token against current hour, previous hour, second previous hour, third previous hour
func ValidateFormToken(arg string) error {
	lowerBound := cfg.Get().FormTimeout*-1 - 1
	for i := 0; i >= lowerBound; i-- {
		if arg == tok(i) {
			return nil
		}
	}
	if arg == tok(1) {
		return nil
	}
	return fmt.Errorf("form token was not issued within the last two hours. \nPlease re-login")
}
