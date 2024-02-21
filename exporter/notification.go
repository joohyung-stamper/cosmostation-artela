package exporter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmostation/cosmostation-coreum/notification"
	"github.com/cosmostation/cosmostation-coreum/types"
	mbltypes "github.com/cosmostation/mintscan-backend-library/types"
	mdschema "github.com/cosmostation/mintscan-database/schema"

	"go.uber.org/zap"

	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
)

// handlePushNotification handles our mobile wallet applications' push notification.
func (ex *Exporter) handlePushNotification(block *tmctypes.ResultBlock, txResp []*sdktypes.TxResponse) error {
	if len(txResp) <= 0 {
		return nil
	}

	for _, tx := range txResp {
		// Other than code equals to 0, it is failed transaction.
		if tx.Code != 0 {
			continue
		}

		msgs := tx.GetTx().GetMsgs()

		for _, msg := range msgs {

			// stdTx, ok := tx.Tx.(auth.StdTx)
			// if !ok {
			// 	return fmt.Errorf("unsupported tx type: %s", tx.Tx)
			// }

			switch m := msg.(type) {
			// case bank.MsgSend:
			case *banktypes.MsgSend:
				zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), tx.TxHash)

				// msgSend := m.(banktypestypes.MsgSend)

				var amount string
				var denom string

				// TODO: need to test for multiple coins in one message.
				if len(m.Amount) > 0 {
					amount = m.Amount[0].Amount.String()
					denom = m.Amount[0].Denom
				}

				payload := types.NewNotificationPayload(types.NotificationPayload{
					From:   m.FromAddress,
					To:     m.ToAddress,
					Txid:   tx.TxHash,
					Amount: amount,
					Denom:  denom,
				})

				// Push notification to both sender and recipient.
				notification := notification.NewNotification()

				fromAccountStatus := notification.VerifyAccountStatus(m.FromAddress)
				if fromAccountStatus {
					tokens, _ := ex.DB.QueryAlarmTokens(m.FromAddress)
					if len(tokens) > 0 {
						notification.Push(*payload, tokens, types.From)
					}
				}

				toAccountStatus := notification.VerifyAccountStatus(m.ToAddress)
				if toAccountStatus {
					tokens, _ := ex.DB.QueryAlarmTokens(m.ToAddress)
					if len(tokens) > 0 {
						notification.Push(*payload, tokens, types.To)
					}
				}

			case *banktypes.MsgMultiSend:
				zap.S().Infof("MsgType: %s | Hash: %s", m.Type(), tx.TxHash)

				// msgMultiSend := m.(banktypes.MsgMultiSend)

				notification := notification.NewNotification()

				// Push notifications to all accounts in inputs
				for _, input := range m.Inputs {
					var amount string
					var denom string

					if len(input.Coins) > 0 {
						amount = input.Coins[0].Amount.String()
						denom = input.Coins[0].Denom
					}

					payload := &types.NotificationPayload{
						From:   input.Address,
						Txid:   tx.TxHash,
						Amount: amount,
						Denom:  denom,
					}

					fromAccountStatus := notification.VerifyAccountStatus(input.Address)
					if fromAccountStatus {
						tokens, _ := ex.DB.QueryAlarmTokens(input.Address)
						if len(tokens) > 0 {
							notification.Push(*payload, tokens, types.From)
						}
					}
				}

				// Push notifications to all accounts in outputs
				for _, output := range m.Outputs {
					var amount string
					var denom string

					if len(output.Coins) > 0 {
						amount = output.Coins[0].Amount.String()
						denom = output.Coins[0].Denom
					}

					payload := &types.NotificationPayload{
						To:     output.Address,
						Txid:   tx.TxHash,
						Amount: amount,
						Denom:  denom,
					}

					toAcctStatus := notification.VerifyAccountStatus(output.Address)
					if toAcctStatus {
						tokens, _ := ex.DB.QueryAlarmTokens(output.Address)
						if len(tokens) > 0 {
							notification.Push(*payload, tokens, types.To)
						}
					}
				}

			default:
				continue
			}
		}
	}

	return nil
}

// SlackRequestBody is a type for sending a message to Slack.
type SlackRequestBody struct {
	Text string `json:"text"`
}

// NotificationToSlack is a function that sends a proposal notification message to Slack.
func (ex *Exporter) NotificationToSlack(msg, url string) error {

	webhookUrl := url

	slackBody, err := json.Marshal(SlackRequestBody{Text: msg})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, webhookUrl, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return errors.New("Non-ok response returned from Slack")
	}
	return nil
}

