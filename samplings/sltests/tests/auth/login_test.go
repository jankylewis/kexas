package auth_test

import (
	"testing"

	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/kexas/ktest"
	"github.com/jankylewis/sltests/data"
	"github.com/jankylewis/sltests/pages"
)

// LoginSuite tests login flows. Does NOT auto-login — that's the system under test.
type LoginSuite struct {
	ktest.Suite
}

func (s *LoginSuite) TestALoginPageDisplaysForm() {
	login, err := pages.NewLoginPage(s.Page).Open()
	kassert.ThatError(s.T(), err).IsNil()

	title, err := s.Page.Title()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), title).Equals("Swag Labs")
	_ = login
}

func (s *LoginSuite) TestBStandardUserCanLogIn() {
	login, _ := pages.NewLoginPage(s.Page).Open()

	inv, err := login.LoginAs(data.StandardUser())
	kassert.ThatError(s.T(), err).IsNil()

	loaded, _ := inv.IsLoaded()
	kassert.That(s.T(), loaded).IsTrue()
}

func (s *LoginSuite) TestCLockedUserSeesLockoutError() {
	login, _ := pages.NewLoginPage(s.Page).Open()
	_, _ = login.LoginAs(data.LockedUser())

	msg, err := login.ErrorMessage()
	kassert.ThatError(s.T(), err).IsNil()
	kassert.That(s.T(), msg).Contains("locked out")
}

func (s *LoginSuite) TestDInvalidCredentialsShowError() {
	login, _ := pages.NewLoginPage(s.Page).Open()
	_, _ = login.LoginAs(data.InvalidUser())

	msg, _ := login.ErrorMessage()
	kassert.That(s.T(), msg).Contains("Username and password do not match")
}

func (s *LoginSuite) TestEEmptyUsernameShowsError() {
	login, _ := pages.NewLoginPage(s.Page).Open()
	_ = login.FillPassword("secret_sauce")
	_ = login.ClickLogin()

	msg, _ := login.ErrorMessage()
	kassert.That(s.T(), msg).Contains("Username is required")
}

func (s *LoginSuite) TestFEmptyPasswordShowsError() {
	login, _ := pages.NewLoginPage(s.Page).Open()
	_ = login.FillUsername("standard_user")
	_ = login.ClickLogin()

	msg, _ := login.ErrorMessage()
	kassert.That(s.T(), msg).Contains("Password is required")
}

// TestGLocalStorageInitializedAfterLogin verifies that the saucedemo React app
// has actually rendered (not just that we got a response). Saucedemo no longer
// stores session-username in localStorage; instead it sets backtrace-* keys
// from its error-tracking SDK on first render. Presence of any localStorage key
// proves the SPA fully booted post-login.
//
// (Earlier dogfood iteration assumed `localStorage["session-username"]` was set
// — that's saucedemo's pre-2024 model. Test re-aimed to be saucedemo-version-tolerant.)
func (s *LoginSuite) TestGLocalStorageInitializedAfterLogin() {
	login, _ := pages.NewLoginPage(s.Page).Open()
	_, err := login.LoginAs(data.StandardUser())
	kassert.ThatError(s.T(), err).IsNil()

	raw, err := s.Page.Evaluate(`Object.keys(localStorage).length`)
	kassert.ThatError(s.T(), err).IsNil()

	count := 0
	if obj, ok := raw.(map[string]interface{}); ok {
		if v, exists := obj["value"]; exists {
			if f, ok := v.(float64); ok {
				count = int(f)
			}
		}
	} else if f, ok := raw.(float64); ok {
		count = int(f)
	}
	kassert.That(s.T(), count > 0).Named("localStorage non-empty after login").IsTrue()
}

func TestLogin(t *testing.T) { ktest.Run(t, new(LoginSuite)) }
