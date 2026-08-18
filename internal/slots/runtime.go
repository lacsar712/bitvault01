package slots

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrSlot = errors.New("slot denied")

func GrowSlotIDs(dst []byte, extra byte) []byte {
	out := make([]byte, len(dst)+1)
	copy(out, dst)
	out[len(dst)] = extra
	return out
}

type SlotBag struct {
	counters map[string]int
}

func NewSlotBag() *SlotBag {
	bag := &SlotBag{}
	bag.counters = make(map[string]int)
	return bag
}

func (b *SlotBag) Set(slot string, counter int) {
	b.counters[slot] = counter
}

func (b *SlotBag) Get(slot string) int {
	return b.counters[slot]
}

func ExportImageFile(root, rel string) (string, error) {
	if strings.TrimSpace(rel) == "" {
		return "", errors.New("empty image path")
	}
	if filepath.IsAbs(rel) {
		return "", errors.New("absolute image path")
	}
	clean := filepath.Clean(rel)
	full := filepath.Join(root, clean)
	relOut, err := filepath.Rel(filepath.Clean(root), full)
	if err != nil {
		return "", err
	}
	if relOut == ".." || strings.HasPrefix(relOut, ".."+string(filepath.Separator)) {
		return "", errors.New("image path escapes root")
	}
	return full, nil
}

func WrapSlotDenied(op, slot string) error {
	if strings.TrimSpace(op) == "" {
		op = "commit"
	}
	if strings.TrimSpace(slot) == "" {
		slot = "unknown"
	}
	return fmt.Errorf("%s: slot %s: %w", op, slot, ErrSlot)
}

func CopyFirmwareHead(blob []byte, n int) []byte {
	if n < 0 {
		n = 0
	}
	if n > len(blob) {
		n = len(blob)
	}
	out := make([]byte, n)
	copy(out, blob[:n])
	return out
}

func WaitCommit(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func DumpCampaign(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	if _, err := w.WriteString(body); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

type Image struct {
	Series string
	Slot   string
}

func (img *Image) SeriesName() string {
	if img == nil {
		return ""
	}
	return img.Series
}

func ParseManifest(b []byte) (map[string]int, error) {
	var m map[string]int
	if len(b) == 0 {
		return nil, errors.New("empty manifest")
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}
