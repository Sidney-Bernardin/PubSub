package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/Sidney-Bernardin/PubSub/server"
)

func main() {

	// Create a logger.
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	logger := zerolog.New(os.Stdout)
	logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Get the address flag.
	addr := flag.String("addr", "0.0.0.0:8080", "An address for the Pub/Sub server to listen on.")
	flag.Parse()

	// Create a server.
	svr := server.NewServer(&logger)

	// Get interrupt/termination signals to go through a new channel.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	logger.Info().Msgf("Listening on %s", *addr)

	go func() {

		// Start the server.
		if err := svr.Start(*addr); err != nil {
			logger.Error().Stack().Err(err).Msg("Server unexpectedly stopped")
			return
		}
	}()

	// Wait for interrupt/termination signals.
	sig := <-signalChan

	logger.Info().
		Str("signal", fmt.Sprintf("%s", sig)).
		Msg("Goodbye, World!")
}
