package push

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/postgres"
	"github.com/henriquesevero/rinohabits-api/internal/domain/notification"
)

const notificationTitle = "RinoHabits"

var sendHours = map[int]string{
	9:  "morning",
	15: "afternoon",
	21: "evening",
}

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

			brt := time.FixedZone("BRT", -3*60*60)
			now = time.Now().In(brt)
			log.Printf("push scheduler: tick %02d:%02d BRT", now.Hour(), now.Minute())

			slot, ok := sendHours[now.Hour()]
			if ok && now.Minute() == 0 {
				s.sendReminders(ctx, slot)
			}
		}
	}()
}

func (s *Scheduler) sendReminders(ctx context.Context, slot string) {
	targets, err := s.repo.ReminderTargets(ctx)
	if err != nil {
		log.Printf("push scheduler: query error: %v", err)
		return
	}
	if len(targets) == 0 {
		log.Printf("push scheduler: no incomplete habits for slot %s", slot)
		return
	}
	log.Printf("push scheduler: sending slot=%s to %d subscriber(s)", slot, len(targets))
	sent := 0
	for _, t := range targets {
		body := buildMessage(slot, firstName(t.UserName), t.Incomplete)
		if err := Send(t, notificationTitle, body, s.vapidPublicKey, s.vapidPrivateKey, s.vapidEmail); err != nil {
			log.Printf("push scheduler: send error: %v", err)
			continue
		}
		sent++
	}
	log.Printf("push scheduler: sent %d/%d for slot=%s", sent, len(targets), slot)
}

func buildMessage(slot, name string, incomplete int) string {
	switch slot {
	case "morning":
		return name + ", bom dia! Não esqueça dos seus hábitos hoje 💪"
	case "afternoon":
		return name + ", " + formatCount(incomplete) + " para completar hoje!"
	default: // evening
		return name + ", último aviso! " + formatCount(incomplete) + " — não perca sua sequência 🔥"
	}
}

func formatCount(n int) string {
	if n == 1 {
		return "ainda falta 1 hábito"
	}
	return "ainda faltam " + strconv.Itoa(n) + " hábitos"
}

func Send(t *notification.ReminderTarget, title, body, pubKey, privKey, email string) error {
	payload, _ := json.Marshal(map[string]string{"title": title, "body": body})
	resp, err := webpush.SendNotification(payload, &webpush.Subscription{
		Endpoint: t.Endpoint,
		Keys:     webpush.Keys{P256dh: t.P256DH, Auth: t.Auth},
	}, &webpush.Options{
		VAPIDPublicKey:  pubKey,
		VAPIDPrivateKey: privKey,
		Subscriber:      email,
		TTL:             3600,
	})
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("push rejected: HTTP %d — %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func firstName(name string) string {
	for i, r := range name {
		if r == ' ' {
			return name[:i]
		}
	}
	return name
}
