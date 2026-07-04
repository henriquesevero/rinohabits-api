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
		`INSERT INTO push_subscriptions (id, user_id, endpoint, p256dh, auth_key)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_id, endpoint) DO UPDATE
		 SET p256dh = EXCLUDED.p256dh, auth_key = EXCLUDED.auth_key`,
		sub.ID, sub.UserID, sub.Endpoint, sub.P256DH, sub.Auth,
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

// ReminderTargets returns all subscriptions for users who still have incomplete habits today (BRT timezone).
func (r PushSubscriptionRepository) ReminderTargets(ctx context.Context) ([]*notification.ReminderTarget, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ps.endpoint, ps.p256dh, ps.auth_key, u.name, COUNT(h.id) AS incomplete
		 FROM push_subscriptions ps
		 JOIN users u ON u.id = ps.user_id
		 JOIN habits h ON h.user_id = ps.user_id
		   AND h.is_active
		   AND h.deleted_at IS NULL
		   AND EXTRACT(ISODOW FROM (NOW() AT TIME ZONE 'America/Sao_Paulo'))::int = ANY(h.active_weekdays)
		 WHERE NOT EXISTS (
		     SELECT 1 FROM daily_logs dl
		     WHERE dl.habit_id = h.id
		       AND dl.log_date = (NOW() AT TIME ZONE 'America/Sao_Paulo')::date
		 )
		 GROUP BY ps.endpoint, ps.p256dh, ps.auth_key, u.name
		 HAVING COUNT(h.id) > 0`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []*notification.ReminderTarget
	for rows.Next() {
		t := &notification.ReminderTarget{}
		if err := rows.Scan(&t.Endpoint, &t.P256DH, &t.Auth, &t.UserName, &t.Incomplete); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, rows.Err()
}
