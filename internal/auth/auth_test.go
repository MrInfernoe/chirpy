package auth_test

import (
	"testing"

	"github.com/MrInfernoe/Chirpy/internal/auth"
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

// 	out, err := auth.HashPassword(test.inword)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	if out != test.expected {
// 		fmt.Printf("%v not equal to %v", out, test.expected)
// 	}
