package gamelogic

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (gs *GameState) CommandSpam(ch *amqp.Channel, words []string) error {
	if len(words) != 2 {
		return errors.New("usage: spam <number>")
	}

	n, err := strconv.Atoi(words[1])
	if err != nil {
		return fmt.Errorf("could not conver to string: %v", err)
	}

	for range n {
		log := GetMaliciousLog()

		err := pubsub.PublishGob(
			ch,
			routing.ExchangePerilTopic,
			routing.GameLogSlug+"."+gs.GetUsername(),
			routing.GameLog{
				CurrentTime: time.Now().UTC(),
				Message:     log,
				Username:    gs.GetUsername(),
			},
		)
		if err != nil {
			return fmt.Errorf("could not publish malicious log: %v", err)
		}
	}

	return nil
}
