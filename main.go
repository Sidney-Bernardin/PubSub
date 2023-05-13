package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"pubsub/server"
)

func main() {

	// Create a logger.
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	logger := zerolog.New(os.Stdout)
	logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Get the address from an environment variable.
	addr, ok := os.LookupEnv("ADDR")
	if !ok {
		addr = "localhost:8080"
	}

	// Create a server.
	svr := server.NewServer(&logger)

	// Get interrupt/termination signals to go through a new channel.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	logger.Info().Msgf("Listening on %s...", addr)

	go func() {

		// Start the server.
		if err := svr.Start(addr); err != nil {
			logger.Error().Stack().Err(err).Msg("Server unexpectedly stopped")
			return
		}
	}()

	// Listen for interrupt/termination signals.
	sig := <-signalChan

	logger.Info().
		Str("signal", fmt.Sprintf("%s", sig)).
		Msg("Goodbye, World!")
}
