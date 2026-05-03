
package kassert_test

import (
	"errors"
	"testing"

	"github.com/jankylewis/kexas/kassert"
)

func TestThat_IsEmpty_String_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "").IsEmpty()
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_IsEmpty_String_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello").IsEmpty()
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_IsEmpty_Slice_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	var empty []int = []int{}
	kassert.That(mock, empty).IsEmpty()
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_IsEmpty_Slice_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	var slice []int = []int{1, 2, 3}
	kassert.That(mock, slice).IsEmpty()
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_IsNotEmpty_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello").IsNotEmpty()
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_IsNotEmpty_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "").IsNotEmpty()
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_HasLength_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello").HasLength(5)
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_HasLength_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello").HasLength(10)
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_HasLength_Slice(t *testing.T) {
	var mock *mockT = &mockT{}
	var slice []int = []int{1, 2, 3}
	kassert.That(mock, slice).HasLength(3)
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_IsGreaterThan_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 10).IsGreaterThan(5)
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_IsGreaterThan_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 5).IsGreaterThan(10)
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_IsLessThan_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 5).IsLessThan(10)
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_IsLessThan_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 10).IsLessThan(5)
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_IsGreaterThanOrEqual_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 10).IsGreaterThanOrEqual(10)
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_IsGreaterThanOrEqual_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 5).IsGreaterThanOrEqual(10)
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_IsLessThanOrEqual_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 5).IsLessThanOrEqual(5)
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_IsLessThanOrEqual_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 10).IsLessThanOrEqual(5)
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_HasType_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 42).HasType(0)
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThat_HasType_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 42).HasType("string")
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_Named(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 42).Named("myValue").Equals(99)
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
	// Check that custom name appears in error message (we'd need to capture the message)
}

func TestThat_Chaining(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello world").
		Contains("hello").
		Contains("world").
		StartsWith("hello").
		EndsWith("world").
		HasLength(11)
	if mock.failed {
		t.Error("Expected all chained assertions to pass")
	}
}

func TestThatError_IsNil_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.ThatError(mock, nil).IsNil()
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThatError_IsNil_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	var err error = errors.New("test error")
	kassert.ThatError(mock, err).IsNil()
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThatError_IsNotNil_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	var err error = errors.New("test error")
	kassert.ThatError(mock, err).IsNotNil()
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThatError_IsNotNil_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.ThatError(mock, nil).IsNotNil()
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThatError_HasMessage_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	var err error = errors.New("connection timeout")
	kassert.ThatError(mock, err).HasMessage("timeout")
	if mock.failed {
		t.Error("Expected assertion to pass")
	}
}

func TestThatError_HasMessage_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	var err error = errors.New("connection timeout")
	kassert.ThatError(mock, err).HasMessage("success")
	if !mock.failed {
		t.Error("Expected assertion to fail")
	}
}

func TestThat_NumericTypes(t *testing.T) {
	var mock *mockT = &mockT{}

	// Test int
	kassert.That(mock, int(10)).IsGreaterThan(5)
	if mock.failed {
		t.Error("Expected int comparison to work")
	}

	// Test int64
	mock = &mockT{}
	kassert.That(mock, int64(10)).IsGreaterThan(int64(5))
	if mock.failed {
		t.Error("Expected int64 comparison to work")
	}

	// Test float64
	mock = &mockT{}
	kassert.That(mock, float64(10.5)).IsGreaterThan(float64(5.5))
	if mock.failed {
		t.Error("Expected float64 comparison to work")
	}
}

// TestFail tests the kassert.Fail function.
func TestFail(t *testing.T) {
	var mock *mockT = &mockT{}

	// Test Fail with simple message
	kassert.Fail(mock, "test failure message")
	if !mock.failed {
		t.Error("Expected Fail to mark test as failed")
	}

	// Reset mock
	mock = &mockT{}

	// Test Fail with formatted message
	kassert.Fail(mock, "test failure %d", 42)
	if !mock.failed {
		t.Error("Expected Fail with formatted message to mark test as failed")
	}
}

// TestFail_EmptyMessage tests Fail with empty message.
func TestFail_EmptyMessage(t *testing.T) {
	var mock *mockT = &mockT{}

	kassert.Fail(mock, "")
	if !mock.failed {
		t.Error("Expected Fail with empty message to mark test as failed")
	}
}
