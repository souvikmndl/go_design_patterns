package amazonlocker

import (
	"errors"
	"time"
)

/*
class Locker:
    - compartments: []Compartment
    - accessTokenMap: Map<String,AccessToken>

    + depositPackage(size): -> null | error
    + pickupPackage(code): -> null | error
    + openExpiredCompartments(): -> null

class Compartment:
    - size: SIZE_ENUM(SMALL | MEDIUM | LARGE)
    - status: STATUS_ENUM (FREE | OCCUPIED)

    + getSize(): SIZE_ENUM
    + open(): void
    + markOccupied(): -> void
    + markFree(): -> void
    + isOccupied(): -> bool

class AccessToken:
    - code: string
    - expiration: timestamp
    - compartment: Compartment

    + getCompartment(): -> Compartment
    + getCode(): -> string
    + isExpired(): boolean
*/

type Locker struct {
	compartments   []Compartment
	accessTokenMap map[string]AccessToken
}

func NewLocker() *Locker {
	return &Locker{
		compartments:   []Compartment{},
		accessTokenMap: map[string]AccessToken{},
	}
}

func (l *Locker) DepositPackage(size Size) (string, error) {
	// Get free compartment with given size
	// Open compartment
	// Update status
	// Generate code and token
	// add to token map
	// return code
	compartment := l.getFreeCompartment(size)
	if compartment == nil {
		return "", errors.New("no compartment of given size is unoccupied")
	}

	compartment.Open()
	// physically place package and close the door
	compartment.MarkOccupied()
	code := "new code"
	accessToken := NewAccessToken(code, time.Now().Add(24*time.Hour), *compartment)
	l.accessTokenMap[code] = *accessToken

	return accessToken.GetCode(), nil
}

func (l *Locker) PickupPackage(code string) error {
	/*
	   - check if code is not empty
	   - validate token (exists, expired)
	   - get compartment
	   - open compartment
	   - mark compartment as free
	   - remove access token from map
	*/
	if code == "" {
		return errors.New("code is empty")
	}

	token, exists := l.accessTokenMap[code]
	if !exists {
		return errors.New("invalid code")
	}

	if token.IsExpired() {
		return errors.New("token has expired")
	}

	compartment := token.GetCompartment()
	compartment.Open()
	compartment.MarkFree()
	delete(l.accessTokenMap, code)
	return nil
}

func (l *Locker) getFreeCompartment(size Size) *Compartment {
	for _, cmt := range l.compartments {
		if cmt.GetSize() == size && !cmt.IsOccupied() {
			return &cmt
		}
	}
	return nil
}

func (l *Locker) OpenExpiredCompartments() {

}

type Size string
type Status string

const (
	Small  Size = "SMALL"
	Medium Size = "MEDIUM"
	Large  Size = "LARGE"

	Free     Status = "FREE"
	Occupied Status = "OCCUPIED"
)

type Compartment struct {
	size   Size
	status Status
}

func NewCompartment(size Size, status Status) *Compartment {
	return &Compartment{
		size:   size,
		status: status,
	}
}

func (c *Compartment) GetSize() Size {
	return c.size
}

func (c *Compartment) Open() {}

func (c *Compartment) MarkOccupied() {
	c.status = Occupied
}

func (c *Compartment) IsOccupied() bool {
	return c.status == Occupied
}

func (c *Compartment) MarkFree() {
	c.status = Free
}

type AccessToken struct {
	code        string
	expiration  time.Time
	compartment Compartment
}

func NewAccessToken(code string, expiration time.Time, compartment Compartment) *AccessToken {
	return &AccessToken{
		code:        code,
		expiration:  expiration,
		compartment: compartment,
	}
}

func (at *AccessToken) GetCompartment() *Compartment {
	return &at.compartment
}

func (at *AccessToken) GetCode() string {
	return at.code
}

func (at *AccessToken) IsExpired() bool {
	return time.Since(at.expiration) > 24*time.Hour
}
