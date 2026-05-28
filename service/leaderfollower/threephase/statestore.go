package threephase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// CoordStore[S] manages all coordination state for the three-phase commit protocol.
// Implemented by [ConfigMapCoordStore] for Kubernetes.
//
// ConfigMap keys used:
//   - "proposed"  — JSON-encoded S; set by external proposers, watched by the leader
//   - "committed" — JSON-encoded S; last successfully committed state
//   - "previous"  — JSON-encoded S; state before last commit (enables rollback after restart)
//   - "round"     — raw JSON bytes; in-flight round for crash recovery
type CoordStore[S any] interface {
	// WatchProposed blocks until a proposed state different from current is
	// available in the store. Returns when "proposed" is non-empty and encodes
	// a value that differs from current.
	WatchProposed(ctx context.Context, current S) (S, error)

	// SetProposed writes a proposed state for the leader to pick up.
	SetProposed(ctx context.Context, proposed S) error

	// LoadCommitted reads the last committed and previous states.
	// ok is false when no committed state has been stored yet.
	LoadCommitted(ctx context.Context) (committed, previous S, ok bool, err error)

	// SetCommitted persists committed and previous after a successful round,
	// and clears the "proposed" and "round" keys atomically.
	SetCommitted(ctx context.Context, committed, previous S) error

	// LoadRound reads the in-flight round bytes (raw JSON) for crash recovery.
	// ok is false when no round is in progress.
	LoadRound(ctx context.Context) (data []byte, ok bool, err error)

	// SaveRound persists the current in-flight round as raw JSON bytes.
	SaveRound(ctx context.Context, data []byte) error

	// ClearRound removes the in-flight round after completion.
	ClearRound(ctx context.Context) error
}

// --- Kubernetes ConfigMap implementation ---

// ConfigMapCoordStore implements [CoordStore][S] using a Kubernetes ConfigMap.
// All four coordination keys ("proposed", "committed", "previous", "round") are
// stored as JSON-encoded strings in the same ConfigMap.
type ConfigMapCoordStore[S any] struct {
	client    kubernetes.Interface
	namespace string
	name      string // ConfigMap name
	l         *zap.Logger
}

// NewConfigMapCoordStore creates a ConfigMapCoordStore that manages coordination
// state in data keys of the named ConfigMap.
func NewConfigMapCoordStore[S any](client kubernetes.Interface, namespace, name string, l *zap.Logger) *ConfigMapCoordStore[S] {
	return &ConfigMapCoordStore[S]{
		client:    client,
		namespace: namespace,
		name:      name,
		l:         l,
	}
}

// WatchProposed implements [CoordStore]. Uses list-then-watch so a proposed
// value set before the leader was elected is never missed.
func (s *ConfigMapCoordStore[S]) WatchProposed(ctx context.Context, current S) (S, error) {
	currentJSON, err := json.Marshal(current)
	if err != nil {
		s.l.Warn("threephase: failed to marshal current state for comparison", zap.Error(err))
	}

	for {
		if ctx.Err() != nil {
			var zero S
			return zero, ctx.Err()
		}

		cm, err := s.client.CoreV1().ConfigMaps(s.namespace).Get(ctx, s.name, metav1.GetOptions{})
		if err != nil {
			select {
			case <-ctx.Done():
				var zero S
				return zero, ctx.Err()
			case <-time.After(5 * time.Second):
			}

			continue
		}

		if proposed := cm.Data["proposed"]; proposed != "" && proposed != string(currentJSON) {
			var state S
			if err := json.Unmarshal([]byte(proposed), &state); err == nil {
				return state, nil
			} else {
				s.l.Warn("threephase: malformed proposed value in ConfigMap — waiting for next update",
					zap.String("configmap", s.name), zap.Error(err))
			}
		}

		// Watch from the resourceVersion of the object we just read so we
		// don't miss updates that arrive between the GET and the Watch.
		watcher, err := s.client.CoreV1().ConfigMaps(s.namespace).Watch(ctx, metav1.ListOptions{
			FieldSelector:   "metadata.name=" + s.name,
			ResourceVersion: cm.ResourceVersion,
		})
		if err != nil {
			select {
			case <-ctx.Done():
				var zero S
				return zero, ctx.Err()
			case <-time.After(5 * time.Second):
			}

			continue
		}

		for event := range watcher.ResultChan() {
			cm, ok := event.Object.(*corev1.ConfigMap)
			if !ok {
				continue
			}

			proposed := cm.Data["proposed"]
			if proposed == "" || proposed == string(currentJSON) {
				continue
			}

			var state S
			if err := json.Unmarshal([]byte(proposed), &state); err != nil {
				s.l.Warn("threephase: malformed proposed value in ConfigMap — waiting for next update",
					zap.String("configmap", s.name), zap.Error(err))

				continue
			}

			watcher.Stop()

			return state, nil
		}

		watcher.Stop()
	}
}

