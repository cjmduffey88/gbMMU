package gbMMU

import (
	"testing"
)

func TestMMU_Boot(t *testing.T) {
	var mmu = NewMMU()
	mmu.Init("boot/dmg/dmg.bin", "roms/Tetris.gb")
	if mmu.Memory[BOOT] != 0x0 {
		t.Errorf("MMU.Init() failed, expected %v, got %v", 0x0, mmu.Memory[BOOT])
	}

	mmu = NewMMU()
	mmu.Init("", "roms/Tetris.gb")
	if mmu.Memory[BOOT] != 0x1 {
		t.Errorf("MMU.Init() failed, expected %v, got %v", 0x1, mmu.Memory[BOOT])
	}
}
