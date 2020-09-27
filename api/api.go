package oauth

import (
	"context"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-github/github"
	"github.com/jinzhu/gorm"
	api "github.com/sm43/oauth/api/gen/api"
	log "github.com/sm43/oauth/api/gen/log"
	"goa.design/goa/v3/security"
	"golang.org/x/oauth2"
	gh "golang.org/x/oauth2/github"
)

// Defines the OAuth config which authenticates user against github
var oauth = &oauth2.Config{
	ClientID:     "e3b231e0f32952e12bfa",                     // Replace clientID with the GH Oauth clientID you created
	ClientSecret: "bf78a2cfe97196795fad206fcb7a85b7cee092b3", // Replace clientSecret with the GH Oauth clientSecret you created
	Endpoint:     gh.Endpoint,
}

const jwtSigningKey = "key" // Replace key with any randon key with which jwt will be signed

var (
	invalidCode   = api.MakeInvalidCode(fmt.Errorf("invalid authorization code"))
	internalError = api.MakeInternalError(fmt.Errorf("failed to authenticate"))
	tokenError    = api.MakeInvalidToken(fmt.Errorf("invalid user token"))
)

type apisrvc struct {
	db     *gorm.DB
	logger *log.Logger
}

type contextKey string

var (
	userIDKey = contextKey("user-id")
)

// NewAPI returns the api service implementation.
func NewAPI(db *gorm.DB, logger *log.Logger) api.Service {
	return &apisrvc{db, logger}
}

// Authenticates users against GitHub OAuth
func (s *apisrvc) Authenticate(ctx context.Context, p *api.AuthenticatePayload) (*api.AuthenticateResult, error) {

	// Using the user authorization code oauth call the GH API
	// to validates the code and if it is valid then gets the
	// access_token for user which will be used to fetch user
	// detail
	token, err := oauth.Exchange(oauth2.NoContext, p.Code)
	if err != nil {
		return nil, invalidCode
	}

	// Using the access_token of user now we fetch the user
	// details from GH
	oauthClient := oauth.Client(oauth2.NoContext, token)
	ghClient := github.NewClient(oauthClient)
	ghUser, _, err := ghClient.Users.Get(oauth2.NoContext, "")
	if err != nil {
		s.logger.Error(err)
		return nil, internalError
	}

	// Once we get user details we add them  to db
	// If user is not created already then only it will
	// create a record for the user
	userID, err := s.addUser(ghUser)
	if err != nil {
		return nil, err
	}

	// This creates user jwt which will have user id
	jwt, err := s.createJWT(userID)
	if err != nil {
		return nil, err
	}

	return &api.AuthenticateResult{Token: jwt}, nil
}

func (s *apisrvc) addUser(user *github.User) (uint, error) {

	q := s.db.Model(&User{}).Where(&User{GithubID: user.GetLogin()})

	newUser := &User{
		Name:     user.GetName(),
		GithubID: user.GetLogin(),
	}
	if err := q.FirstOrCreate(newUser).Error; err != nil {
		s.logger.Error(err)
		return 0, internalError
	}

	return newUser.ID, nil
}

func (s *apisrvc) createJWT(id uint) (string, error) {

	claim := jwt.MapClaims{
		"id": id,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	jwt, err := token.SignedString([]byte(jwtSigningKey))
	if err != nil {
		return "", internalError
	}

	return jwt, nil
}

// JWTAuth implements the authorization logic for service "api" for the "jwt"
// security scheme.
func (s *apisrvc) JWTAuth(ctx context.Context, token string, scheme *security.JWTScheme) (context.Context, error) {

	// Checks if JWT is valid
	claims := make(jwt.MapClaims)

	// Parse JWT token
	_, err := jwt.ParseWithClaims(token, claims,
		func(_ *jwt.Token) (interface{}, error) {
			return []byte(jwtSigningKey), nil
		})
	if err != nil {
		return nil, err
	}

	// Gets id from JWT
	userID, ok := claims["id"].(float64)
	if !ok {
		return ctx, tokenError
	}

	// Pass id from JWT in Context
	return WithUserID(ctx, uint(userID)), nil
}

// WithUserID adds user id in context passed to it
func WithUserID(ctx context.Context, id uint) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// Find user details
func (s *apisrvc) Details(ctx context.Context, p *api.DetailsPayload) (res *api.User, err error) {

	// Fetch the id from context
	userID := UserID(ctx)

	// Fetch the user with id from db
	var user User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, internalError
	}

	// returns user details
	return &api.User{ID: user.ID, Name: user.Name, GithubID: user.GithubID}, nil
}

// UserID fetch the user id from passed context
func UserID(ctx context.Context) uint {
	return ctx.Value(userIDKey).(uint)
}