// SetProposed writes a proposed state into the ConfigMap.
func (s *ConfigMapCoordStore[S]) SetProposed(ctx context.Context, proposed S) error {
	raw, err := json.Marshal(proposed)
	if err != nil {
		return fmt.Errorf("marshal proposed: %w", err)
	}

	patch, err := json.Marshal(map[string]any{
		"data": map[string]string{"proposed": string(raw)},
	})
	if err != nil {
		return fmt.Errorf("marshal patch: %w", err)
	}

	_, err = s.client.CoreV1().ConfigMaps(s.namespace).Patch(
		ctx, s.name, types.MergePatchType, patch, metav1.PatchOptions{},
	)

	return err
}

// LoadCommitted reads the committed and previous states from the ConfigMap.
func (s *ConfigMapCoordStore[S]) LoadCommitted(ctx context.Context) (S, S, bool, error) {
	var committed, previous S

	cm, err := s.client.CoreV1().ConfigMaps(s.namespace).Get(ctx, s.name, metav1.GetOptions{})
	if err != nil {
		return committed, previous, false, fmt.Errorf("get ConfigMap %q: %w", s.name, err)
	}

	committedRaw := cm.Data["committed"]
	if committedRaw == "" {
		return committed, previous, false, nil
	}

	if err := json.Unmarshal([]byte(committedRaw), &committed); err != nil {
		return committed, previous, false, fmt.Errorf("unmarshal committed from ConfigMap %q: %w", s.name, err)
	}

	if previousRaw := cm.Data["previous"]; previousRaw != "" {
		_ = json.Unmarshal([]byte(previousRaw), &previous) // best-effort; zero on error
	}

	return committed, previous, true, nil
}

// SetCommitted persists committed and previous, and clears proposed and round.
func (s *ConfigMapCoordStore[S]) SetCommitted(ctx context.Context, committed, previous S) error {
	committedRaw, err := json.Marshal(committed)
	if err != nil {
		return fmt.Errorf("marshal committed: %w", err)
	}

	previousRaw, err := json.Marshal(previous)
	if err != nil {
		return fmt.Errorf("marshal previous: %w", err)
	}

	patch, err := json.Marshal(map[string]any{
		"data": map[string]string{
			"committed": string(committedRaw),
			"previous":  string(previousRaw),
			"proposed":  "",
			"round":     "",
		},
	})
	if err != nil {
		return fmt.Errorf("marshal patch: %w", err)
	}

	_, err = s.client.CoreV1().ConfigMaps(s.namespace).Patch(
		ctx, s.name, types.MergePatchType, patch, metav1.PatchOptions{},
	)

	return err
}

// LoadRound reads the raw in-flight round bytes from the ConfigMap.
func (s *ConfigMapCoordStore[S]) LoadRound(ctx context.Context) ([]byte, bool, error) {
	cm, err := s.client.CoreV1().ConfigMaps(s.namespace).Get(ctx, s.name, metav1.GetOptions{})
	if err != nil {
		return nil, false, fmt.Errorf("get ConfigMap %q: %w", s.name, err)
	}

	raw := cm.Data["round"]
	if raw == "" {
		return nil, false, nil
	}

	return []byte(raw), true, nil
}

// SaveRound persists raw round bytes into the ConfigMap.
func (s *ConfigMapCoordStore[S]) SaveRound(ctx context.Context, data []byte) error {
	patch, err := json.Marshal(map[string]any{
		"data": map[string]string{"round": string(data)},
	})
	if err != nil {
		return fmt.Errorf("marshal patch: %w", err)
	}

	_, err = s.client.CoreV1().ConfigMaps(s.namespace).Patch(
		ctx, s.name, types.MergePatchType, patch, metav1.PatchOptions{},
	)

	return err
}

// ClearRound removes the in-flight round by setting the key to empty.
func (s *ConfigMapCoordStore[S]) ClearRound(ctx context.Context) error {
	patch, err := json.Marshal(map[string]any{
		"data": map[string]string{"round": ""},
	})
	if err != nil {
		return fmt.Errorf("marshal patch: %w", err)
	}

	_, err = s.client.CoreV1().ConfigMaps(s.namespace).Patch(
		ctx, s.name, types.MergePatchType, patch, metav1.PatchOptions{},
	)

	return err
}
