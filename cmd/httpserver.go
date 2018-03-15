// Copyright (C) 2018 vLife Systems Ltd <http://vlifesystems.com>
// Licensed under an MIT licence.  Please see LICENSE.md for details.

package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/vlifesystems/rulehunter/logger"
	"github.com/vlifesystems/rulehunter/quitter"
)

func httpServer(
	port int,
	wwwDir string,
	q *quitter.Quitter,
	l logger.Logger,
) {
	if port <= 0 {
		return
	}
	q.Add()
	defer q.Done()
	shutdownSent := false
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	http.Handle("/", http.FileServer(http.Dir(wwwDir)))
	l.Info(fmt.Sprintf("Starting http server on port: %d", port))
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !shutdownSent {
				l.Error(fmt.Errorf("http server: %s", err))
			}
		}
	}()

	// Wait until Quit is sent on channel
	<-q.C
	shutdownSent = true
	if err := srv.Shutdown(context.Background()); err != nil {
		l.Error(fmt.Errorf("http server: %s", err))
	} else {
		l.Info(fmt.Sprintf("Shutdown http server on port: %d", port))
	}
}
