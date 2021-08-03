package request

type retrier struct {
	maxRetryCount int
}

func newRetrier(maxRetryCount int) retrier {
	return retrier{maxRetryCount}
}

type retryFn func(attempt int) (err error)

func (r retrier) retry(fn retryFn) error {
	var err error
	attempt := 0
	for {
		if r.maxRetryCount == attempt {
			break
		}
		err = fn(attempt)
		if err == nil {
			break
		}
		attempt++
	}
	return err
}
