package data

// User represents a saucedemo test account.
type User struct {
	Username string
	Password string
}

// Standard saucedemo password — same for every account.
const sauceDemoPassword string = "secret_sauce"

// StandardUser is the happy-path account used by most tests.
func StandardUser() User { return User{Username: "standard_user", Password: sauceDemoPassword} }

// LockedUser is the account that errors out at login with "user has been locked out".
func LockedUser() User { return User{Username: "locked_out_user", Password: sauceDemoPassword} }

// ProblemUser logs in successfully but renders broken images on inventory.
func ProblemUser() User { return User{Username: "problem_user", Password: sauceDemoPassword} }

// PerformanceGlitchUser logs in successfully with intentional latency.
func PerformanceGlitchUser() User {
	return User{Username: "performance_glitch_user", Password: sauceDemoPassword}
}

// ErrorUser logs in successfully but errors on certain workflows.
func ErrorUser() User { return User{Username: "error_user", Password: sauceDemoPassword} }

// VisualUser logs in successfully but with visual rendering bugs.
func VisualUser() User { return User{Username: "visual_user", Password: sauceDemoPassword} }

// InvalidUser is a non-existent account — login should fail.
func InvalidUser() User { return User{Username: "no_such_user", Password: "wrong_password"} }
