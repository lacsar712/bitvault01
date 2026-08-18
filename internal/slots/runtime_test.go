package slots

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAfterWriteRejectsCounterRollback(t *testing.T) {
	min := ""
	get := func() (string, error) { return min, nil }
	set := func(v string) error { min = v; return nil }
	if err := AfterWrite(get, set, "slot=a counter=12"); err != nil {
		t.Fatal(err)
	}
	if err := AfterWrite(get, set, "slot=b counter=4"); err == nil {
		t.Fatal("expected anti-rollback to reject a lower counter")
	}
}

func TestGrowSlotIDsNoWriteThrough(t *testing.T) {
	dst := make([]byte, 2, 8)
	copy(dst, []byte("AB"))
	got := GrowSlotIDs(dst, 'C')
	got[0] = 'X'
	if dst[0] != 'A' {
		t.Fatal("GrowSlotIDs wrote through into the slot id buffer")
	}
}

func TestSlotBagSetGet(t *testing.T) {
	bag := NewSlotBag()
	bag.Set("a", 12)
	if bag.Get("a") != 12 {
		t.Fatal("counter not stored")
	}
}

func TestExportImageFileRejectsEscape(t *testing.T) {
	if _, err := ExportImageFile(t.TempDir(), filepath.Join("..", "boot")); err == nil {
		t.Fatal("expected path escape to be rejected")
	}
}

func TestWrapSlotDeniedIs(t *testing.T) {
	err := WrapSlotDenied("commit", "a")
	if !errors.Is(err, ErrSlot) {
		t.Fatalf("lost ErrSlot: %v", err)
	}
}

func TestCopyFirmwareHeadIndependent(t *testing.T) {
	src := []byte{0x7f, 'E', 'L', 'F', 1, 2, 3, 4}
	got := CopyFirmwareHead(src, 4)
	got[0] = 0
	if src[0] != 0x7f {
		t.Fatal("CopyFirmwareHead aliased the firmware blob")
	}
}

func TestWaitCommitHonorsCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	err := WaitCommit(ctx, 600*time.Millisecond)
	if err == nil {
		t.Fatal("expected cancel error")
	}
	if time.Since(start) > 250*time.Millisecond {
		t.Fatalf("WaitCommit ignored cancel, elapsed=%s", time.Since(start))
	}
}

func TestDumpCampaignPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "campaign.txt")
	body := "slot=a counter=12\n"
	if err := DumpCampaign(path, body); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != body {
		t.Fatalf("got %q", b)
	}
}

func TestNilImageSeriesName(t *testing.T) {
	var img *Image
	if img.SeriesName() != "" {
		t.Fatalf("got %q", img.SeriesName())
	}
}

func TestParseManifestRejectsInvalid(t *testing.T) {
	if _, err := ParseManifest([]byte("counter=12")); err == nil {
		t.Fatal("expected JSON error")
	}
}
