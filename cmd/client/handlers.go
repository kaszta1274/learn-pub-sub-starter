package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.AckTypeAck
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		mo := gs.HandleMove(move)
		switch mo {
		case gamelogic.MoveOutComeSafe:
			return pubsub.AckTypeAck
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+move.Player.Username,
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				return pubsub.AckTypeNackRequeue
			}
			return pubsub.AckTypeAck
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.AckTypeNackDiscard
		default:
			return pubsub.AckTypeNackDiscard
		}
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.AckTypeNackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.AckTypeNackDiscard
		case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon:
			logMessage := fmt.Sprintf("%s won a war against %s", winner, loser)
			err := publishGameLog(
				ch,
				routing.GameLog{
					CurrentTime: time.Now().UTC(),
					Message:     logMessage,
					Username:    gs.GetUsername(),
				},
			)
			if err != nil {
				return pubsub.AckTypeNackRequeue
			}
			return pubsub.AckTypeAck
		case gamelogic.WarOutcomeDraw:
			logMessage := fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			err := publishGameLog(
				ch,
				routing.GameLog{
					CurrentTime: time.Now().UTC(),
					Message:     logMessage,
					Username:    gs.GetUsername(),
				},
			)
			if err != nil {
				return pubsub.AckTypeNackRequeue
			}
			return pubsub.AckTypeAck
		default:
			fmt.Println("error: unknown war outcome")
			return pubsub.AckTypeNackDiscard
		}
	}
}

func publishGameLog(ch *amqp.Channel, gl routing.GameLog) error {
	err := pubsub.PublishGob(
		ch,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+gl.Username,
		gl,
	)
	return err
}
