package dashboard

import (
	"context"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hornet/pkg/daemon"
	"github.com/iotaledger/hornet/pkg/tipselect"
	"github.com/iotaledger/hornet/plugins/urts"
)

func runTipSelMetricWorker() {

	// check if URTS plugin is enabled
	if Plugin.App.IsPluginSkipped(urts.Plugin) {
		return
	}

	onTipSelPerformed := events.NewClosure(func(metrics *tipselect.TipSelStats) {
		hub.BroadcastMsg(&Msg{Type: MsgTypeTipSelMetric, Data: metrics})
	})

	if err := Plugin.Daemon().BackgroundWorker("Dashboard[TipSelMetricUpdater]", func(ctx context.Context) {
		deps.TipSelector.Events.TipSelPerformed.Attach(onTipSelPerformed)
		<-ctx.Done()
		Plugin.LogInfo("Stopping Dashboard[TipSelMetricUpdater] ...")
		deps.TipSelector.Events.TipSelPerformed.Detach(onTipSelPerformed)
		Plugin.LogInfo("Stopping Dashboard[TipSelMetricUpdater] ... done")
	}, daemon.PriorityDashboard); err != nil {
		Plugin.LogPanicf("failed to start worker: %s", err)
	}
}
