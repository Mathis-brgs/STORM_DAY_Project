// Package metrics expose des métriques Prometheus pour le message-service.
// Implémentation sans dépendance externe : exposition en texte Prometheus via net/http.
package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// Metrics regroupe tous les compteurs/histogrammes du message-service.
type Metrics struct {
	// Compteurs atomiques
	receivedTotal    atomic.Int64 // messages reçus via NATS
	insertedTotal    atomic.Int64 // messages insérés en DB avec succès
	insertErrorTotal atomic.Int64 // erreurs d'insertion DB

	// BatchSize : accumulation pour calculer moyenne/total
	batchCount atomic.Int64 // nombre de flush effectués
	batchSum   atomic.Int64 // somme des tailles de batch

	// Durée totale de flush (en microsecondes) pour calculer la moyenne
	flushDurationUsTotal atomic.Int64
}

// New crée et enregistre les métriques.
func New() *Metrics {
	return &Metrics{}
}

// IncReceived incrémente le compteur de messages reçus.
func (m *Metrics) IncReceived() { m.receivedTotal.Add(1) }

// AddInserted incrémente le compteur de messages insérés (n = taille du batch).
func (m *Metrics) AddInserted(n int) { m.insertedTotal.Add(int64(n)) }

// IncInsertError incrémente le compteur d'erreurs DB.
func (m *Metrics) IncInsertError() { m.insertErrorTotal.Add(1) }

// ObserveBatch enregistre une observation de flush batch.
func (m *Metrics) ObserveBatch(size int, durationUs int64) {
	m.batchCount.Add(1)
	m.batchSum.Add(int64(size))
	m.flushDurationUsTotal.Add(durationUs)
}

// Handler retourne un http.HandlerFunc qui expose les métriques en format Prometheus text.
func (m *Metrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")

		received := m.receivedTotal.Load()
		inserted := m.insertedTotal.Load()
		insertErr := m.insertErrorTotal.Load()
		batchCount := m.batchCount.Load()
		batchSum := m.batchSum.Load()
		flushUs := m.flushDurationUsTotal.Load()

		fmt.Fprintf(w, "# HELP messages_received_total Messages reçus via NATS\n")
		fmt.Fprintf(w, "# TYPE messages_received_total counter\n")
		fmt.Fprintf(w, "messages_received_total %d\n\n", received)

		fmt.Fprintf(w, "# HELP messages_inserted_total Messages insérés en DB avec succès\n")
		fmt.Fprintf(w, "# TYPE messages_inserted_total counter\n")
		fmt.Fprintf(w, "messages_inserted_total %d\n\n", inserted)

		fmt.Fprintf(w, "# HELP messages_insert_errors_total Erreurs insertion DB\n")
		fmt.Fprintf(w, "# TYPE messages_insert_errors_total counter\n")
		fmt.Fprintf(w, "messages_insert_errors_total %d\n\n", insertErr)

		fmt.Fprintf(w, "# HELP messages_batch_flush_total Nombre de flush batch effectués\n")
		fmt.Fprintf(w, "# TYPE messages_batch_flush_total counter\n")
		fmt.Fprintf(w, "messages_batch_flush_total %d\n\n", batchCount)

		var avgBatch float64
		if batchCount > 0 {
			avgBatch = float64(batchSum) / float64(batchCount)
		}
		fmt.Fprintf(w, "# HELP messages_batch_avg_size Taille moyenne des batchs\n")
		fmt.Fprintf(w, "# TYPE messages_batch_avg_size gauge\n")
		fmt.Fprintf(w, "messages_batch_avg_size %.2f\n\n", avgBatch)

		var avgFlushMs float64
		if batchCount > 0 {
			avgFlushMs = float64(flushUs) / float64(batchCount) / 1000.0
		}
		fmt.Fprintf(w, "# HELP messages_batch_flush_avg_ms Durée moyenne d'un flush batch (ms)\n")
		fmt.Fprintf(w, "# TYPE messages_batch_flush_avg_ms gauge\n")
		fmt.Fprintf(w, "messages_batch_flush_avg_ms %.3f\n", avgFlushMs)
	}
}