// SetMessageForProposalOccur is a function that sets a message to notify that a proposal has occurred.
func (ex *Exporter) SetMessageForProposalOccur(proposal *mdschema.Proposal) string {
	uri := ex.Config.Web.URI

	chainID, err := ex.App.Client.RPC.GetNetworkChainID()
	if err != nil {
		url, err := url.Parse(uri)
		if err != nil {
			chainID = uri
		} else {
			chainID = strings.TrimPrefix(url.RequestURI(), "/")
		}
	}

	msg := fmt.Sprintf("[%s] 새로운 프로포절이 생성되었습니다.\n"+
		"Number : %d\n"+
		"Title : %s\n"+
		"Submit Time : %s\n"+
		"Deposit End Time : %s\n"+
		"[%s/proposals/%d]\n",
		chainID, proposal.ID, proposal.Title,
		proposal.SubmitTime, proposal.DepositEndTime, uri, proposal.ID)

	return msg
}

// SetMessageForVoting is a function that sets a message asking you to vote on a proposal.
func (ex *Exporter) SetMessageForVoting(proposal *mdschema.Proposal) string {
	uri := ex.Config.Web.URI

	chainID, err := ex.App.Client.RPC.GetNetworkChainID()
	if err != nil {
		url, err := url.Parse(uri)
		if err != nil {
			chainID = uri
		} else {
			chainID = strings.TrimPrefix(url.RequestURI(), "/")
		}
	}
	return fmt.Sprintf("[%s] 투표가 진행중입니다.\n"+
		"Number : %d\n"+
		"Proposal Title : %s\n"+
		"Voting End Time : %s\n"+
		"[%s/proposals/%d]\n",
		chainID, proposal.ID, proposal.Title,
		proposal.VotingEndTime, uri, proposal.ID)
}

// ProposalNotificationToSlack 함수는 슬랙에 메시지를 보내기 위한 함수로 고루틴을 사용해서 실행해야함
func (ex *Exporter) ProposalNotificationToSlack(id uint64) {

	// LCD를 통해 프로포절 정보를 받아오는 것은 DB로만 정보를 받아올 경우 투표기간 돌입 상태를 확인 못할수도 있기 때문
	proposalByLCD, err := ex.GetProposal_v1(id)
	if err != nil {
		zap.L().Error("failed get proposal info from lcd", zap.Error(err))
	}

	//TODO : 추후 버전을 위해 noti 상태 + 프로포절 상태 받아오는 함수 만들기
	proposalByDB, err := ex.DB.GetProposal(id)

	if err != nil {
		zap.L().Error("failed query proposal noti status", zap.Error(err))
	}

	//테스트용
	//proposalByLCD.ProposalStatus = proposalByDB.ProposalStatus

	propStatus := proposalByLCD.ProposalStatus
	notificationStatus := proposalByDB.NotificationStatus
	if (propStatus == mbltypes.StatusRejected || propStatus == mbltypes.StatusPassed) && (notificationStatus != mbltypes.VOTINGNOTIFIED) {
		err := ex.DB.UpdateProposalNotiStatus(id, mbltypes.VOTINGNOTIFIED)
		if err != nil {
			zap.L().Error("failed insert or update proposal noti status", zap.Error(err))
		}
	} else if propStatus == mbltypes.StatusVotingPeriod && notificationStatus != mbltypes.VOTINGNOTIFIED {
		err := ex.NotificationToSlack(ex.SetMessageForVoting(proposalByLCD), ex.Config.Slack.WebHook)
		if err != nil {
			zap.L().Error("failed proposal voting notification to slack ", zap.Error(err))
		}
		err = ex.DB.UpdateProposalNotiStatus(id, mbltypes.VOTINGNOTIFIED)
		if err != nil {
			zap.L().Error("failed insert or update proposal noti status", zap.Error(err))
		}
	} else if propStatus == mbltypes.StatusDepositPeriod && notificationStatus != mbltypes.SUBMITNOTIFIED {
		err := ex.NotificationToSlack(ex.SetMessageForProposalOccur(proposalByLCD), ex.Config.Slack.WebHook)
		if err != nil {
			zap.L().Error("failed proposal occur notification to slack ", zap.Error(err))
		}
		err = ex.DB.UpdateProposalNotiStatus(id, mbltypes.SUBMITNOTIFIED)
		if err != nil {
			zap.L().Error("failed insert or update proposal noti status", zap.Error(err))
		}
	}

}
