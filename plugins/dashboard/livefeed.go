package dashboard

import (
	"time"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"

	"github.com/gohornet/hornet/pkg/model/milestone"
	tanglemodel "github.com/gohornet/hornet/pkg/model/tangle"
	"github.com/gohornet/hornet/pkg/shutdown"
	"github.com/gohornet/hornet/plugins/tangle"
)

func runLiveFeed() {

	newTxZeroValueRateLimiter := time.NewTicker(time.Second / 10)
	newTxValueRateLimiter := time.NewTicker(time.Second / 20)

	onReceivedNewTransaction := events.NewClosure(func(cachedTx *tanglemodel.CachedMessage, latestMilestoneIndex milestone.Index, latestSolidMilestoneIndex milestone.Index) {
		cachedTx.ConsumeMessage(func(msg *tangle.Message) {
			if !tanglemodel.IsNodeSyncedWithThreshold() {
				return
			}

			if tx.Tx.Value == 0 {
				select {
				case <-newTxZeroValueRateLimiter.C:
					hub.BroadcastMsg(&Msg{Type: MsgTypeTxZeroValue, Data: &LivefeedTransaction{Hash: tx.Tx.Hash, Value: tx.Tx.Value}})
				default:
				}
			} else {
				select {
				case <-newTxValueRateLimiter.C:
					hub.BroadcastMsg(&Msg{Type: MsgTypeTxValue, Data: &LivefeedTransaction{Hash: tx.Tx.Hash, Value: tx.Tx.Value}})
				default:
				}
			}
		})
	})

	onLatestMilestoneIndexChanged := events.NewClosure(func(msIndex milestone.Index) {
		if msTailTxHash := getMilestoneTailHash(msIndex); msTailTxHash != nil {
			hub.BroadcastMsg(&Msg{Type: MsgTypeMs, Data: &LivefeedMilestone{Hash: msTailTxHash.Hex(), Index: msIndex}})
		}
	})

	daemon.BackgroundWorker("Dashboard[TxUpdater]", func(shutdownSignal <-chan struct{}) {
		tangle.Events.ReceivedNewMessage.Attach(onReceivedNewTransaction)
		defer tangle.Events.ReceivedNewMessage.Detach(onReceivedNewTransaction)
		tangle.Events.LatestMilestoneIndexChanged.Attach(onLatestMilestoneIndexChanged)
		defer tangle.Events.LatestMilestoneIndexChanged.Detach(onLatestMilestoneIndexChanged)

		<-shutdownSignal

		log.Info("Stopping Dashboard[TxUpdater] ...")
		newTxZeroValueRateLimiter.Stop()
		newTxValueRateLimiter.Stop()
		log.Info("Stopping Dashboard[TxUpdater] ... done")
	}, shutdown.PriorityDashboard)
}