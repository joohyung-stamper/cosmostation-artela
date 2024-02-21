package notification

import (
	"time"

	"github.com/cosmostation/cosmostation-coreum/db"
	"github.com/cosmostation/cosmostation-coreum/types"

	mblconfig "github.com/cosmostation/mintscan-backend-library/config"
	"go.uber.org/zap"

	resty "github.com/go-resty/resty/v2"
)

// Notification wraps around configuration for push notification for mobile apps.
type Notification struct {
	cfg    *mblconfig.Config
	client *resty.Client
	db     *db.Database
}

// NewNotification returns new notification instance.
func NewNotification() *Notification {
	fileBaseName := "chain-exporter"
	config := mblconfig.ParseConfig(fileBaseName)

	client := resty.New().
		SetHostURL(config.Alarm.PushServer).
		SetTimeout(time.Duration(5 * time.Second))

	database := db.Connect(&config.DB)

	return &Notification{config, client, database}
}

// Push sends push notification to local notification server and it delivers the message to
// its respective device. Uses a push notification micro server called gorush.
// More information can be found here in this link. https://github.com/appleboy/gorush
func (nof *Notification) Push(np types.NotificationPayload, tokens []string, target string) {
	var notifications []types.Notification

	// Create new notification payload for a user sending tokens
	if target == types.From {
		platform := int8(2)
		title := types.NotificationSentTitle + np.Amount + np.Denom
		message := types.NotificationSentMessage + np.Amount + np.Denom
		data := types.NewNotificationData(np.From, np.Txid, types.Sent)
		temp := types.NewNotification(tokens, platform, title, message, data)

		notifications = append(notifications, temp)

		zap.S().Infof("sent notification - hash: %s | from: %s", np.Txid, np.From)
	}

	// Create new notification payload for a user receiving tokens
	if target == types.To {
		platform := int8(2)
		title := types.NotificationSentTitle + np.Amount + np.Denom
		message := types.NotificationSentMessage + np.Amount + np.Denom
		data := types.NewNotificationData(np.To, np.Txid, types.Received)
		temp := types.NewNotification(tokens, platform, title, message, data)

		notifications = append(notifications, temp)

		zap.S().Infof("sent notification - hash: %s | to: %s", np.Txid, np.To)
	}

	if len(notifications) > 0 {
		nsp := types.NotificationServerPayload{
			Notifications: notifications,
		}

		// Send push notification
		nof.client.R().SetBody(nsp)
	}
}

// VerifyAccountStatus verifes account status before sending notification to its local server.
func (nof *Notification) VerifyAccountStatus(address string) bool {
	acct, _ := nof.db.QueryAccountMobile(address)

	// Check account's alarm token
	if acct.AlarmToken == "" {
		return false
	}

	// Check user's alarm status
	if !acct.AlarmStatus {
		return false
	}

	return true
}
