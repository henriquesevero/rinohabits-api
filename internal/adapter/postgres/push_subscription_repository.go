package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henriquesevero/rinohabits-api/internal/domain/notification"
)

type PushSubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewPushSubscriptionRepository(pool *pgxpool.Pool) PushSubscriptionRepository {
	return PushSubscriptionRepository{pool: pool}
}

func (r PushSubscriptionRepository) Save(ctx context.Context, sub *notification.PushSubscription) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO push_subscriptions (id, user_id, endpoint, p256dh, auth_key, reminder_hour)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (user_id, endpoint) DO UPDATE
		 SET p256dh = EXCLUDED.p256dh, auth_key = EXCLUDED.auth_key, reminder_hour = EXCLUDED.reminder_hour`,
		sub.ID, sub.UserID, sub.Endpoint, sub.P256DH, sub.Auth, sub.ReminderHour,
	)
	return err
}

func (r PushSubscriptionRepository) DeleteByEndpoint(ctx context.Context, userID uuid.UUID, endpoint string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM push_subscriptions WHERE user_id = $1 AND endpoint = $2`,
		userID, endpoint,
	)
	return err
}

// ReminderTargetsForHour returns subscriptions where reminder_hour matches
// and the user has at least one incomplete habit scheduled for today.
func (r PushSubscriptionRepository) ReminderTargetsForHour(ctx context.Context, hour int) ([]*notification.ReminderTarget, error) {
	// EXTRACT(ISODOW FROM CURRENT_DATE): 1=Mon … 7=Sun — matches active_weekdays format
	rows, err := r.pool.Query(ctx,
		`SELECT ps.endpoint, ps.p256dh, ps.auth_key, COUNT(h.id) AS incomplete
		 FROM push_subscriptions ps
		 JOIN habits h
		   ON h.user_id = ps.user_id
		  AND h.is_active
		  AND h.deleted_at IS NULL
		  AND EXTRACT(ISODOW FROM CURRENT_DATE)::int = ANY(h.active_weekdays)
		 WHERE ps.reminder_hour = $1
		   AND NOT EXISTS (
		       SELECT 1 FROM daily_logs dl
		       WHERE dl.habit_id = h.id
		         AND dl.log_date = CURRENT_DATE
		   )
		 GROUP BY ps.endpoint, ps.p256dh, ps.auth_key
		 HAVING COUNT(h.id) > 0`,
		hour,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []*notification.ReminderTarget
	for rows.Next() {
		t := &notification.ReminderTarget{}
		if err := rows.Scan(&t.Endpoint, &t.P256DH, &t.Auth, &t.Incomplete); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, rows.Err()
}
