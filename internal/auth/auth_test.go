package auth_test

import (
	"testing"
	"time"

	"github.com/MrInfernoe/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {

	type testCase struct {
		name     string
		inword   string
		expected string
	}

	cases := []testCase{
		{
			"password1",
			"thisIsapassword",
			"tDxTVmJCOGfNFeGf/fYeRQ$D5LlHK5SUvMBia3lcEdaN/wyZR0OZavR/txEIgPVxOM", // replace with the knwon hash
		},
		{
			"password2",
			"thisisalsoapassword",
			"L8mPiVsjFDy/BdNea/5F/w$qSUrpBKdmbefXBqw5WbhdJja9oPfCKxGsVPrJBZzoak", // also replace
		},
	}

	for _, test := range cases {
		runOk := t.Run(test.name, func(t *testing.T) {

			out, err := auth.HashPassword(test.inword)
			if err != nil {
				t.Errorf("could not hash: %v\n", err)
			}

			passes, err := auth.CheckPasswordHash(test.inword, out)
			if err != nil {
				t.Errorf("could not check hash: %v\n", err)
			}
			if !passes {
				t.Errorf("did not pass\n")
			}

		})
		if !runOk {
			t.Errorf("could not run\n")
		}
	}
}

func TestJWT(t *testing.T) {

	type makeParams struct {
		userId      uuid.UUID
		tokenSecret string
		expiresIn   time.Duration
	}

	type testCase struct {
		name                string
		make                makeParams
		validateTokenSecret string
		expected            uuid.UUID
	}

	userId := uuid.New()
	cases := []testCase{
		{
			"valid",
			makeParams{
				userId,
				"thisIsATokenSecret",
				time.Second,
			},
			"thisIsATokenSecret",
			userId,
		},
		{
			"expired token",
			makeParams{
				userId,
				"thisIsAnotherTokenSecret",
				time.Nanosecond,
			},
			"thisIsAnotherTokenSecret",
			userId,
		},
		{
			"invalid signature",
			makeParams{
				userId,
				"thisIsAlsoAnotherTokenSecret",
				time.Second,
			},
			"thisIsAnInvalidTokenSecret",
			userId,
		},
	}

	for _, test := range cases {
		runOk := t.Run(test.name, func(t *testing.T) {

			JWT, err := auth.MakeJWT(test.make.userId, test.make.tokenSecret, test.make.expiresIn)
			if err != nil {
				t.Errorf("could not make JWT: %v\n", err)
			}

			validUserId, err := auth.ValidateJWT(JWT, test.validateTokenSecret)
			if err != nil {
				if test.name == "expired token" {
					if err.Error() != "token has invalid claims: token is expired" {
						t.Errorf("unexpected error: %v\n", err)
					}
				}
				if test.name == "invalid signature" {
					if err.Error() != "token signature is invalid: signature is invalid" {
						t.Errorf("unexpected error: %v\n", err)
					}
				}
			}

			// if err != nil {
			// 	t.Errorf("could not validate token: %v\n", err)
			// }

			if validUserId != test.make.userId {
				if test.name != "expired token" && test.name != "invalid signature" {
					t.Errorf("%v not equal to %v\n", validUserId, test.make.userId)
				}
			}
		})
		if !runOk {
			t.Errorf("could not run\n")
		}
	}
}
