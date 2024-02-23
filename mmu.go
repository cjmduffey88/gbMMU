package gbMMU

import (
	"github.com/cjmduffey88/bits"
	"github.com/cjmduffey88/gbMBC"
	"os"
)

const (
	JP1 = 0xFF00 // JoyPad
	SB  = 0xFF01 // Serial Transfer Data
	SC  = 0xFF02 // Serial Transfer Control
	DMA = 0xFF46 // DMA Transfer

	DIV  = 0xFF04 // Divider Register
	TIMA = 0xFF05 // Timer Counter
	TMA  = 0xFF06 // Timer Modulo
	TAC  = 0xFF07 // Timer Control

	IF = 0xFF0F // Interrupted Flag
	IE = 0xFFFF // Interrupted Enable

	LCDC = 0xFF40 // LCD Control
	STAT = 0xFF41 // LCD Status
	SCY  = 0xFF42 // Scroll Y
	SCX  = 0xFF43 // Scroll X
	LY   = 0xFF44 // LCD Y-Coordinate
	LYC  = 0xFF45 // LY Compare
	BGP  = 0xFF47 // BG Palette Data
	OBP0 = 0xFF48 // Object Palette 0 Data
	OBP1 = 0xFF49 // Object Palette 1 Data
	WY   = 0xFF4A // Window Y Position
	WX   = 0xFF4B // Window X Position

	NR10 = 0xFF10 // Sound Mode 1 Register Sweep
	NR11 = 0xFF11 // Sound Mode 1 Register Wave Pattern Duty
	NR12 = 0xFF12 // Sound Mode 1 Register Envelope
	NR13 = 0xFF13 // Sound Mode 1 Register Frequency Low
	NR14 = 0xFF14 // Sound Mode 1 Register Frequency High
	NR21 = 0xFF16 // Sound Mode 2 Register Wave Pattern Duty
	NR22 = 0xFF17 // Sound Mode 2 Register Envelope
	NR23 = 0xFF18 // Sound Mode 2 Register Frequency Low
	NR24 = 0xFF19 // Sound Mode 2 Register Frequency High
	NR30 = 0xFF1A // Sound Mode 3 Register Sound On/Off
	NR31 = 0xFF1B // Sound Mode 3 Register Sound Length
	NR32 = 0xFF1C // Sound Mode 3 Register Select Output Level
	NR33 = 0xFF1D // Sound Mode 3 Register Frequency Low
	NR34 = 0xFF1E // Sound Mode 3 Register Frequency High
	NR41 = 0xFF20 // Sound Mode 4 Register Sound Length
	NR42 = 0xFF21 // Sound Mode 4 Register Envelope
	NR43 = 0xFF22 // Sound Mode 4 Register Polynomial Counter
	NR44 = 0xFF23 // Sound Mode 4 Register Counter/Consecutive; Inital
	NR50 = 0xFF24 // Channel Control / ON-OFF / Volume
	NR51 = 0xFF25 // Selection of Sound Output Terminal
	NR52 = 0xFF26 // Sound ON/OFF
	WAVE = 0xFF30 // Wave Pattern RAM

	BOOT = 0xFF50 // Boot ROM Disable
)

const (
	IntVBlank = uint8(iota)
	IntLCDStat
	IntTimer
	IntSerial
	IntJoyPad
)

type MMU struct {
	MBC        *gbMBC.MBC
	Boot       []byte
	Memory     []byte
	DMAEnabled bool  // DMA Transfer Enabled
	Steps      uint8 // DMA Steps
}

func NewMMU() *MMU {
	var mmu = new(MMU)
	mmu.Boot = make([]byte, 256)
	mmu.Memory = make([]byte, 0x10000)
	return mmu
}

func (mmu *MMU) Init(pathBoot, pathROM string) {
	switch pathBoot {
	case "":
		mmu.Memory[BOOT] = 0x01
	default:
		boot, err := os.ReadFile(pathBoot)
		if err != nil {
			panic(err)
		}
		copy(mmu.Boot, boot)
	}
	mmu.MBC = gbMBC.NewMBC(pathROM)
}

func (mmu *MMU) Read(address uint16) byte {
	switch {
	// Boot ROM
	case address < 0x100 && mmu.Memory[BOOT] == 0x00:
		return mmu.Boot[address]
	// MBC ROM
	case address < 0x8000:
		return mmu.MBC.Read(address)
	// MBC RAM
	case address >= 0xA000 && address <= 0xBFFF:
		return mmu.MBC.Read(address)

	default:
		return mmu.Memory[address]
	}
}

func (mmu *MMU) Write(address uint16, value byte) {
	switch {
	// MBC ROM
	case address <= 0x8000:
		mmu.MBC.Write(address, value)
	// MBC RAM
	case address >= 0xA000 && address <= 0xBFFF:
		mmu.MBC.Write(address, value)
	// WRAM (With Echo)
	case address >= 0xC000 && address <= 0xDDFF:
		mmu.Memory[address] = value
		mmu.Memory[address+0x2000] = value
	// WRAM (Without Echo)
	case address > 0xDDFF && address <= 0xDFFF:
		mmu.Memory[address] = value
	// Echo WRAM
	case address >= 0xE000 && address <= 0xFDFF:
		mmu.Memory[address] = value
		mmu.Memory[address-0x2000] = value

	// DIV Register (Writing anything to DIV resets it to 0)
	case address == DIV:
		mmu.Memory[DIV] = 0

	default:
		mmu.Memory[address] = value
	}
}

func (mmu *MMU) InterruptEnabled(interrupt uint8) bool {
	return bits.CheckBit(mmu.Memory[IE], interrupt)
}

func (mmu *MMU) InterruptFlagged(interrupt uint8) bool {
	return bits.CheckBit(mmu.Memory[IF], interrupt)
}

func (mmu *MMU) InterruptFlagSet(interrupt uint8) {
	bits.SetBit(mmu.Memory[IF], interrupt)
}

func (mmu *MMU) InterruptFlagClear(interrupt uint8) {
	bits.ClearBit(mmu.Memory[IF], interrupt)
}

func (mmu *MMU) DMATransfer(startAddress uint16) {
	for i := 0; i < 0xA0; i++ {
		mmu.Memory[0xFE00+i] = mmu.Read(startAddress + uint16(i))
	}
	mmu.DMAEnabled = true
}

func (mmu *MMU) Step() {
	if mmu.DMAEnabled {
		if mmu.Steps == 160 {
			mmu.DMAEnabled = false
			mmu.Steps = 0
			return
		}
		mmu.Steps++
	}
}
