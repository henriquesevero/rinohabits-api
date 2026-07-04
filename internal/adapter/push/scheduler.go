package push

import (
	"context"
	"encoding/json"
	"log"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/postgres"
)

type Scheduler struct {
	repo            postgres.PushSubscriptionRepository
	vapidPrivateKey string
	vapidPublicKey  string
	vapidEmail      string
}

func NewScheduler(pool *pgxpool.Pool, privateKey, publicKey, email string) *Scheduler {
	return &Scheduler{
		repo:            postgres.NewPushSubscriptionRepository(pool),
		vapidPrivateKey: privateKey,
		vapidPublicKey:  publicKey,
		vapidEmail:      email,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	go func() {
		for {
			now := time.Now()
			// Wake up at the top of the next minute
			next := now.Truncate(time.Minute).Add(time.Minute)
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Until(next)):
			}

			// Only act at the top of each hour
			if time.Now().Minute() != 0 {
				continue
			}

			s.sendReminders(ctx, time.Now().Hour())
		}
	}()
}

func (s *Scheduler) sendReminders(ctx context.Context, hour int) {
	if s.vapidPrivateKey == "" || s.vapidPublicKey == "" {
		return
	}

	targets, err := s.repo.ReminderTargetsForHour(ctx, hour)
	if err != nil {
		log.Printf("push scheduler: query error: %v", err)
		return
	}

	for _, t := range targets {
		payload, _ := json.Marshal(map[string]string{
			"title": "RinoHabits",
			"body":  formatBody(t.Incomplete),
		})

		resp, err := webpush.SendNotification(payload, &webpush.Subscription{
			Endpoint: t.Endpoint,
			Keys: webpush.Keys{
				P256dh: t.P256DH,
				Auth:   t.Auth,
			},
		}, &webpush.Options{
			VAPIDPublicKey:  s.vapidPublicKey,
			VAPIDPrivateKey: s.vapidPrivateKey,
			Subscriber:      "mailto:" + s.vapidEmail,
			TTL:             3600,
		})
		if err != nil {
			log.Printf("push scheduler: send error: %v", err)
			continue
		}
		resp.Body.Close()
	}
}

func formatBody(n int) string {
	if n == 1 {
		return "Você ainda tem 1 hábito para completar hoje!"
	}
	return "Você ainda tem " + itoa(n) + " hábitos para completar hoje!"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
