Requirements:
1. Carrier deposits a package by specifying size (small, medium, large)
   - System assigns an available compartment of matching size
   - Opens compartment and returns access token, or error if no space
2. Upon successful deposit, an access token is generated and returned
   - One access token per package
3. User retrieves package by entering access token
   - System validates code and opens compartment
   - Throws specific error if code is invalid or expired
4. Access tokens expire after 7 days
   - Expired codes are rejected if used for pickup
   - Package remains in compartment until staff removes it
5. Staff can open all expired compartments to manually handle packages
   - System opens all compartments with expired tokens
   - Staff physically removes packages and returns them to sender
6. Invalid access tokens are rejected with clear error messages
   - Wrong code, already used, or expired - user gets specific feedback

Out of scope:
- How the package gets to the locker (delivery logistics)
- How the access token reaches the customer (SMS/email notification)
- Lockout after failed access token attempts
- UI/rendering layer
- Multiple locker stations
- Payment or pricing

Entities:
- Locker: It is the orchestrator. Locker contains Compartments and AccessTokens to open them. Delivery person comes
          enters size and other details, if a compartment with given size is free, it opens, package is placed, and an
          AccessToken is generated, while the user gets a code for the token. Locker also allows user to enter code, 
          validate(expired, exists, not empty) code and open the compartment
- Compartments: It is where a package is placed. It will have size and status (occupied/free). Compartment can be opened
                or closed
- AccessTokens: Tokens that map 1-1 with a compartment, allows users to enter their code and open the right compartment.

Class Design:

class Locker:
    - compartments: []Compartment
    - accessTokenMap: Map<String,AccessToken>
    
    + depositPackage(size): -> string | error
    + pickupPackage(code): -> null | error
    + openExpiredCompartments(): -> null

class Compartment:
    - size: SIZE_ENUM(SMALL | MEDIUM | LARGE)
    - status: STATUS_ENUM (FREE | OCCUPIED)

    + getSize(): SIZE_ENUM
    + getStatus(): STATUS_ENUM
    + setStatus(STATUS_ENUM):
    + open(): void 
    + close(): void

class AccessToken:
    - code: string
    - expiration: timestamp
    - compartment: Compartment

    + getCompartment(): -> Compartment
    + getCode(): -> string
    + isExpired(): boolean