package push

import (
	"context"
	"encoding/json"
	"log"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/postgres"
	"github.com/henriquesevero/rinohabits-api/internal/domain/notification"
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
	if s.vapidPrivateKey == "" || s.vapidPublicKey == "" {
		log.Println("push scheduler: VAPID keys not set — scheduler disabled")
		return
	}
	log.Println("push scheduler: started")
	go func() {
		for {
			now := time.Now()
			next := now.Truncate(time.Minute).Add(time.Minute)
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Until(next)):
			}

			now = time.Now()
			s.sendReminders(ctx, now.Hour(), now.Minute())
		}
	}()
}

func (s *Scheduler) sendReminders(ctx context.Context, hour, minute int) {
	targets, err := s.repo.ReminderTargets(ctx, hour, minute)
	if err != nil {
		log.Printf("push scheduler: query error: %v", err)
		return
	}
	if len(targets) == 0 {
		return
	}
	log.Printf("push scheduler: sending to %d subscriber(s) at %02d:%02d", len(targets), hour, minute)
	for _, t := range targets {
		if err := Send(t, "RinoHabits", formatBody(t.Incomplete), s.vapidPublicKey, s.vapidPrivateKey, s.vapidEmail); err != nil {
			log.Printf("push scheduler: send error: %v", err)
		} else {
			log.Printf("push scheduler: sent OK")
		}
	}
}

// Send delivers a single push notification to one target.
func Send(t *notification.ReminderTarget, title, body, pubKey, privKey, email string) error {
	payload, _ := json.Marshal(map[string]string{"title": title, "body": body})
	resp, err := webpush.SendNotification(payload, &webpush.Subscription{
		Endpoint: t.Endpoint,
		Keys:     webpush.Keys{P256dh: t.P256DH, Auth: t.Auth},
	}, &webpush.Options{
		VAPIDPublicKey:  pubKey,
		VAPIDPrivateKey: privKey,
		Subscriber:      "mailto:" + email,
		TTL:             3600,
	})
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
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
