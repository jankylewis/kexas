
package kassert_test

import (
	"testing"

	"github.com/jankylewis/kexas/kassert"
)

// mockT is a mock testing.T for testing assertions.
type mockT struct {
	failed  bool
	message string
}

func (m *mockT) Helper()                                   {}
func (m *mockT) Error(args ...interface{})                 { m.failed = true }
func (m *mockT) Errorf(format string, args ...interface{}) { m.failed = true; m.message = format }
func (m *mockT) Fatal(args ...interface{})                 { m.failed = true }
func (m *mockT) Fatalf(format string, args ...interface{}) { m.failed = true }

func TestThat_Equals_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 42).Equals(42)
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_Equals_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 42).Equals(99)
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_NotEquals_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 42).NotEquals(99)
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_NotEquals_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 42).NotEquals(42)
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_IsNil_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	var nilPtr *int = nil
	kassert.That(mock, nilPtr).IsNil()
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_IsNil_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	var value int = 42
	kassert.That(mock, &value).IsNil()
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_IsNotNil_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	var value int = 42
	kassert.That(mock, &value).IsNotNil()
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_IsNotNil_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	var nilPtr *int = nil
	kassert.That(mock, nilPtr).IsNotNil()
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_IsTrue_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, true).IsTrue()
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_IsTrue_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, false).IsTrue()
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_IsFalse_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, false).IsFalse()
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_IsFalse_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, true).IsFalse()
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_Contains_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello world").Contains("world")
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_Contains_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello world").Contains("xyz")
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_NotContains_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello world").NotContains("xyz")
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_NotContains_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello world").NotContains("world")
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_StartsWith_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello world").StartsWith("hello")
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_StartsWith_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello world").StartsWith("world")
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_EndsWith_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello world").EndsWith("world")
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_EndsWith_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello world").EndsWith("hello")
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}
