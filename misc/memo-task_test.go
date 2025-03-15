package misc_test

import (
	"strconv"
	"testing"
	"time"

	"lyonbot.github.com/my_app/misc"
)

func TestMemoTask(t *testing.T) {
	var count int
	type TestResult struct {
		A int
	}

	task := misc.MakeMemoTask(
		1024,
		10*time.Second,
		func(i int) string {
			return strconv.Itoa(i)
		},
		func(i int) (*TestResult, error) {
			count++
			return &TestResult{
				A: i,
			}, nil
		},
	)

	for i := 0; i < 100; i++ {
		task.ExecuteWithCache(i)
	}
	for i := 0; i < 100; i++ {
		item, err := task.ExecuteWithCache(i)
		if err != nil {
			t.Fatal(err)
		}
		if item.A != i {
			t.Fatalf("item.A should be %d, got %d", i, item.A)
		}
	}

	if count != 100 {
		t.Fatal("count should be 100")
	}
}
