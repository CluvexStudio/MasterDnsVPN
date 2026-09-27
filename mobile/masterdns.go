package masterdns

import (
	"context"
	"errors"
	"sync"

	"masterdnsvpn-go/internal/client"
	"masterdnsvpn-go/internal/config"
)

var (
	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
)

func Start(configPath string, resolversPath string, logPath string) error {
	mu.Lock()
	defer mu.Unlock()
	if cancel != nil {
		return nil
	}
	if configPath == "" {
		return errors.New("masterdns: empty config path")
	}

	overrides := config.ClientConfigOverrides{}
	if resolversPath != "" {
		rp := resolversPath
		overrides.ResolversFilePath = &rp
	}

	app, err := client.Bootstrap(configPath, logPath, overrides)
	if err != nil {
		return err
	}

	ctx, c := context.WithCancel(context.Background())
	cancel = c
	done = make(chan struct{})
	go func(d chan struct{}) {
		defer close(d)
		_ = app.Run(ctx)
	}(done)
	return nil
}

func Stop() {
	mu.Lock()
	c := cancel
	d := done
	cancel = nil
	done = nil
	mu.Unlock()

	if c != nil {
		c()
	}
	if d != nil {
		<-d
	}
}

func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return cancel != nil
}
