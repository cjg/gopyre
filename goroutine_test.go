package gopyre

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConcurrentExecSameCode(t *testing.T) {
	const workers = 20
	const iterations = 5

	errCh := make(chan error, workers*iterations)
	var wg sync.WaitGroup

	for w := range workers {
		w := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range iterations {
				fmt.Printf("Worker %d Iteration %d\n", w, i)
				x := w*iterations + i
				y := x * 2
				result, err := Exec(`x + y`, map[string]any{"x": x, "y": y})
				if err != nil {
					errCh <- err
					return
				}
				got, ok := result.(float64)
				if !ok {
					errCh <- fmt.Errorf("unexpected result type %T", result)
					return
				}
				want := float64(x + y)
				if got != want {
					errCh <- fmt.Errorf("unexpected result: got %v want %v", got, want)
					return
				}
			}
			fmt.Printf("Worked %d done\n", w)
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}
}

func TestConcurrentSameVariableMultipleValue(t *testing.T) {
	const workers = 25
	
	errCh := make(chan error, workers)
	var wg sync.WaitGroup

	for w := range workers {
		wg.Add(1)
		go func () {
			defer wg.Done()
			result, err := Exec(`
import time
time.sleep(2)
w
			`, map[string]any{"w": w})
			if err != nil {
				errCh <- err
				return
			}
			got, ok := result.(float64)
			if !ok {
				errCh <- fmt.Errorf("unexpected result type %T", result)
				return
			}
			if got != float64(w) {
				errCh <- fmt.Errorf("unexpected counter value: got %v want %0.1f", got, float64(w))
				return
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}
}

func TestConcurrentExecIsolation(t *testing.T) {
	const workers = 25

	errCh := make(chan error, workers)
	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := Exec(`
counter = globals().get("counter", 0) + 1
counter
`, nil)
			if err != nil {
				errCh <- err
				return
			}
			got, ok := result.(float64)
			if !ok {
				errCh <- fmt.Errorf("unexpected result type %T", result)
				return
			}
			if got != 1 {
				errCh <- fmt.Errorf("unexpected counter value: got %v want 1", got)
				return
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}
}
