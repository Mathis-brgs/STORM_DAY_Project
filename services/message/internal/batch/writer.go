package batch

import (
	"errors"
	"time"

	"github.com/Mathis-brgs/storm-project/services/message/internal/metrics"
	"github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
)

const (
	flushInterval = 50 * time.Millisecond
	maxBatchSize  = 500
	chanBuffer    = 10_000
)

type request struct {
	msg   *models.ChatMessage
	respC chan response
}

type response struct {
	saved *models.ChatMessage
	err   error
}

// Writer accumule les messages et les insère en batch.
type Writer struct {
	in      chan request
	repo    repo.MessageRepo
	metrics *metrics.Metrics
}

// New crée un BatchWriter et démarre le goroutine de flush.
func New(r repo.MessageRepo, m *metrics.Metrics) *Writer {
	w := &Writer{
		in:      make(chan request, chanBuffer),
		repo:    r,
		metrics: m,
	}
	go w.run()
	return w
}

// Submit soumet un message au batch et bloque jusqu'à ce qu'il soit inséré en DB.
// Thread-safe, peut être appelé depuis plusieurs goroutines simultanément.
func (w *Writer) Submit(msg *models.ChatMessage) (*models.ChatMessage, error) {
	w.metrics.IncReceived()
	ch := make(chan response, 1)
	w.in <- request{msg: msg, respC: ch}
	res := <-ch
	return res.saved, res.err
}

func (w *Writer) run() {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	buf := make([]request, 0, maxBatchSize)

	flush := func() {
		if len(buf) == 0 {
			return
		}

		msgs := make([]*models.ChatMessage, len(buf))
		for i, r := range buf {
			msgs[i] = r.msg
		}

		start := time.Now()
		saved, err := w.repo.BulkSaveMessages(msgs)
		durationUs := time.Since(start).Microseconds()

		w.metrics.ObserveBatch(len(buf), durationUs)

		if err != nil {
			w.metrics.IncInsertError()
			for _, r := range buf {
				r.respC <- response{err: err}
			}
		} else {
			if len(saved) != len(buf) {
				batchErr := errors.New("batch result count mismatch")
				for _, r := range buf {
					r.respC <- response{err: batchErr}
				}
			} else {
				w.metrics.AddInserted(len(saved))
				for i, r := range buf {
					r.respC <- response{saved: saved[i]}
				}
			}
		}

		buf = buf[:0]
	}

	for {
		select {
		case req := <-w.in:
			buf = append(buf, req)
			if len(buf) >= maxBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
