package combolist

import "context"

type Combolist struct {
	ctx             context.Context
	credentialsList []*Credentials
	threadCount     int
}

func newCombolist(ctx context.Context, threadCount int, path string) *Combolist {
	return &Combolist{
		ctx:             ctx,
		threadCount:     threadCount,
		credentialsList: NewCredentialsList(path),
	}
}

func Start(ctx context.Context, threadCount int, path string) <-chan *Credentials {
	cl := newCombolist(ctx, threadCount, path)
	return cl.start()
}

func (cl *Combolist) start() <-chan *Credentials {
	stream := make(chan *Credentials, cl.threadCount)
	go func() {
		defer close(stream)
		for _, credentials := range cl.credentialsList {
			select {
			case <-cl.ctx.Done():
				return
			case stream <- credentials:
			}
		}
	}()
	return stream
}
